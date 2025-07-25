package obfuscator

import (
	"go/ast"
	"go/token"
	"math/rand"
)

// MetamorphicEngine provides functions to generate varied but functionally equivalent code.
type MetamorphicEngine struct{}

// GenerateJunkCodeBlock creates a block of random, non-functional "junk" code.
func (e *MetamorphicEngine) GenerateJunkCodeBlock() []ast.Stmt {
	// Randomly choose a junk code template
	template := rand.Intn(2)
	switch template {
	case 0:
		return e.generateMathJunk()
	case 1:
		return []ast.Stmt{e.generateOpaquePredicate()}
	default:
		return []ast.Stmt{}
	}
}

// generateMathJunk creates a block of pointless arithmetic operations.
func (e *MetamorphicEngine) generateMathJunk() []ast.Stmt {
	x, y, z := NewName(), NewName(), NewName()
	return []ast.Stmt{
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(x)}, Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1337"}},
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(y)}, Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.BinaryExpr{X: ast.NewIdent(x), Op: token.MUL, Y: ast.NewIdent(x)}},
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(z)}, Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.BinaryExpr{X: ast.NewIdent(y), Op: token.XOR, Y: ast.NewIdent(x)}},
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("_")}, Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent(z)},
		},
	}
}

// generateOpaquePredicate creates an if statement with a condition that is always true.
func (e *MetamorphicEngine) generateOpaquePredicate() ast.Stmt {
	x, y := NewName(), NewName()
	return &ast.IfStmt{
		Init: &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(x)}, Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "99"}},
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
					Rhs: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"junk"`}},
				},
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("_")}, Tok: token.ASSIGN,
					Rhs: []ast.Expr{ast.NewIdent(y)},
				},
			},
		},
	}
}
