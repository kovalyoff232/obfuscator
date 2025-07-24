package obfuscator

import (
	"go/ast"
	"go/token"
	"math/rand"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

func ObfuscateControlFlow(node *ast.File) {
	astutil.Apply(node, func(cursor *astutil.Cursor) bool {
		block, ok := cursor.Node().(*ast.BlockStmt)
		if !ok || len(block.List) < 1 {
			return true
		}

		if parent := cursor.Parent(); parent != nil {
			if _, ok := parent.(*ast.CaseClause); ok {
				return true
			}
			if sw, ok := parent.(*ast.SwitchStmt); ok && sw.Body == block {
				return true
			}
		}

		// Check if the block contains a return statement. If so, skipping it to avoid "missing return" errors.
		for _, stmt := range block.List {
			if _, ok := stmt.(*ast.ReturnStmt); ok {
				return true
			}
		}

		if rand.Intn(100) < 30 { // 30% chance
			newStmts := createOpaqueSwitch(block.List)
			block.List = newStmts
		}

		return true
	}, nil)
}

func createOpaqueSwitch(stmts []ast.Stmt) []ast.Stmt {
	ctrlVarName := "o_ctrl_" + strconv.Itoa(rand.Intn(1000))

	// This variable will always be 0, making the switch predictable but opaque.
	initCtrlVar := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(ctrlVarName)},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
	}

	// The default case holds the original, real code.
	defaultCase := &ast.CaseClause{
		List: nil, // Represents the 'default' case
		Body: stmts,
	}

	// Create several junk cases that will never be executed.
	junkCases := []ast.Stmt{}
	numJunkCases := rand.Intn(3) + 1 // 1 to 3 junk cases
	for i := 0; i < numJunkCases; i++ {
		deadVarName := "o_dead_" + strconv.Itoa(rand.Intn(1000))
		junkCases = append(junkCases, &ast.CaseClause{
			List: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(i + 1)}}, // Cases 1, 2, 3...
			Body: []ast.Stmt{
				// Add some dead code inside the junk case to make it look real.
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(deadVarName)},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(rand.Intn(100))}},
				},
				// "Use" the variable to avoid "declared and not used" error.
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("_")},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{ast.NewIdent(deadVarName)},
				},
			},
		})
	}
	
	// Add the default case to the list of cases.
	allCases := append(junkCases, defaultCase)


	switchStmt := &ast.SwitchStmt{
		Tag: ast.NewIdent(ctrlVarName),
		Body: &ast.BlockStmt{
			List: allCases,
		},
	}

	return []ast.Stmt{initCtrlVar, switchStmt}
}

