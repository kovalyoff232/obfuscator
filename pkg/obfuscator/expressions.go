package obfuscator

import (
	"go/ast"
	"go/token"
	"go/types"
	"math/rand"

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

		// 50% chance to obfuscate any given expression to avoid overly complex code.
		if rand.Intn(2) == 0 {
			return true
		}

		// Check if we have type info. If not, we can't safely obfuscate.
		if info == nil {
			return true
		}

		// Check the type of the expression. We only want to obfuscate integer operations.
		if t, ok := info.TypeOf(binaryExpr.X).(*types.Basic); ok {
			if t.Info()&types.IsInteger == 0 {
				return true // Not an integer, don't obfuscate.
			}
		} else {
			return true // Not a basic type, can't be sure it's an integer.
		}

		template := rand.Intn(3)
		var newExpr ast.Expr

		switch binaryExpr.Op {
		case token.ADD:
			newExpr = obfuscateAdd(binaryExpr.X, binaryExpr.Y, template)
		case token.SUB:
			newExpr = obfuscateSub(binaryExpr.X, binaryExpr.Y, template)
		case token.XOR:
			newExpr = obfuscateXor(binaryExpr.X, binaryExpr.Y, template)
		}

		if newExpr != nil {
			// Safely replace the current node with the new expression.
			cursor.Replace(newExpr)
			// We replaced the node, so we shouldn't process its children further.
			return false
		}

		return true
	}, nil)
}

// a + b -> a - (-b)
func add_transform1(x, y ast.Expr) ast.Expr {
	return &ast.BinaryExpr{
		X:  x,
		Op: token.SUB,
		Y: &ast.ParenExpr{
			X: &ast.UnaryExpr{Op: token.SUB, X: y},
		},
	}
}

// a + b -> (a | b) + (a & b)
func add_transform2(x, y ast.Expr) ast.Expr {
	return &ast.BinaryExpr{
		X: &ast.ParenExpr{
			X: &ast.BinaryExpr{X: x, Op: token.OR, Y: y},
		},
		Op: token.ADD,
		Y: &ast.ParenExpr{
			X: &ast.BinaryExpr{X: x, Op: token.AND, Y: y},
		},
	}
}

// a + b -> 2*(a&b) + (a^b)
func add_transform3(x, y ast.Expr) ast.Expr {
	return &ast.BinaryExpr{
		X: &ast.BinaryExpr{
			X:  &ast.BasicLit{Kind: token.INT, Value: "2"},
			Op: token.MUL,
			Y: &ast.ParenExpr{
				X: &ast.BinaryExpr{X: x, Op: token.AND, Y: y},
			},
		},
		Op: token.ADD,
		Y: &ast.ParenExpr{
			X: &ast.BinaryExpr{X: x, Op: token.XOR, Y: y},
		},
	}
}

// a - b -> a + (-b)
func sub_transform1(x, y ast.Expr) ast.Expr {
	return &ast.BinaryExpr{
		X:  x,
		Op: token.ADD,
		Y: &ast.ParenExpr{
			X: &ast.UnaryExpr{Op: token.SUB, X: y},
		},
	}
}

// a - b -> (a^b) - 2*(!a & b)
func sub_transform2(x, y ast.Expr) ast.Expr {
	return &ast.BinaryExpr{
		X: &ast.ParenExpr{
			X: &ast.BinaryExpr{X: x, Op: token.XOR, Y: y},
		},
		Op: token.SUB,
		Y: &ast.BinaryExpr{
			X:  &ast.BasicLit{Kind: token.INT, Value: "2"},
			Op: token.MUL,
			Y: &ast.ParenExpr{
				X: &ast.BinaryExpr{
					X:  &ast.UnaryExpr{Op: token.XOR, X: x}, // Bitwise NOT is ^ in Go
					Op: token.AND,
					Y:  y,
				},
			},
		},
	}
}

// a - b -> a + ^b + 1
func sub_transform3(x, y ast.Expr) ast.Expr {
	return &ast.BinaryExpr{
		X: &ast.BinaryExpr{
			X:  x,
			Op: token.ADD,
			Y:  &ast.UnaryExpr{Op: token.XOR, X: y},
		},
		Op: token.ADD,
		Y:  &ast.BasicLit{Kind: token.INT, Value: "1"},
	}
}

// a ^ b -> (a|b) &^ (a&b)
func xor_transform1(x, y ast.Expr) ast.Expr {
	return &ast.BinaryExpr{
		X: &ast.ParenExpr{
			X: &ast.BinaryExpr{X: x, Op: token.OR, Y: y},
		},
		Op: token.AND_NOT, // AND NOT
		Y: &ast.ParenExpr{
			X: &ast.BinaryExpr{X: x, Op: token.AND, Y: y},
		},
	}
}

// a ^ b -> (a &^ b) | (b &^ a)
func xor_transform2(x, y ast.Expr) ast.Expr {
	return &ast.BinaryExpr{
		X: &ast.ParenExpr{
			X: &ast.BinaryExpr{X: x, Op: token.AND_NOT, Y: y},
		},
		Op: token.OR,
		Y: &ast.ParenExpr{
			X: &ast.BinaryExpr{X: y, Op: token.AND_NOT, Y: x},
		},
	}
}

func obfuscateAdd(x, y ast.Expr, template int) ast.Expr {
	switch template {
	case 0:
		return add_transform1(x, y)
	case 1:
		return add_transform2(x, y)
	default:
		return add_transform3(x, y)
	}
}

func obfuscateSub(x, y ast.Expr, template int) ast.Expr {
	switch template {
	case 0:
		return sub_transform1(x, y)
	case 1:
		return sub_transform2(x, y)
	default:
		return sub_transform3(x, y)
	}
}

func obfuscateXor(x, y ast.Expr, template int) ast.Expr {
	switch template {
	case 0:
		return xor_transform1(x, y)
	default:
		return xor_transform2(x, y)
	}
}