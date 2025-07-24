package obfuscator

import (
	"go/ast"
	"go/token"
	"math/rand"

	"golang.org/x/tools/go/ast/astutil"
)

// AntiDebugPass injects time-based anti-debugging checks into function bodies.
type AntiDebugPass struct{}

func (p *AntiDebugPass) Apply(fset *token.FileSet, file *ast.File) error {
	astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		funcDecl, ok := cursor.Node().(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil || len(funcDecl.Body.List) == 0 {
			return true
		}

		if len(funcDecl.Body.List) < 2 || rand.Intn(100) < 70 {
			return true
		}

		// Ensure necessary packages are imported before modifying the tree.
		astutil.AddImport(fset, file, "time")
		astutil.AddImport(fset, file, "os")

		startTimeVar := ast.NewIdent("o_debug_start_" + funcDecl.Name.Name)

		startStmt := &ast.AssignStmt{
			Lhs: []ast.Expr{startTimeVar},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.CallExpr{
				Fun: &ast.SelectorExpr{X: ast.NewIdent("time"), Sel: ast.NewIdent("Now")},
			}},
		}

		checkExpr := &ast.BinaryExpr{
			X: &ast.CallExpr{
				Fun:  &ast.SelectorExpr{X: ast.NewIdent("time"), Sel: ast.NewIdent("Since")},
				Args: []ast.Expr{startTimeVar},
			},
			Op: token.GTR,
			Y: &ast.BinaryExpr{
				X:  &ast.SelectorExpr{X: ast.NewIdent("time"), Sel: ast.NewIdent("Second")},
				Op: token.MUL,
				Y:  &ast.BasicLit{Kind: token.INT, Value: "2"},
			},
		}

		exitCall := &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun:  &ast.SelectorExpr{X: ast.NewIdent("os"), Sel: ast.NewIdent("Exit")},
				Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}},
			},
		}

		checkStmt := &ast.IfStmt{
			Cond: checkExpr,
			Body: &ast.BlockStmt{List: []ast.Stmt{exitCall}},
		}

		// --- Smart Insertion Logic ---
		var newBodyList []ast.Stmt
		newBodyList = append(newBodyList, startStmt)
		returnFound := false

		for _, stmt := range funcDecl.Body.List {
			if _, ok := stmt.(*ast.ReturnStmt); ok {
				// Insert the check *before* the return statement.
				newBodyList = append(newBodyList, checkStmt)
				returnFound = true
			}
			newBodyList = append(newBodyList, stmt)
		}

		// If no return statement was found, append the check at the end.
		if !returnFound {
			newBodyList = append(newBodyList, checkStmt)
		}

		funcDecl.Body.List = newBodyList

		return true
	}, nil)

	return nil
}