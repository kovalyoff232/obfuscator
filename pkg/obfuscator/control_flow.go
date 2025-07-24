package obfuscator

import (
	"go/ast"
	"go/token"
	"math/rand"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

// ControlFlow flattens the control flow of function bodies.
// It breaks down a function's body into basic blocks and places them
// inside a single loop with a switch statement that dictates the flow of execution.
func ControlFlow(f *ast.File) {
	astutil.Apply(f, func(cursor *astutil.Cursor) bool {
		// We are looking for function declarations
		funcDecl, ok := cursor.Node().(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil || len(funcDecl.Body.List) == 0 {
			return true
		}

		// Skip functions that are too small to be worth flattening
		if len(funcDecl.Body.List) < 2 {
			return true
		}

		// Skip functions with return statements that might cause "missing return" errors
		// A more sophisticated analysis would be needed to handle these correctly.
		hasReturn := false
		ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
			if _, ok := n.(*ast.ReturnStmt); ok {
				hasReturn = true
			}
			return !hasReturn
		})
		if hasReturn {
			return true
		}

		flattenedBody := flattenFuncBody(funcDecl.Body)
		funcDecl.Body = flattenedBody

		// We handled this function, no need to inspect its children further
		return false
	}, nil)
}

// flattenFuncBody takes a block statement (a function body) and applies control flow flattening.
func flattenFuncBody(body *ast.BlockStmt) *ast.BlockStmt {
	// The control variable for the switch statement
	ctrlVarName := "o_ctrl_flow_" + strconv.Itoa(rand.Intn(10000))
	ctrlVar := ast.NewIdent(ctrlVarName)

	// 1. Initialize the control variable: var o_ctrl_flow_XXX = 0
	initStmt := &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names:  []*ast.Ident{ctrlVar},
					Values: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
				},
			},
		},
	}

	// 2. Create the list of cases for the switch. Each statement becomes a case.
	var cases []ast.Stmt
	numStmts := len(body.List)
	for i, stmt := range body.List {
		// The body of the case contains the original statement
		caseBody := []ast.Stmt{stmt}

		// After the statement, update the control variable to point to the next block
		nextState := i + 1
		if nextState >= numStmts {
			// If this is the last statement, set state to -1 to break the loop
			nextState = -1
		}

		updateCtrlVar := &ast.AssignStmt{
			Lhs: []ast.Expr{ctrlVar},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(nextState)}},
		}
		caseBody = append(caseBody, updateCtrlVar)

		// Create the case clause: case i: ...
		clause := &ast.CaseClause{
			List: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(i)}},
			Body: caseBody,
		}
		cases = append(cases, clause)
	}

	// 3. Create the main switch statement
	switchStmt := &ast.SwitchStmt{
		Tag:  ctrlVar,
		Body: &ast.BlockStmt{List: cases},
	}

	// 4. Create the infinite for loop that contains the switch
	loopStmt := &ast.ForStmt{
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				// Add a condition to break the loop
				&ast.IfStmt{
					Cond: &ast.BinaryExpr{
						X:  ctrlVar,
						Op: token.LSS, // Less than 0
						Y:  &ast.BasicLit{Kind: token.INT, Value: "0"},
					},
					Body: &ast.BlockStmt{List: []ast.Stmt{&ast.BranchStmt{Tok: token.BREAK}}},
				},
				switchStmt,
			},
		},
	}

	// 5. Assemble the new function body
	newBody := &ast.BlockStmt{
		List: []ast.Stmt{
			initStmt,
			loopStmt,
		},
	}

	return newBody
}