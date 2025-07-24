package obfuscator

import (
	"go/ast"
	"go/token"
	"math/rand"
)

// ObfuscateExpressions обходит AST и заменяет простые бинарные выражения на более сложные.
func ObfuscateExpressions(node *ast.File) {
	ast.Inspect(node, func(n ast.Node) bool {
		// Ищем бинарные выражения, например `a + b`.
		expr, ok := n.(*ast.BinaryExpr)
		if !ok {
			return true
		}

		// С некоторой вероятностью заменяем выражение.
		if rand.Intn(100) < 50 {
			if newExpr := obfuscateSingleExpr(expr); newExpr != nil {
				*expr = *newExpr
			}
		}

		return true
	})
}

// obfuscateSingleExpr принимает бинарное выражение и возвращает его обфусцированную версию.
func obfuscateSingleExpr(expr *ast.BinaryExpr) *ast.BinaryExpr {
	switch expr.Op {
	case token.ADD:
		// a + b  ->  a - (-b)
		return &ast.BinaryExpr{
			X:  expr.X,
			Op: token.SUB,
			Y: &ast.UnaryExpr{
				Op: token.SUB,
				X:  expr.Y,
			},
		}
	case token.SUB:
		// a - b  ->  a + (-b)
		return &ast.BinaryExpr{
			X:  expr.X,
			Op: token.ADD,
			Y: &ast.UnaryExpr{
				Op: token.SUB,
				X:  expr.Y,
			},
		}
	// Можно добавить больше правил для других операторов, например, XOR.
	// case token.XOR:
	// 	// a ^ b -> (a & ^b) | (^a & b)
	// 	return &ast.BinaryExpr{
	// 		X: &ast.ParenExpr{X: &ast.BinaryExpr{X: expr.X, Op: token.AND_NOT, Y: expr.Y}},
	// 		Op: token.OR,
	// 		Y: &ast.ParenExpr{X: &ast.BinaryExpr{X: &ast.UnaryExpr{Op: token.XOR, X: expr.X}, Op: token.AND, Y: expr.Y}},
	// 	}
	default:
		return nil
	}
}