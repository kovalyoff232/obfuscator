package obfuscator
import (
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/ast/astutil"
)
// ObfuscateExpressions traverses the AST and replaces simple binary expressions
// with more complex, but functionally equivalent, forms.
func ObfuscateExpressions(file *ast.File, info *types.Info) {
	astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		node := cursor.Node()
		binaryExpr, ok := node.(*ast.BinaryExpr)
		if !ok {
			return true
		}
		if randInt(2) == 0 {
			return true
		}
		if info == nil {
			return true
		}
		var newExpr ast.Expr
		template := randInt(2)
		// Check if it's an integer operation
		if t, ok := info.TypeOf(binaryExpr.X).(*types.Basic); ok && t.Info()&types.IsInteger != 0 {
			switch binaryExpr.Op {
			case token.ADD:
				newExpr = obfuscateAdd(binaryExpr.X, binaryExpr.Y, int(template))
			case token.SUB:
				newExpr = obfuscateSub(binaryExpr.X, binaryExpr.Y, int(template))
			case token.XOR:
				newExpr = obfuscateXor(binaryExpr.X, binaryExpr.Y, int(template))
			}
		}
		// Check if it's a boolean operation
		if t, ok := info.TypeOf(binaryExpr.X).(*types.Basic); ok && t.Info()&types.IsBoolean != 0 {
			switch binaryExpr.Op {
			case token.LAND:
				newExpr = obfuscateLand(binaryExpr.X, binaryExpr.Y, int(template))
			case token.LOR:
				newExpr = obfuscateLor(binaryExpr.X, binaryExpr.Y, int(template))
			}
		}
		if newExpr != nil {
			cursor.Replace(newExpr)
			return false
		}
		return true
	}, nil)
}
func obfuscateAdd(x, y ast.Expr, template int) ast.Expr {
	switch template {
	case 0:
		// a + b -> (a ^ b) + 2 * (a & b)
		return &ast.BinaryExpr{
			X: &ast.ParenExpr{X: &ast.BinaryExpr{X: x, Op: token.XOR, Y: y}},
			Op: token.ADD,
			Y: &ast.BinaryExpr{
				X:  &ast.BasicLit{Kind: token.INT, Value: "2"},
				Op: token.MUL,
				Y:  &ast.ParenExpr{X: &ast.BinaryExpr{X: x, Op: token.AND, Y: y}},
			},
		}
	default:
		// a + b -> (a | b) + (a & b)
		return &ast.BinaryExpr{
			X: &ast.ParenExpr{X: &ast.BinaryExpr{X: x, Op: token.OR, Y: y}},
			Op: token.ADD,
			Y: &ast.ParenExpr{X: &ast.BinaryExpr{X: x, Op: token.AND, Y: y}},
		}
	}
}
func obfuscateSub(x, y ast.Expr, template int) ast.Expr {
	switch template {
	case 0:
		// a - b -> a + ^b + 1
		return &ast.BinaryExpr{
			X:  &ast.BinaryExpr{X: x, Op: token.ADD, Y: &ast.UnaryExpr{Op: token.XOR, X: y}},
			Op: token.ADD,
			Y:  &ast.BasicLit{Kind: token.INT, Value: "1"},
		}
	default:
		// a - b -> (a^b) - 2*(!a & b)
		return &ast.BinaryExpr{
			X: &ast.ParenExpr{X: &ast.BinaryExpr{X: x, Op: token.XOR, Y: y}},
			Op: token.SUB,
			Y: &ast.BinaryExpr{
				X:  &ast.BasicLit{Kind: token.INT, Value: "2"},
				Op: token.MUL,
				Y: &ast.ParenExpr{X: &ast.BinaryExpr{
					X:  &ast.UnaryExpr{Op: token.XOR, X: x},
					Op: token.AND,
					Y:  y,
				}},
			},
		}
	}
}
func obfuscateXor(x, y ast.Expr, template int) ast.Expr {
	switch template {
	case 0:
		// a ^ b -> (a | b) &^ (a & b)
		return &ast.BinaryExpr{
			X:  &ast.ParenExpr{X: &ast.BinaryExpr{X: x, Op: token.OR, Y: y}},
			Op: token.AND_NOT,
			Y:  &ast.ParenExpr{X: &ast.BinaryExpr{X: x, Op: token.AND, Y: y}},
		}
	default:
		// a ^ b -> (a &^ b) | (b &^ a)
		return &ast.BinaryExpr{
			X:  &ast.ParenExpr{X: &ast.BinaryExpr{X: x, Op: token.AND_NOT, Y: y}},
			Op: token.OR,
						Y:  &ast.ParenExpr{X: &ast.BinaryExpr{X: y, Op: token.AND_NOT, Y: x}},
		}
	}
}
func obfuscateLand(x, y ast.Expr, template int) ast.Expr {
	switch template {
	case 0:
		// a && b -> !(!a || !b)
		return &ast.UnaryExpr{
			Op: token.NOT,
			X: &ast.ParenExpr{
				X: &ast.BinaryExpr{
					X:  &ast.UnaryExpr{Op: token.NOT, X: x},
					Op: token.LOR,
					Y:  &ast.UnaryExpr{Op: token.NOT, X: y},
				},
			},
		}
	default:
		// a && b -> a == b && a
		return &ast.BinaryExpr{
			X:  &ast.BinaryExpr{X: x, Op: token.EQL, Y: y},
			Op: token.LAND,
			Y:  x,
		}
	}
}
func obfuscateLor(x, y ast.Expr, template int) ast.Expr {
	switch template {
	case 0:
		// a || b -> !(!a && !b)
		return &ast.UnaryExpr{
			Op: token.NOT,
			X: &ast.ParenExpr{
				X: &ast.BinaryExpr{
					X:  &ast.UnaryExpr{Op: token.NOT, X: x},
					Op: token.LAND,
					Y:  &ast.UnaryExpr{Op: token.NOT, X: y},
				},
			},
		}
	default:
		// a || b -> a != b || a
		return &ast.BinaryExpr{
			X:  &ast.BinaryExpr{X: x, Op: token.NEQ, Y: y},
			Op: token.LOR,
			Y:  x,
		}
	}
}
