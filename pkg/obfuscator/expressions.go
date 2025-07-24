package obfuscator

import (
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
	if expr.Op == token.ADD {
		if isStringLit(expr.X) || isStringLit(expr.Y) {
			return nil
		}
	}
	
	switch expr.Op {
	case token.ADD:
		return &ast.BinaryExpr{
			X:  expr.X,
			Op: token.SUB,
			Y: &ast.UnaryExpr{
				Op: token.SUB,
				X:  expr.Y,
			},
		}
	case token.SUB:
		return &ast.BinaryExpr{
			X:  expr.X,
			Op: token.ADD,
			Y: &ast.UnaryExpr{
				Op: token.SUB,
				X:  expr.Y,
			},
		}
	default:
		return nil
	}
}

func isStringLit(expr ast.Expr) bool {
	lit, ok := expr.(*ast.BasicLit)
	return ok && lit.Kind == token.STRING
}
