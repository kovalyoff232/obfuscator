package obfuscator

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"
)

// InsertDeadCode traverses the AST and injects various patterns of junk code
// into function bodies to hinder manual analysis.
func InsertDeadCode(file *ast.File) {
	astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		// Check if we are inside a function declaration
		funcDecl, isFunc := cursor.Parent().(*ast.FuncDecl)
		if !isFunc || funcDecl.Name == nil {
			return true
		}

		// Do not insert dead code into init functions
		if funcDecl.Name.Name == "init" {
			return true
		}

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
		if randInt(3) == 0 { // 33% chance to insert code into any given block
			var junkStmts []ast.Stmt
			template := randInt(3) // Choose one of the templates

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
				insertIndex := randInt(int64(len(block.List)))
				block.List = append(block.List[:insertIndex], append(junkStmts, block.List[insertIndex:]...)...)
			}
		}

		return true
	}, nil)
}

// createMathJunk creates a block of pointless arithmetic operations.
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
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("_")}, Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent(z)},
		},
	}
}

// createOpaquePredicateJunk creates an if statement with a condition that is always true
// but harder for a static analyzer to prove.
func createOpaquePredicateJunk() []ast.Stmt {
	x, y := NewName(), NewName()
	var cond ast.Expr

	template := randInt(3)
	switch template {
	case 0:
		// (x*x - 1) == (x-1)*(x+1)
		cond = &ast.BinaryExpr{
			X: &ast.BinaryExpr{
				X:  &ast.BinaryExpr{X: ast.NewIdent(x), Op: token.MUL, Y: ast.NewIdent(x)},
				Op: token.SUB,
				Y:  &ast.BasicLit{Kind: token.INT, Value: "1"},
			},
			Op: token.EQL,
			Y: &ast.BinaryExpr{
				X:  &ast.ParenExpr{X: &ast.BinaryExpr{X: ast.NewIdent(x), Op: token.SUB, Y: &ast.BasicLit{Kind: token.INT, Value: "1"}}},
				Op: token.MUL,
				Y:  &ast.ParenExpr{X: &ast.BinaryExpr{X: ast.NewIdent(x), Op: token.ADD, Y: &ast.BasicLit{Kind: token.INT, Value: "1"}}},
			},
		}
	case 1:
		// 7*x - 3*x - 4*x == 0
		cond = &ast.BinaryExpr{
			X: &ast.BinaryExpr{
				X: &ast.BinaryExpr{
					X:  &ast.BinaryExpr{X: &ast.BasicLit{Kind: token.INT, Value: "7"}, Op: token.MUL, Y: ast.NewIdent(x)},
					Op: token.SUB,
					Y:  &ast.BinaryExpr{X: &ast.BasicLit{Kind: token.INT, Value: "3"}, Op: token.MUL, Y: ast.NewIdent(x)},
				},
				Op: token.SUB,
				Y:  &ast.BinaryExpr{X: &ast.BasicLit{Kind: token.INT, Value: "4"}, Op: token.MUL, Y: ast.NewIdent(x)},
			},
			Op: token.EQL,
			Y:  &ast.BasicLit{Kind: token.INT, Value: "0"},
		}
	default:
		// y := x*x; (y - x*x) == 0
		cond = &ast.BinaryExpr{
			X: &ast.BinaryExpr{
				X:  ast.NewIdent(y),
				Op: token.SUB,
				Y:  &ast.BinaryExpr{X: ast.NewIdent(x), Op: token.MUL, Y: ast.NewIdent(x)},
			},
			Op: token.EQL,
			Y:  &ast.BasicLit{Kind: token.INT, Value: "0"},
		}
	}

	return []ast.Stmt{&ast.IfStmt{
		Init: &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(x)}, Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "123"}},
		},
		Cond: cond,
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
