package obfuscator

import (
	"go/ast"
	"go/token"
	"math/rand"

	"golang.org/x/tools/go/ast/astutil"
)

// AntiDebugPass injects time-based and ptrace-based anti-debugging checks.
type AntiDebugPass struct{}

func (p *AntiDebugPass) Apply(fset *token.FileSet, file *ast.File) error {
	// Inject time-based checks into functions
	astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		funcDecl, ok := cursor.Node().(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil || len(funcDecl.Body.List) == 0 {
			return true
		}

		// Don't inject into very small functions or every single function
		if len(funcDecl.Body.List) < 2 || rand.Intn(100) < 70 {
			return true
		}

		// Ensure necessary packages are imported
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
				Y:  &ast.BasicLit{Kind: token.INT, Value: "3"}, // Increased threshold
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

		var newBodyList []ast.Stmt
		newBodyList = append(newBodyList, startStmt)
		returnFound := false

		for _, stmt := range funcDecl.Body.List {
			if _, ok := stmt.(*ast.ReturnStmt); ok {
				newBodyList = append(newBodyList, checkStmt)
				returnFound = true
			}
			newBodyList = append(newBodyList, stmt)
		}

		if !returnFound {
			newBodyList = append(newBodyList, checkStmt)
		}

		funcDecl.Body.List = newBodyList

		return true
	}, nil)

	// Inject ptrace check, but only once per file, and only in main package.
	if file.Name.Name == "main" {
		injectPtraceCheck(fset, file)
	}

	return nil
}

// injectPtraceCheck adds a ptrace anti-debugging function and an init function to call it.
// This check is only for Linux.
func injectPtraceCheck(fset *token.FileSet, file *ast.File) {
	// Check if we already injected this
	for _, decl := range file.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok && f.Name.Name == "o_anti_debug_tracer" {
			return // Already injected
		}
	}

	astutil.AddImport(fset, file, "syscall")

	// Create the ptrace check function:
	// func o_anti_debug_tracer() {
	//     if _, _, err := syscall.Syscall(syscall.SYS_PTRACE, syscall.PTRACE_TRACEME, 0, 0); err != 0 {
	//         syscall.Exit(1)
	//     }
	// }
	ptraceFunc := &ast.FuncDecl{
		Name: ast.NewIdent("o_anti_debug_tracer"),
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.IfStmt{
					Init: &ast.AssignStmt{
						Lhs: []ast.Expr{ast.NewIdent("_"), ast.NewIdent("_"), ast.NewIdent("err")},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{X: ast.NewIdent("syscall"), Sel: ast.NewIdent("Syscall")},
								Args: []ast.Expr{
									&ast.SelectorExpr{X: ast.NewIdent("syscall"), Sel: ast.NewIdent("SYS_PTRACE")},
									&ast.SelectorExpr{X: ast.NewIdent("syscall"), Sel: ast.NewIdent("PTRACE_TRACEME")},
									&ast.BasicLit{Kind: token.INT, Value: "0"},
									&ast.BasicLit{Kind: token.INT, Value: "0"},
								},
							},
						},
					},
					Cond: &ast.BinaryExpr{
						X:  ast.NewIdent("err"),
						Op: token.NEQ,
						Y:  &ast.BasicLit{Kind: token.INT, Value: "0"},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ExprStmt{
								X: &ast.CallExpr{
									Fun:  &ast.SelectorExpr{X: ast.NewIdent("syscall"), Sel: ast.NewIdent("Exit")},
									Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}},
								},
							},
						},
					},
				},
			},
		},
	}

	// Create the init function to call the ptrace check:
	// func init() {
	//     o_anti_debug_tracer()
	// }
	initFunc := &ast.FuncDecl{
		Name: ast.NewIdent("init"),
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ExprStmt{X: &ast.CallExpr{Fun: ast.NewIdent("o_anti_debug_tracer")}},
			},
		},
	}

	// Add a build constraint to the file
	file.Comments = append(file.Comments, &ast.CommentGroup{
		List: []*ast.Comment{
			{Text: "//go:build linux"},
		},
	})

	// Add the new functions to the file's declarations
	file.Decls = append(file.Decls, ptraceFunc, initFunc)
}
