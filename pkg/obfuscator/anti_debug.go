package obfuscator

import (
	"fmt"
	"go/ast"
	"go/token"
	"math/rand"

	"golang.org/x/tools/go/ast/astutil"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// randomIdentifier generates a random string to be used as a Go identifier.
func randomIdentifier(prefix string) string {
	b := make([]byte, 12)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return prefix + "_" + string(b)
}

// AntiDebugPass injects polymorphic time-based and ptrace-based anti-debugging checks.
type AntiDebugPass struct{}

func (p *AntiDebugPass) Apply(fset *token.FileSet, file *ast.File) error {
	// Inject polymorphic time-based checks into functions
	astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		funcDecl, ok := cursor.Node().(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil || len(funcDecl.Body.List) == 0 {
			return true
		}

		// Don't inject into very small functions or every single function
		if len(funcDecl.Body.List) < 2 || rand.Intn(100) < 70 {
			return true
		}

		astutil.AddImport(fset, file, "time")
		astutil.AddImport(fset, file, "os")

		startTimeVar := ast.NewIdent(randomIdentifier("o_debug_start"))

		startStmt := &ast.AssignStmt{
			Lhs: []ast.Expr{startTimeVar},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.CallExpr{
				Fun: &ast.SelectorExpr{X: ast.NewIdent("time"), Sel: ast.NewIdent("Now")},
			}},
		}

		// Polymorphic check expression
		var checkExpr ast.Expr
		threshold := &ast.BinaryExpr{
			X:  &ast.SelectorExpr{X: ast.NewIdent("time"), Sel: ast.NewIdent("Second")},
			Op: token.MUL,
			Y:  &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", rand.Intn(3)+2)}, // 2-4 seconds
		}
		sinceExpr := &ast.CallExpr{
			Fun:  &ast.SelectorExpr{X: ast.NewIdent("time"), Sel: ast.NewIdent("Since")},
			Args: []ast.Expr{startTimeVar},
		}

		if rand.Intn(2) == 0 {
			// time.Since(start) > threshold
			checkExpr = &ast.BinaryExpr{X: sinceExpr, Op: token.GTR, Y: threshold}
		} else {
			// threshold < time.Since(start)
			checkExpr = &ast.BinaryExpr{X: threshold, Op: token.LSS, Y: sinceExpr}
		}

		exitCall := &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun:  &ast.SelectorExpr{X: ast.NewIdent("os"), Sel: ast.NewIdent("Exit")},
				Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}},
			},
		}

		checkStmt := &ast.IfStmt{Cond: checkExpr, Body: &ast.BlockStmt{List: []ast.Stmt{exitCall}}}

		// Smart insertion
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
		injectPolymorphicPtraceCheck(fset, file)
	}

	return nil
}

// injectPolymorphicPtraceCheck adds a ptrace anti-debugging function with a random name
// and calls it from an init function.
func injectPolymorphicPtraceCheck(fset *token.FileSet, file *ast.File) {
	funcName := randomIdentifier("o_anti_debug_ptrace")

	// Check if we already injected a ptrace check (by looking for the build tag)
	for _, cgroup := range file.Comments {
		for _, c := range cgroup.List {
			if c.Text == "//go:build linux" {
				return // Assume already injected
			}
		}
	}

	astutil.AddImport(fset, file, "syscall")
	astutil.AddImport(fset, file, "os")

	// Create the polymorphic ptrace check function
	ptraceFunc := &ast.FuncDecl{
		Name: ast.NewIdent(funcName),
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
									Fun:  &ast.SelectorExpr{X: ast.NewIdent("os"), Sel: ast.NewIdent("Exit")},
									Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}},
								},
							},
						},
					},
				},
			},
		},
	}

	// Find or create an init function
	var initFunc *ast.FuncDecl
	for _, decl := range file.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok && f.Name.Name == "init" {
			initFunc = f
			break
		}
	}
	if initFunc == nil {
		initFunc = &ast.FuncDecl{
			Name: ast.NewIdent("init"),
			Type: &ast.FuncType{Params: &ast.FieldList{}},
			Body: &ast.BlockStmt{},
		}
		file.Decls = append(file.Decls, initFunc)
	}

	// Add the call to the ptrace check at the beginning of the init function
	initFunc.Body.List = append([]ast.Stmt{
		&ast.ExprStmt{X: &ast.CallExpr{Fun: ast.NewIdent(funcName)}},
	}, initFunc.Body.List...)

	// Add the ptrace function itself
	file.Decls = append(file.Decls, ptraceFunc)

	// Add a build constraint to the file to ensure it only compiles on Linux
	file.Comments = append(file.Comments, &ast.CommentGroup{
		List: []*ast.Comment{
			{Text: "//go:build linux"},
		},
	})
}