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
	StringKeys       []byte   // Holds the keys for each string.
}

// NewStringEncryptionPass creates a new pass instance.
func NewStringEncryptionPass() *StringEncryptionPass {
	return &StringEncryptionPass{}
}

// Apply finds all string literals, encrypts them, and replaces them with a call
// to the global decryption function.
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
		switch pt := parent.(type) {
		case *ast.ImportSpec: // Cannot replace import paths: import "path"
			return true
		case *ast.Field: // Cannot replace struct tags: `json:"my_tag"`
			if pt.Tag == node {
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

// encryptString performs a simple XOR encryption with a one-time key.
func (p *StringEncryptionPass) encryptString(s string) (string, byte) {
	key := generateKey()
	data := []byte(s)
	encrypted := make([]byte, len(data))
	for i, b := range data {
		encrypted[i] = b ^ key
	}
	return string(encrypted), key
}

// generateKey creates a random byte for XOR encryption.
func generateKey() byte {
	b := make([]byte, 1)
	rand.Read(b)
	// Avoid a key of 0, as it doesn't encrypt.
	if b[0] == 0 {
		return 1
	}
	return b[0]
}
