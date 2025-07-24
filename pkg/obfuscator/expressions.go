package obfuscator

import (
	"fmt"
	"go/ast"
	"go/token"
	"math/rand"
)

func ObfuscateExpressions(node *ast.File) {
	ast.Inspect(node, func(n ast.Node) bool {
		expr, ok := n.(*ast.BinaryExpr)
		if !ok {
			return true
		}

		if rand.Intn(100) < 50 {
			if newExpr := obfuscateSingleExpr(expr); newExpr != nil {
				*expr = *newExpr
			}
		}

		return true
	})
}

func obfuscateSingleExpr(expr *ast.BinaryExpr) *ast.BinaryExpr {
	// Avoid obfuscating string concatenation
	if (expr.Op == token.ADD) {
		if lit, ok := expr.X.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			return nil
		}
		if lit, ok := expr.Y.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			return nil
		}
	}

	k := &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", rand.Intn(1000)+1)}

	switch expr.Op {
	case token.ADD:
		// A + B  =>  (A + K) + (B - K)
		return &ast.BinaryExpr{
			X: &ast.ParenExpr{X: &ast.BinaryExpr{X: expr.X, Op: token.ADD, Y: k}},
			Op: token.ADD,
			Y: &ast.ParenExpr{X: &ast.BinaryExpr{X: expr.Y, Op: token.SUB, Y: k}},
		}
	case token.SUB:
		// A - B  =>  (A + K) - (B + K)
		return &ast.BinaryExpr{
			X: &ast.ParenExpr{X: &ast.BinaryExpr{X: expr.X, Op: token.ADD, Y: k}},
			Op: token.SUB,
			Y: &ast.ParenExpr{X: &ast.BinaryExpr{X: expr.Y, Op: token.ADD, Y: k}},
		}
	// Add more cases for other operators like MUL, QUO if desired
	default:
		return nil
	}
}

func isStringLit(expr ast.Expr) bool {
	lit, ok := expr.(*ast.BasicLit)
	return ok && lit.Kind == token.STRING
}
