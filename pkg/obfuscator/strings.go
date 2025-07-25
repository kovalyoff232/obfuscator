package obfuscator

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"go/ast"
	"go/token"
	"log"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

// StringEncryptionPass handles the string encryption process using AES-CTR.
type StringEncryptionPass struct {
	EncryptedStrings []string // Holds all encrypted strings from the package.
	Keys             [][]byte // Holds the AES key for each string.
	IVs              [][]byte // Holds the IV for each string.
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
		if call, ok := parent.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if x, ok := sel.X.(*ast.Ident); ok && x.Name == "fmt" {
					return true
				}
			}
		}

		switch pt := parent.(type) {
		case *ast.ImportSpec:
			return true
		case *ast.Field:
			if pt.Tag == node {
				return true
			}
		case *ast.CallExpr:
			if ident, ok := pt.Fun.(*ast.Ident); ok && ident.Name == "panic" {
				return true
			}
		}

		if len(node.Value) <= 2 {
			return true
		}

		unquoted, err := strconv.Unquote(node.Value)
		if err != nil {
			return true
		}

		// Encrypt the string and store it.
		encryptedData, key, iv := p.encryptStringAES(unquoted)
		if encryptedData == nil {
			// Encryption failed, skip this string.
			return true
		}
		stringIndex := len(p.EncryptedStrings)
		p.EncryptedStrings = append(p.EncryptedStrings, string(encryptedData))
		p.Keys = append(p.Keys, key)
		p.IVs = append(p.IVs, iv)

		// Create an expression to call the global decryption function.
		callExpr := &ast.CallExpr{
			Fun: ast.NewIdent(obf.StringDecryptionFuncName),
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(stringIndex)},
			},
		}

		cursor.Replace(callExpr)
		return false
	}, nil)

	return nil
}

// encryptStringAES performs AES-CTR encryption.
func (p *StringEncryptionPass) encryptStringAES(s string) ([]byte, []byte, []byte) {
	key := make([]byte, 16) // AES-128
	iv := make([]byte, 16)  // AES block size
	if _, err := rand.Read(key); err != nil {
		log.Printf("Warning: failed to generate random key: %v. Skipping string.", err)
		return nil, nil, nil
	}
	if _, err := rand.Read(iv); err != nil {
		log.Printf("Warning: failed to generate random IV: %v. Skipping string.", err)
		return nil, nil, nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("Warning: failed to create AES cipher: %v. Skipping string.", err)
		return nil, nil, nil
	}

	stream := cipher.NewCTR(block, iv)
	encrypted := make([]byte, len(s))
	stream.XORKeyStream(encrypted, []byte(s))

	return encrypted, key, iv
}
