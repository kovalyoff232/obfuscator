package obfuscator

import (
	"fmt"
	"go/ast"
	"go/token"
	"math/rand"

	"golang.org/x/tools/go/ast/astutil"
)

const dispatcherFuncName = "o_dispatch"

type funcInfo struct {
	decl    *ast.FuncDecl
	id      string
	file    *ast.File
}

type callIndirectionPass struct {
	funcs map[string]*funcInfo
	mainFile *ast.File
}

func ApplyCallIndirection(files map[string]*ast.File) error {
	pass := &callIndirectionPass{
		funcs: make(map[string]*funcInfo),
	}

	if err := pass.collectFuncs(files); err != nil {
		return fmt.Errorf("error collecting funcs: %w", err)
	}

	if len(pass.funcs) == 0 {
		fmt.Println("   - Call indirection: no functions found to replace.")
		return nil
	}

	if err := pass.rewriteCalls(files); err != nil {
		return fmt.Errorf("error rewriting calls: %w", err)
	}

	if err := pass.injectDispatcher(); err != nil {
		return fmt.Errorf("error injecting dispatcher: %w", err)
	}

	return nil
}

func (p *callIndirectionPass) collectFuncs(files map[string]*ast.File) error {
	for _, file := range files {
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				if fn.Name.Name == "main" || fn.Name.Name == "init" || fn.Name.Name == dispatcherFuncName {
					continue
				}
				if fn.Name == nil {
					continue
				}

				funcName := fn.Name.Name
				p.funcs[funcName] = &funcInfo{
					decl:    fn,
					id:      fmt.Sprintf("id_%d", rand.Intn(100000)),
					file:    file,
				}
				fmt.Printf("    - Found function for call indirection: %s\n", funcName)
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

func (p *callIndirectionPass) rewriteCalls(files map[string]*ast.File) error {
	for _, file := range files {
		astutil.Apply(file, func(cursor *astutil.Cursor) bool {
			call, ok := cursor.Node().(*ast.CallExpr)
			if !ok {
				return true
			}

			ident, ok := call.Fun.(*ast.Ident)
			if !ok {
				return true
			}

			if info, exists := p.funcs[ident.Name]; exists {
				fmt.Printf("    - Rewriting call to %s() in file %s\n", ident.Name, file.Name.Name)

				newCall := &ast.CallExpr{
					Fun: ast.NewIdent(dispatcherFuncName),
					Args: []ast.Expr{
						&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, info.id)},
					},
				}
				newCall.Args = append(newCall.Args, call.Args...)

				if info.decl.Type.Results != nil && len(info.decl.Type.Results.List) > 0 {
					returnType := info.decl.Type.Results.List[0].Type
					assertExpr := &ast.TypeAssertExpr{
						X:    newCall,
						Type: returnType,
					}
					cursor.Replace(assertExpr)
				} else {
					cursor.Replace(newCall)
				}
			}
			return true
		}, nil)
	}
	return nil
}

func (p *callIndirectionPass) injectDispatcher() error {
	if p.mainFile == nil {
		return fmt.Errorf("main file not found for dispatcher injection")
	}

	var cases []ast.Stmt
	for name, info := range p.funcs {
		caseClause := &ast.CaseClause{
			List: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, info.id)}},
		}

		originalCall := &ast.CallExpr{
			Fun: ast.NewIdent(name),
		}

		for i, field := range info.decl.Type.Params.List {
			for range field.Names {
				arg := &ast.TypeAssertExpr{
					X: &ast.IndexExpr{
						X:     ast.NewIdent("args"),
										Index: &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", i)},
					},
					Type: field.Type,
				}
				originalCall.Args = append(originalCall.Args, arg)
			}
		}

		if info.decl.Type.Results != nil && len(info.decl.Type.Results.List) > 0 {
			caseClause.Body = []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{originalCall}}}
		} else {
			caseClause.Body = []ast.Stmt{&ast.ExprStmt{X: originalCall}, &ast.ReturnStmt{}}
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

	dispatcherFunc := &ast.FuncDecl{
		Name: ast.NewIdent(dispatcherFuncName),
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: []*ast.Field{
				{Names: []*ast.Ident{ast.NewIdent("id")}, Type: ast.NewIdent("string")},
				{Names: []*ast.Ident{ast.NewIdent("args")}, Type: &ast.Ellipsis{Elt: ast.NewIdent("interface{}")}},
			}},
			Results: &ast.FieldList{List: []*ast.Field{
				{Type: ast.NewIdent("interface{}")},
			}},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.SwitchStmt{
					Tag:  ast.NewIdent("id"),
					Body: &ast.BlockStmt{List: cases},
				},
				&ast.ReturnStmt{Results: []ast.Expr{ast.NewIdent("nil")}},
			},
		},
	}

	p.mainFile.Decls = append(p.mainFile.Decls[:len(p.mainFile.Imports)], append([]ast.Decl{dispatcherFunc}, p.mainFile.Decls[len(p.mainFile.Imports):]...)...)
	fmt.Printf("    - Injected dispatcher '%s' into file %s\n", dispatcherFuncName, p.mainFile.Name.Name)

	return nil
}
