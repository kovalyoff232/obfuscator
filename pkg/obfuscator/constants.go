package obfuscator

import (
	"crypto/rand"
	"fmt"
	"go/ast"
	"go/token"
	"math/big"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

// ObfuscateConstants traverses the AST and replaces integer literals with
// more complex, functionally equivalent expressions.
func ObfuscateConstants(file *ast.File) {
	astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		node := cursor.Node()

		lit, ok := node.(*ast.BasicLit)
		if !ok || lit.Kind != token.INT {
			return true
		}

		// New check: Do not obfuscate integers inside a byte slice literal.
		if isInsideByteSlice(cursor) {
			return true
		}

		if isInsideConstDecl(file, cursor) {
			return true
		}

		if randInt(100) < 50 {
			return true
		}

		val, err := strconv.ParseInt(lit.Value, 0, 64)
		if err != nil {
			return true
		}

		if val >= -2 && val <= 2 {
			return true
		}

		newNode := generateObfuscatedIntExpr(val)
		cursor.Replace(newNode)

		return false
	}, nil)
}

// isInsideByteSlice checks if the cursor's current position is inside a byte slice literal.
func isInsideByteSlice(cursor *astutil.Cursor) bool {
	// A composite literal is something like `[]byte{1, 2, 3}`.
	// We check if the parent node is a composite literal.
	compLit, ok := cursor.Parent().(*ast.CompositeLit)
	if !ok {
		return false
	}

	// If it is, we check its type.
	// The type for `[]byte` is an ArrayType with Elt being an Ident with name "byte".
	if arrayType, ok := compLit.Type.(*ast.ArrayType); ok {
		if ident, ok := arrayType.Elt.(*ast.Ident); ok {
			return ident.Name == "byte"
		}
	}
	return false
}

// isInsideConstDecl checks if the cursor's current position is inside a 'const' block.
func isInsideConstDecl(file *ast.File, cursor *astutil.Cursor) bool {
	path, _ := astutil.PathEnclosingInterval(file, cursor.Node().Pos(), cursor.Node().End())
	if path == nil {
		return false
	}

	for _, node := range path {
		if decl, ok := node.(*ast.GenDecl); ok && decl.Tok == token.CONST {
			return true
		}
	}
	return false
}

// generateObfuscatedIntExpr creates a binary expression that evaluates to the original value.
func generateObfuscatedIntExpr(val int64) ast.Expr {
	k := randInt(1000) + 1 // A random integer to use in the expression.

	// Randomly choose one of the reliable obfuscation techniques.
	method := randInt(2)
	switch method {
	case 0:
		// Technique 1: val => (val ^ k) ^ k
		// This is a robust method using the properties of XOR.
		return &ast.BinaryExpr{
			X: &ast.ParenExpr{
				X: &ast.BinaryExpr{
					X:  &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", val)},
					Op: token.XOR,
					Y:  &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", k)},
				},
			},
			Op: token.XOR,
			Y:  &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", k)},
		}
	case 1:
		// Technique 2: val => (val + k) - k
		return &ast.BinaryExpr{
			X: &ast.ParenExpr{
				X: &ast.BinaryExpr{
					X:  &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", val)},
					Op: token.ADD,
					Y:  &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", k)},
				},
			},
			Op: token.SUB,
			Y:  &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", k)},
		}
	default:
		// Fallback, should not be reached.
		return &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", val)}
	}
}

// randInt generates a cryptographically random integer up to a max value.
func randInt(max int64) int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return max / 2 // Fallback
	}
	return n.Int64()
}