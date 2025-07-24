package obfuscator

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"
)

// AntiDebugPass injects a dynamic, time-based key generation mechanism
// that is used by the string encryption pass. This weaves the anti-debugging
// check directly into the program's data flow.
type AntiDebugPass struct{}

func (p *AntiDebugPass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	// This pass should only run on the main package to ensure the init logic is executed once.
	if file.Name.Name != "main" {
		return nil
	}

	// Check if we have already run this pass on a file in the main package.
	for _, decl := range file.Decls {
		if gd, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range gd.Specs {
				if vs, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range vs.Names {
						if name.Name == obf.WeavingKeyVarName {
							return nil // Weaving key variable already exists.
						}
					}
				}
			}
		}
	}

	astutil.AddImport(fset, file, "time")

	// 1. Declare the global variable for the weaving key.
	// var o_wkey_... int64
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
	// func init() {
	//   t := time.Now().UnixNano()
	//   // Some meaningless work
	//   for i := 0; i < 100; i++ { _ = i }
	//   o_wkey_... = time.Now().UnixNano() - t
	// }
	startTimeVar := ast.NewIdent("t")
	initFunc := &ast.FuncDecl{
		Name: ast.NewIdent("init"),
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				// t := time.Now().UnixNano()
				&ast.AssignStmt{
					Lhs: []ast.Expr{startTimeVar},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.CallExpr{Fun: &ast.SelectorExpr{X: ast.NewIdent("time"), Sel: ast.NewIdent("Now")}},
							Sel: ast.NewIdent("UnixNano"),
						},
					}},
				},
				// A small, meaningless loop to slightly alter execution time.
				// If a debugger is attached, the time difference will be huge.
				&ast.ForStmt{
					Init: &ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent("i")}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}},
					Cond: &ast.BinaryExpr{X: ast.NewIdent("i"), Op: token.LSS, Y: &ast.BasicLit{Kind: token.INT, Value: "100"}},
					Post: &ast.IncDecStmt{X: ast.NewIdent("i"), Tok: token.INC},
					Body: &ast.BlockStmt{List: []ast.Stmt{
						&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent("_")}, Tok: token.ASSIGN, Rhs: []ast.Expr{ast.NewIdent("i")}},
					}},
				},
				// o_wkey_... = time.Now().UnixNano() - t
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(obf.WeavingKeyVarName)},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{
						&ast.BinaryExpr{
							X: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   &ast.CallExpr{Fun: &ast.SelectorExpr{X: ast.NewIdent("time"), Sel: ast.NewIdent("Now")}},
									Sel: ast.NewIdent("UnixNano"),
								},
							},
							Op: token.SUB,
							Y:  startTimeVar,
						},
					},
				},
			},
		},
	}

	// Find the index of the last import declaration.
	lastImportIndex := -1
	for i, decl := range file.Decls {
		if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
			lastImportIndex = i
		}
	}

	// Insert the new declarations after the last import.
	if lastImportIndex != -1 {
		file.Decls = append(file.Decls[:lastImportIndex+1], append([]ast.Decl{keyVarDecl, initFunc}, file.Decls[lastImportIndex+1:]...)...)
	} else {
		// If there are no imports, add the declarations to the top of the file.
		file.Decls = append([]ast.Decl{keyVarDecl, initFunc}, file.Decls...)
	}

	return nil
}
