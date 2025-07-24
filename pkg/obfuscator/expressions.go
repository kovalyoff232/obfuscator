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
		// We are looking for binary expressions, like `a + b` or `x * y`.
		expr, ok := n.(*ast.BinaryExpr)
		if !ok {
			return true
		}

		// Apply obfuscation with a certain probability to avoid bloating the code too much.
		if rand.Intn(100) < 50 {
			if newExpr := obfuscateSingleExpr(expr); newExpr != nil {
				*expr = *newExpr
			}
		}

		return true
	})
}

// obfuscateSingleExpr takes a binary expression and returns a new, obfuscated one.
// It selects a transformation randomly from a set of possible transformations for each operator.
func obfuscateSingleExpr(expr *ast.BinaryExpr) *ast.BinaryExpr {
	// We must not obfuscate string concatenation, as it would lead to compilation errors.
	if isStringExpression(expr) {
		return nil
	}

	// Randomly pick a transformation technique.
	transformationIndex := rand.Intn(2)

	switch expr.Op {
	case token.ADD: // a + b
		if transformationIndex == 0 {
			// a + b  =>  a - (-b)
			return &ast.BinaryExpr{
				X:  expr.X,
				Op: token.SUB,
				Y: &ast.ParenExpr{
					X: &ast.UnaryExpr{Op: token.SUB, X: expr.Y},
				},
			}
		}
		// a + b  =>  (a^b) + 2*(a&b)
		return &ast.BinaryExpr{
			X: &ast.ParenExpr{
				X: &ast.BinaryExpr{X: expr.X, Op: token.XOR, Y: expr.Y},
			},
			Op: token.ADD,
			Y: &ast.BinaryExpr{
				X:  &ast.BasicLit{Kind: token.INT, Value: "2"},
				Op: token.MUL,
				Y: &ast.ParenExpr{
					X: &ast.BinaryExpr{X: expr.X, Op: token.AND, Y: expr.Y},
				},
			},
		}

	case token.SUB: // a - b
		// a - b  =>  a + (-b)
		return &ast.BinaryExpr{
			X:  expr.X,
			Op: token.ADD,
			Y: &ast.ParenExpr{
				X: &ast.UnaryExpr{Op: token.SUB, X: expr.Y},
			},
		}

	case token.XOR: // a ^ b
		if transformationIndex == 0 {
			// a ^ b => (a | b) - (a & b)
			return &ast.BinaryExpr{
				X: &ast.ParenExpr{
					X: &ast.BinaryExpr{X: expr.X, Op: token.OR, Y: expr.Y},
				},
				Op: token.SUB,
				Y: &ast.ParenExpr{
					X: &ast.BinaryExpr{X: expr.X, Op: token.AND, Y: expr.Y},
				},
			}
		}
		// a ^ b => (a & ^b) | (^a & b)
		return &ast.BinaryExpr{
			X: &ast.ParenExpr{
				X: &ast.BinaryExpr{
					X:  expr.X,
					Op: token.AND,
					Y:  &ast.UnaryExpr{Op: token.XOR, X: expr.Y},
				},
			},
			Op: token.OR,
			Y: &ast.ParenExpr{
				X: &ast.BinaryExpr{
					X:  &ast.UnaryExpr{Op: token.XOR, X: expr.X},
					Op: token.AND,
					Y:  expr.Y,
				},
			},
		}
	}

	return nil
}

// isStringExpression checks if a binary expression involves string literals.
// This is a helper to avoid applying arithmetic obfuscations to string concatenations.
func isStringExpression(expr *ast.BinaryExpr) bool {
	if expr.Op != token.ADD {
		return false
	}
	// A very basic check. A proper implementation would use type checking.
	if lit, ok := expr.X.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		return true
	}
	if lit, ok := expr.Y.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		return true
	}
	return false
}