package obfuscator

import (
	"fmt"
	"go/ast"
	"go/token"
	"math/rand"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

// ObfuscateConstants traverses the AST and replaces integer literals with
// equivalent binary expressions to hide the original values.
func ObfuscateConstants(file *ast.File) {
	astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		node := cursor.Node()

		// We are interested in basic literals, specifically integers.
		lit, ok := node.(*ast.BasicLit)
		if !ok || lit.Kind != token.INT {
			return true // Continue traversal.
		}

		// It's crucial to avoid obfuscating constants in 'const' declarations,
		// as they often require a value that can be evaluated at compile time.
		// Our generated expressions can only be evaluated at runtime.
		if isInsideConstDecl(file, cursor) {
			return true // Skip this one, but continue.
		}

		// Don't obfuscate with a 100% probability.
		if rand.Intn(100) < 50 {
			return true
		}

		val, err := strconv.ParseInt(lit.Value, 10, 64)
		if err != nil {
			return true // Not a valid integer, skip.
		}

		// Avoid obfuscating small, common numbers like 0, 1, 2.
		if val >= -2 && val <= 2 {
			return true
		}

		// Replace the integer literal with a generated expression.
		newNode := generateObfuscatedIntExpr(val)
		cursor.Replace(newNode)

		// We replaced the node, so we should not traverse its children
		// because the new node is already obfuscated.
		return false
	}, nil)
}

// isInsideConstDecl checks if the cursor's current position is inside a 'const' block.
func isInsideConstDecl(file *ast.File, cursor *astutil.Cursor) bool {
	// astutil.PathEnclosingInterval is the correct way to get the chain of nodes
	// from the root of the AST to the current node's position.
	path, _ := astutil.PathEnclosingInterval(file, cursor.Node().Pos(), cursor.Node().End())
	if path == nil {
		return false // Should not happen in practice
	}

	// We iterate over the path (from the current node up to the root)
	// to see if any of the parent nodes is a 'const' declaration.
	for _, node := range path {
		// A GenDecl (generic declaration) with the token CONST represents a const block.
		if decl, ok := node.(*ast.GenDecl); ok && decl.Tok == token.CONST {
			return true
		}
	}
	return false
}

// generateObfuscatedIntExpr creates a binary expression that evaluates to the original value.
func generateObfuscatedIntExpr(val int64) ast.Expr {
	k := rand.Int63n(1000) + 1 // A random integer to use in the expression.

	// Randomly choose one of the obfuscation techniques.
	method := rand.Intn(3) // Increased to 3 for more variety
	switch method {
	case 0:
		// Technique 1: val => (val + k) - k
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
	case 1:
		// Technique 2: val => (val ^ k) ^ k
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
	case 2:
		// Technique 3: val => (val - k) + k
		return &ast.BinaryExpr{
			X: &ast.ParenExpr{
				X: &ast.BinaryExpr{
					X:  &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", val)},
					Op: token.SUB,
					Y:  &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", k)},
				},
			},
			Op: token.ADD,
			Y:  &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", k)},
		}
	default:
		// Fallback, should not be reached.
		return &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", val)}
	}
}