package obfuscator

import (
	"go/ast"
	"go/token"
	"math/rand"

	"golang.org/x/tools/go/ast/astutil"
)

// InsertDeadCode traverses the AST and injects various patterns of junk code
// into function bodies to hinder manual analysis.
func InsertDeadCode(file *ast.File) {
	astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		block, ok := cursor.Node().(*ast.BlockStmt)
		if !ok || len(block.List) == 0 {
			return true
		}

		// Check if the parent is a switch statement. We can't add statements directly
		// to a switch, only case clauses.
		if parent := cursor.Parent(); parent != nil {
			if _, ok := parent.(*ast.SwitchStmt); ok {
				return true
			}
		}

		// Don't insert dead code into tiny blocks.
		if len(block.List) < 2 {
			return true
		}

		// Insert dead code at a random position within the block.
		if rand.Intn(3) == 0 { // 33% chance to insert code into any given block
			var junkStmts []ast.Stmt
			template := rand.Intn(3) // Choose one of the templates

			switch template {
			case 0:
				junkStmts = createMathJunk()
			case 1:
				junkStmts = createOpaquePredicateJunk()
			case 2:
				junkStmts = createAllocationJunk()
			}

			// Insert the junk statement at a random index.
			if len(block.List) > 0 {
				// Ensure we don't insert into an empty block, though we checked for len > 1
				insertIndex := rand.Intn(len(block.List))
				block.List = append(block.List[:insertIndex], append(junkStmts, block.List[insertIndex:]...)...)
			}
		}

		return true
	}, nil)
}

// createMathJunk creates a block of pointless arithmetic operations.
// E.g., x := 123; y := x * x; z := y - x; _ = z
func createMathJunk() []ast.Stmt {
	x, y, z := NewName(), NewName(), NewName()
	return []ast.Stmt{
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(x)}, Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "123"}},
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(y)}, Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.BinaryExpr{X: ast.NewIdent(x), Op: token.MUL, Y: ast.NewIdent(x)}},
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(z)}, Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.BinaryExpr{X: ast.NewIdent(y), Op: token.SUB, Y: ast.NewIdent(x)}},
		},
		// Make sure the variable is "used" to satisfy the compiler.
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("_")}, Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent(z)},
		},
	}
}

// createOpaquePredicateJunk creates an if statement with a condition that is always true
// but hard for a static analyzer to prove.
// E.g., x := 123; if (x*x + 1) > 0 { _ = "junk" }
func createOpaquePredicateJunk() []ast.Stmt {
	x, y := NewName(), NewName()
	return []ast.Stmt{&ast.IfStmt{
		Init: &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(x)}, Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "123"}},
		},
		Cond: &ast.BinaryExpr{
			X: &ast.BinaryExpr{
				X:  &ast.BinaryExpr{X: ast.NewIdent(x), Op: token.MUL, Y: ast.NewIdent(x)},
				Op: token.ADD,
				Y:  &ast.BasicLit{Kind: token.INT, Value: "1"},
			},
			Op: token.GTR,
			Y:  &ast.BasicLit{Kind: token.INT, Value: "0"},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(y)}, Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: "\"junk\""}},
				},
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("_")}, Tok: token.ASSIGN,
					Rhs: []ast.Expr{ast.NewIdent(y)},
				},
			},
		},
	}}
}

// createAllocationJunk creates junk code that allocates memory and then "uses" it.
func createAllocationJunk() []ast.Stmt {
	sliceVar := NewName()
	return []ast.Stmt{
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(sliceVar)}, Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.CallExpr{
				Fun: ast.NewIdent("make"),
				Args: []ast.Expr{
					&ast.ArrayType{Elt: ast.NewIdent("byte")},
					&ast.BasicLit{Kind: token.INT, Value: "1024"},
				},
			}},
		},
		// "Use" the variable to avoid compiler errors.
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("_")}, Tok: token.ASSIGN,
			Rhs: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{ast.NewIdent(sliceVar)}}},
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(sliceVar)}, Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent("nil")},
		},
	}
}
