package obfuscator

import (
	"crypto/rand"
	"go/ast"
	"go/token"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

// StringEncryptionPass handles the string encryption process.
// It collects all strings, stores them in a global encrypted slice,
// and replaces the original string literals with calls to a global decryption function.
type StringEncryptionPass struct {
	EncryptedStrings []string // Holds all encrypted strings from the package.
	StringKeys       [][]byte // Holds the multi-byte keys for each string.
}

// NewStringEncryptionPass creates a new pass instance.
func NewStringEncryptionPass() *StringEncryptionPass {
	return &StringEncryptionPass{}
}

// Apply finds all string literals, encrypts them, and replaces them with a call
// to the global decryption function. It skips strings used for logging/errors.
func (p *StringEncryptionPass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	// This pass should only run on the main package to ensure the decryption logic is present.
	if file.Name.Name != "main" {
		return nil
	}

	astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		node, ok := cursor.Node().(*ast.BasicLit)
		if !ok || node.Kind != token.STRING {
			return true
		}

		// --- Safety Checks ---
		parent := cursor.Parent()
		if call, ok := parent.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if x, ok := sel.X.(*ast.Ident); ok && x.Name == "fmt" {
					// It's a call to a function in the fmt package, skip it.
					// This is a broad but safe heuristic to avoid breaking print statements.
					return true
				}
			}
		}

		switch pt := parent.(type) {
		case *ast.ImportSpec: // Cannot replace import paths: import "path"
			return true
		case *ast.Field: // Cannot replace struct tags: `json:"my_tag"`
			if pt.Tag == node {
				return true
			}
		// Don't encrypt panic messages
		case *ast.CallExpr:
			if ident, ok := pt.Fun.(*ast.Ident); ok && ident.Name == "panic" {
				return true
			}
		}

		// Don't encrypt empty strings.
		if len(node.Value) <= 2 {
			return true
		}

		unquoted, err := strconv.Unquote(node.Value)
		if err != nil {
			return true // Should not happen with valid string literals.
		}

		// Encrypt the string and store it.
		encryptedData, key := p.encryptString(unquoted)
		stringIndex := len(p.EncryptedStrings)
		p.EncryptedStrings = append(p.EncryptedStrings, encryptedData)
		p.StringKeys = append(p.StringKeys, key)

		// Create an expression to call the global decryption function.
		// E.g., o_decrypt(12)
		callExpr := &ast.CallExpr{
			Fun: ast.NewIdent(obf.StringDecryptionFuncName),
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(stringIndex)},
			},
		}

		// Replace the string literal with the call expression.
		cursor.Replace(callExpr)

		return false // We replaced the node.
	}, nil)

	return nil
}



// encryptString performs a rolling XOR encryption with a multi-byte key.
func (p *StringEncryptionPass) encryptString(s string) (string, []byte) {
	key := generateKey(8) // Generate an 8-byte key
	data := []byte(s)
	encrypted := make([]byte, len(data))
	for i, b := range data {
		encrypted[i] = b ^ key[i%len(key)]
	}
	return string(encrypted), key
}

// generateKey creates a random key of the specified size for XOR encryption.
func generateKey(size int) []byte {
	key := make([]byte, size)
	rand.Read(key)
	// Ensure key is not all zeros, which would result in no encryption.
	isAllZeros := true
	for _, b := range key {
		if b != 0 {
			isAllZeros = false
			break
		}
	}
	if isAllZeros {
		key[0] = 1 // or generate again
	}
	return key
}
