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

	astutil.AddImport(fset, file, "syscall")
	astutil.AddImport(fset, file, "runtime")

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
	// --- Ptrace check (Linux specific, wrapped for cross-platform compilation) ---
	ptraceCheck := &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.SelectorExpr{X: ast.NewIdent("runtime"), Sel: ast.NewIdent("GOOS")},
			Op: token.EQL,
			Y:  &ast.BasicLit{Kind: token.STRING, Value: `"linux"`},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{
			&ast.IfStmt{
				Init: &ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("pres"), ast.NewIdent("_"), ast.NewIdent("_")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
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
			},
		}},
	}

	// --- Final key calculation ---
	// The key is now based on stable checks (ptrace, vm)
	finalCalc := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(obf.WeavingKeyVarName)},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			&ast.BinaryExpr{
				X:  ast.NewIdent("ptraceComponent"),
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
			// Initialize ptraceComponent to 0
			&ast.DeclStmt{Decl: &ast.GenDecl{Tok: token.VAR, Specs: []ast.Spec{
				&ast.ValueSpec{
					Names:  []*ast.Ident{ast.NewIdent("ptraceComponent")},
					Type:   ast.NewIdent("int64"),
					Values: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
				},
			}}},
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
