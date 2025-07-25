package obfuscator

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"
)

// AntiDebugPass injects dynamic, multi-faceted anti-debugging checks.
// It combines time-based detection, ptrace checks (on Linux), and the result
// from the anti-vm pass to generate a dynamic key. This key is then used by
// the string encryption pass, weaving the checks into the program's data flow.
type AntiDebugPass struct{}

// Apply injects the anti-debugging logic into the main package.
func (p *AntiDebugPass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	// This pass should only run on the main package.
	if file.Name.Name != "main" {
		return nil
	}

	// Check if we have already run this pass.
	if isVarDeclared(file, obf.WeavingKeyVarName) {
		return nil
	}

	astutil.AddImport(fset, file, "time")
	astutil.AddImport(fset, file, "syscall")

	// 1. Declare the global variable for the weaving key.
	keyVarDecl := &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{ast.NewIdent(obf.WeavingKeyVarName)},
				Type:  ast.NewIdent("int64"),
			},
		},
	}

	// 2. Create the init function that calculates the key.
	initFunc := createAntiDebugInitFunc(obf)

	// 3. Insert the new declarations after the last import.
	insertDeclsAfterImports(file, []ast.Decl{keyVarDecl, initFunc})

	return nil
}

// createAntiDebugInitFunc generates the AST for an init function that combines multiple checks.
func createAntiDebugInitFunc(obf *Obfuscator) *ast.FuncDecl {
	// --- Time-based check ---
	timeComponentExpr := &ast.BinaryExpr{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.CallExpr{Fun: &ast.SelectorExpr{X: ast.NewIdent("time"), Sel: ast.NewIdent("Now")}},
				Sel: ast.NewIdent("UnixNano"),
			},
		},
		Op: token.SUB,
		Y:  ast.NewIdent("t"),
	}

	// --- Ptrace check (Linux specific) ---
	ptraceCheck := &ast.IfStmt{
		Init: &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("pres"), ast.NewIdent("_"), ast.NewIdent("_")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{ // This needs to be a slice
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{X: ast.NewIdent("syscall"), Sel: ast.NewIdent("Syscall")},
					Args: []ast.Expr{
						&ast.SelectorExpr{X: ast.NewIdent("syscall"), Sel: ast.NewIdent("SYS_PTRACE")},
						&ast.CallExpr{Fun: ast.NewIdent("uintptr"), Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}}, // PTRACE_TRACEME
						&ast.CallExpr{Fun: ast.NewIdent("uintptr"), Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}},
						&ast.CallExpr{Fun: ast.NewIdent("uintptr"), Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}},
					},
				},
			},
		},
		Cond: &ast.BinaryExpr{X: ast.NewIdent("pres"), Op: token.NEQ, Y: &ast.BasicLit{Kind: token.INT, Value: "0"}},
		Body: &ast.BlockStmt{List: []ast.Stmt{
			&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent("ptraceComponent")}, Tok: token.ASSIGN, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1337"}}},
		}},
		Else: &ast.BlockStmt{List: []ast.Stmt{
			&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent("ptraceComponent")}, Tok: token.ASSIGN, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}},
		}},
	}

	// --- Final key calculation ---
	finalCalc := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(obf.WeavingKeyVarName)},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{ // This needs to be a slice
			&ast.BinaryExpr{
				X: &ast.BinaryExpr{
					X:  ast.NewIdent("timeComponent"),
					Op: token.ADD,
					Y:  ast.NewIdent("ptraceComponent"),
				},
				Op: token.ADD,
				Y: &ast.BinaryExpr{
					X: &ast.CallExpr{
						Fun:  ast.NewIdent("int64"),
						Args: []ast.Expr{ast.NewIdent(vmCheckVarName)},
					},
					Op:  token.MUL,
					Y:   &ast.BasicLit{Kind: token.INT, Value: "9999"},
				},
			},
		},
	}

	// --- Assemble the function body ---
	initBody := &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.DeclStmt{Decl: &ast.GenDecl{Tok: token.VAR, Specs: []ast.Spec{&ast.ValueSpec{Names: []*ast.Ident{ast.NewIdent("ptraceComponent")}, Type: ast.NewIdent("int64")}}}},
			&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent("t")}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.SelectorExpr{X: &ast.CallExpr{Fun: &ast.SelectorExpr{X: ast.NewIdent("time"), Sel: ast.NewIdent("Now")}}, Sel: ast.NewIdent("UnixNano")}}}},
			&ast.ForStmt{
				Init: &ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent("i")}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}},
				Cond: &ast.BinaryExpr{X: ast.NewIdent("i"), Op: token.LSS, Y: &ast.BasicLit{Kind: token.INT, Value: "100"}},
				Post: &ast.IncDecStmt{X: ast.NewIdent("i"), Tok: token.INC},
				Body: &ast.BlockStmt{List: []ast.Stmt{&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent("_")}, Tok: token.ASSIGN, Rhs: []ast.Expr{ast.NewIdent("i")}}}},
			},
			&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent("timeComponent")}, Tok: token.DEFINE, Rhs: []ast.Expr{timeComponentExpr}},
			ptraceCheck,
			finalCalc,
		},
	}

	return &ast.FuncDecl{
		Name: ast.NewIdent("init"),
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: initBody,
	}
}
