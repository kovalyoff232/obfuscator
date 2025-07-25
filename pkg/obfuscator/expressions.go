package obfuscator

import (
	"go/ast"
	"go/token"
	"math/rand"
)

// ObfuscateExpressions traverses the AST and replaces binary expressions with more complex,
// but functionally equivalent forms.
func ObfuscateExpressions(node *ast.File) {
	ast.Inspect(node, func(n ast.Node) bool {
		expr, ok := n.(*ast.BinaryExpr)
		if !ok {
			return true
		}

		// We must not obfuscate string concatenation.
		// A proper implementation would use type checking, but this is a reasonable heuristic.
		if isStringExpression(expr) {
			return true
		}

		// Apply obfuscation with a certain probability to avoid bloating the code too much.
		if rand.Intn(100) < 50 {
			if newExpr := obfuscateSingleExpr(expr); newExpr != nil {
				*expr = *newExpr
				// Do not visit the children of the new expression,
				// otherwise we will obfuscate it again and again.
				return false
			}
		}

		return true
	})
}

// obfuscateSingleExpr takes a binary expression and returns a new, obfuscated one.
func obfuscateSingleExpr(expr *ast.BinaryExpr) *ast.BinaryExpr {
	// Each case contains a list of possible transformations for that operator.
	// A random one is chosen.
	switch expr.Op {
	case token.ADD: // a + b
		transformations := []func(x, y ast.Expr) *ast.BinaryExpr{
			// a + b  =>  a - (-b)
			func(x, y ast.Expr) *ast.BinaryExpr {
				return &ast.BinaryExpr{
					X:  x,
					Op: token.SUB,
					Y:  &ast.ParenExpr{X: &ast.UnaryExpr{Op: token.SUB, X: y}},
				}
			},
			// a + b  =>  (a^b) + 2*(a&b)
			func(x, y ast.Expr) *ast.BinaryExpr {
				return &ast.BinaryExpr{
					X: &ast.ParenExpr{X: &ast.BinaryExpr{X: x, Op: token.XOR, Y: y}},
					Op: token.ADD,
					Y: &ast.BinaryExpr{
						X:  &ast.BasicLit{Kind: token.INT, Value: "2"},
						Op: token.MUL,
						Y:  &ast.ParenExpr{X: &ast.BinaryExpr{X: x, Op: token.AND, Y: y}},
					},
				}
			},
		}
		return transformations[rand.Intn(len(transformations))](expr.X, expr.Y)

	case token.SUB: // a - b
		transformations := []func(x, y ast.Expr) *ast.BinaryExpr{
			// a - b  =>  a + (-b)
			func(x, y ast.Expr) *ast.BinaryExpr {
				return &ast.BinaryExpr{
					X:  x,
					Op: token.ADD,
					Y:  &ast.ParenExpr{X: &ast.UnaryExpr{Op: token.SUB, X: y}},
				}
			},
			// a - b => (a & ^b) - (^a & b)
			func(x, y ast.Expr) *ast.BinaryExpr {
				return &ast.BinaryExpr{
					X:  &ast.ParenExpr{X: &ast.BinaryExpr{X: x, Op: token.AND, Y: &ast.UnaryExpr{Op: token.XOR, X: y}}},
					Op: token.SUB,
					Y:  &ast.ParenExpr{X: &ast.BinaryExpr{X: &ast.UnaryExpr{Op: token.XOR, X: x}, Op: token.AND, Y: y}},
				}
			},
		}
		return transformations[rand.Intn(len(transformations))](expr.X, expr.Y)

	case token.XOR: // a ^ b
		transformations := []func(x, y ast.Expr) *ast.BinaryExpr{
			// a ^ b => (a | b) - (a & b)
			func(x, y ast.Expr) *ast.BinaryExpr {
				return &ast.BinaryExpr{
					X: &ast.ParenExpr{X: &ast.BinaryExpr{X: x, Op: token.OR, Y: y}},
					Op: token.SUB,
					Y: &ast.ParenExpr{X: &ast.BinaryExpr{X: x, Op: token.AND, Y: y}},
				}
			},
			// a ^ b => (a & ^b) | (^a & b)
			func(x, y ast.Expr) *ast.BinaryExpr {
				return &ast.BinaryExpr{
					X:  &ast.ParenExpr{X: &ast.BinaryExpr{X: x, Op: token.AND, Y: &ast.UnaryExpr{Op: token.XOR, X: y}}},
					Op: token.OR,
					Y:  &ast.ParenExpr{X: &ast.BinaryExpr{X: &ast.UnaryExpr{Op: token.XOR, X: x}, Op: token.AND, Y: y}},
				}
			},
		}
		return transformations[rand.Intn(len(transformations))](expr.X, expr.Y)

	case token.OR: // a | b
		// a | b => (a ^ b) + (a & b)
		return &ast.BinaryExpr{
			X: &ast.ParenExpr{X: &ast.BinaryExpr{X: expr.X, Op: token.XOR, Y: expr.Y}},
			Op: token.ADD,
			Y: &ast.ParenExpr{X: &ast.BinaryExpr{X: expr.X, Op: token.AND, Y: expr.Y}},
		}
	}

	return nil
}

func isStringExpression(expr *ast.BinaryExpr) bool {
	if expr.Op != token.ADD {
		return false
	}

	var isString bool
	ast.Inspect(expr, func(n ast.Node) bool {
		if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			isString = true
			return false // Stop inspecting
		}
		// Heuristic: if a function call is part of the expression, assume it could return a string.
		// This is broad, but safer than breaking the build.
		if _, ok := n.(*ast.CallExpr); ok {
			isString = true
			return false
		}
		return true
	})

	return isString
}
