package obfuscator

import (
	"fmt"
	"go/ast"
	"go/token"
	"math/rand"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

func InsertDeadCode(node *ast.File) {
	astutil.Apply(node, func(cursor *astutil.Cursor) bool {
		block, ok := cursor.Node().(*ast.BlockStmt)
		if !ok || len(block.List) == 0 {
			return true
		}

		// --- SAFETY CHECKS ---

		// 1. Check parent node. Do not insert code into blocks that are
		// direct children of switch statements, as variable declarations
		// are not allowed there (only inside a case).
		parent := cursor.Parent()
		if _, isSwitch := parent.(*ast.SwitchStmt); isSwitch {
			return true // Skip this block
		}
		if _, isSwitch := parent.(*ast.TypeSwitchStmt); isSwitch {
			return true // Skip this block
		}

		// 2. Avoid inserting dead code in blocks that end with a return statement,
		// as it can sometimes lead to "missing return" compiler errors.
		hasReturn := false
		for _, stmt := range block.List {
			if _, ok := stmt.(*ast.ReturnStmt); ok {
				hasReturn = true
				break
			}
		}
		if hasReturn {
			return true
		}

		// --- INSERTION ---

		// Insert with a certain probability
		if rand.Intn(100) < 30 { // 30% chance
			newStmt, newDecls := createOpaqueDeadIfStmt()
			// Prepend declarations to ensure variables are in scope
			block.List = append(newDecls, block.List...)
			// Append the actual dead code block
			block.List = append(block.List, newStmt)
		}

		return true
	}, nil)
}

// createOpaqueDeadIfStmt creates an if statement with a condition that is
// always false but is difficult for a static analyzer to prove.
// It returns the if statement and any variable declarations needed for it.
func createOpaqueDeadIfStmt() (ast.Stmt, []ast.Stmt) {
	varName1 := fmt.Sprintf("o_dead_%d", rand.Intn(1000))
	varName2 := fmt.Sprintf("o_dead_%d", rand.Intn(1000)+1000)
	varName3 := fmt.Sprintf("o_dead_%d", rand.Intn(1000)+2000)

	// Declarations for the variables used in the opaque predicate.
	decls := []ast.Stmt{
		&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names:  []*ast.Ident{ast.NewIdent(varName1)},
						Values: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(rand.Intn(100) + 2)}},
					},
					&ast.ValueSpec{
						Names:  []*ast.Ident{ast.NewIdent(varName2)},
						Values: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(rand.Intn(100) + 2)}},
					},
				},
			},
		},
	}

	// The opaque predicate: (varName1 * varName2) + 1 < (varName1 * varName2)
	opaqueCondition := &ast.BinaryExpr{
		X: &ast.ParenExpr{
			X: &ast.BinaryExpr{
				X:  &ast.BinaryExpr{X: ast.NewIdent(varName1), Op: token.MUL, Y: ast.NewIdent(varName2)},
				Op: token.ADD,
				Y:  &ast.BasicLit{Kind: token.INT, Value: "1"},
			},
		},
		Op: token.LSS, // Less than
		Y:  &ast.BinaryExpr{X: ast.NewIdent(varName1), Op: token.MUL, Y: ast.NewIdent(varName2)},
	}

	// The body of the if statement contains junk code that will never be executed.
	deadIfStmt := &ast.IfStmt{
		Cond: opaqueCondition,
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(varName3)},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: ast.NewIdent("len"),
							Args: []ast.Expr{
								&ast.BasicLit{Kind: token.STRING, Value: `"dead_code"`},
							},
						},
					},
				},
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("_")},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{ast.NewIdent(varName3)},
				},
			},
		},
	}

	return deadIfStmt, decls
}
