package obfuscator
import (
	"fmt"
	"go/ast"
	"go/token"
	"math/rand"
	"strconv"
	"golang.org/x/tools/go/ast/astutil"
)
type funcInfo struct {
	decl     *ast.FuncDecl
	id       int
	file     *ast.File
	isMethod bool
}
type CallIndirectionPass struct {
	funcs              map[string]*funcInfo
	mainFile           *ast.File
	dispatcherFuncName string
	maskingKey         int // A static key component to add noise.
	nextFuncID         int
}
func (p *CallIndirectionPass) Apply(obf *Obfuscator, fset *token.FileSet, files map[string]*ast.File) error {
	fmt.Println("  - Applying call indirection with dynamic keying...")
	p.funcs = make(map[string]*funcInfo)
	p.dispatcherFuncName = NewName()
	p.maskingKey = rand.Intn(1<<16) + 1 // A static, non-zero random integer.
	p.nextFuncID = 1
	if err := p.collectFuncs(files); err != nil {
		return fmt.Errorf("error collecting funcs: %w", err)
	}
	if len(p.funcs) == 0 {
		fmt.Println("   - Call indirection: no functions found to replace.")
		return nil
	}
	if err := p.rewriteCalls(files); err != nil {
		return fmt.Errorf("error rewriting calls: %w", err)
	}
	if err := p.injectDispatcher(obf); err != nil {
		return fmt.Errorf("error injecting dispatcher: %w", err)
	}
	return nil
}
func (p *CallIndirectionPass) collectFuncs(files map[string]*ast.File) error {
	for _, file := range files {
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				if fn.Name.Name == "main" || fn.Name.Name == "init" || fn.Name.Name == p.dispatcherFuncName {
					continue
				}
				if fn.Name == nil {
					continue
				}
				funcName := fn.Name.Name
				p.funcs[funcName] = &funcInfo{
					decl:     fn,
					id:       p.nextFuncID,
					file:     file,
					isMethod: fn.Recv != nil,
				}
				p.nextFuncID++
			}
		}
		if file.Name.Name == "main" {
			p.mainFile = file
		}
	}
	if p.mainFile == nil && len(p.funcs) > 0 {
		for _, file := range files {
			p.mainFile = file
			break
		}
	}
	return nil
}
func (p *CallIndirectionPass) rewriteCalls(files map[string]*ast.File) error {
	for path, file := range files {
		astutil.Apply(file, func(cursor *astutil.Cursor) bool {
			call, ok := cursor.Node().(*ast.CallExpr)
			if !ok {
				return true
			}
			var info *funcInfo
			var recv ast.Expr
			var funcName string
			switch fun := call.Fun.(type) {
			case *ast.Ident:
				funcName = fun.Name
				info = p.funcs[funcName]
			case *ast.SelectorExpr:
				funcName = fun.Sel.Name
				info = p.funcs[funcName]
				if info != nil && info.isMethod {
					recv = fun.X
				}
			}
			if info == nil {
				return true
			}
			fmt.Printf("    - Rewriting call to %s() in file %s\n", funcName, path)
			newCall := &ast.CallExpr{
				Fun: ast.NewIdent(p.dispatcherFuncName),
			}
			// Obfuscate the ID with the static component. The dynamic part happens in the dispatcher.
			obfuscatedID := info.id ^ p.maskingKey
			newCall.Args = append(newCall.Args, &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(obfuscatedID)})
			if info.isMethod {
				newCall.Args = append(newCall.Args, recv)
			}
			newCall.Args = append(newCall.Args, call.Args...)
			if info.decl.Type.Results != nil && len(info.decl.Type.Results.List) > 0 {
				returnType := info.decl.Type.Results.List[0].Type
				if types, ok := returnType.(*ast.Ident); ok && types.Name == "error" {
					cursor.Replace(newCall)
				} else {
					assertExpr := &ast.TypeAssertExpr{
						X:    newCall,
						Type: returnType,
					}
					cursor.Replace(assertExpr)
				}
			} else {
				cursor.Replace(newCall)
			}
			return false
		}, nil)
	}
	return nil
}
func (p *CallIndirectionPass) injectDispatcher(obf *Obfuscator) error {
	if p.mainFile == nil {
		return fmt.Errorf("main file not found for dispatcher injection")
	}
	var cases []ast.Stmt
	for name, info := range p.funcs {
		caseClause := &ast.CaseClause{
			List: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(info.id)}},
		}
		var originalCall ast.Expr
		argOffset := 0
		if info.isMethod {
			recvType := info.decl.Recv.List[0].Type
			recvAssert := &ast.TypeAssertExpr{
				X:    ast.NewIdent("args[0]"),
				Type: recvType,
			}
			originalCall = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   recvAssert,
					Sel: ast.NewIdent(name),
				},
			}
			argOffset = 1
		} else {
			originalCall = &ast.CallExpr{
				Fun: ast.NewIdent(name),
			}
		}
		callExpr := originalCall.(*ast.CallExpr)
		argIndex := argOffset
		for _, field := range info.decl.Type.Params.List {
			numNames := len(field.Names)
			if numNames == 0 {
				numNames = 1
			}
			for i := 0; i < numNames; i++ {
				arg := &ast.TypeAssertExpr{
					X: &ast.IndexExpr{
						X:     ast.NewIdent("args"),
						Index: &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", argIndex)},
					},
					Type: field.Type,
				}
				callExpr.Args = append(callExpr.Args, arg)
				argIndex++
			}
		}
		if info.decl.Type.Results != nil && len(info.decl.Type.Results.List) > 0 {
			caseClause.Body = []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{originalCall}}}
		} else {
			caseClause.Body = []ast.Stmt{&ast.ExprStmt{X: originalCall}, &ast.ReturnStmt{Results: []ast.Expr{ast.NewIdent("nil")}}}
		}
		cases = append(cases, caseClause)
	}
	cases = append(cases, &ast.CaseClause{
		List: nil, // default
		Body: []ast.Stmt{
			&ast.ExprStmt{X: &ast.CallExpr{
				Fun: ast.NewIdent("panic"),
				Args: []ast.Expr{
					&ast.BasicLit{Kind: token.STRING, Value: `"unknown function id in dispatcher"`},
				},
			}},
		},
	})
	// The local key is derived from the static masking key and the dynamic anti-debug key.
	localKeyCalculation := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("localKey")},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.BinaryExpr{
			X:  &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(p.maskingKey)},
			Op: token.XOR,
			Y:  &ast.CallExpr{Fun: ast.NewIdent("int"), Args: []ast.Expr{ast.NewIdent(obf.WeavingKeyVarName)}},
		}},
	}
	// The final ID is unmasked using this newly derived local key.
	idUnmasking := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("id")},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.BinaryExpr{
			X:  ast.NewIdent("obfuscatedID"),
			Op: token.XOR,
			Y:  ast.NewIdent("localKey"),
		}},
	}
	dispatcherFunc := &ast.FuncDecl{
		Name: ast.NewIdent(p.dispatcherFuncName),
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: []*ast.Field{
				{Names: []*ast.Ident{ast.NewIdent("obfuscatedID")}, Type: ast.NewIdent("int")},
				{Names: []*ast.Ident{ast.NewIdent("args")}, Type: &ast.Ellipsis{Elt: ast.NewIdent("interface{}")}},
			}},
			Results: &ast.FieldList{List: []*ast.Field{
				{Type: ast.NewIdent("interface{}")},
			}},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				localKeyCalculation,
				idUnmasking,
				&ast.SwitchStmt{
					Tag:  ast.NewIdent("id"),
					Body: &ast.BlockStmt{List: cases},
				},
				&ast.ReturnStmt{Results: []ast.Expr{ast.NewIdent("nil")}},
			},
		},
	}
	insertDeclsAfterImports(p.mainFile, []ast.Decl{dispatcherFunc})
	return nil
}
