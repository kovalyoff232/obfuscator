package obfuscator

import (
	"crypto/rand"
	"fmt"
	"go/ast"
	"go/token"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

// EncryptStrings_v2 finds all string literals, encrypts them, and replaces them
// with an inlined, self-contained decryption block.
func EncryptStrings_v2(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	// We only want to perform this on the main package, where the anti-debug key is available.
	if file.Name.Name != "main" {
		return nil
	}

	astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		node, ok := cursor.Node().(*ast.BasicLit)
		if !ok || node.Kind != token.STRING {
			return true
		}

		// --- Safety Checks ---
		// Check the parent node to ensure we can safely replace a literal with an expression.
		parent := cursor.Parent()
		switch p := parent.(type) {
		case *ast.ImportSpec:
			// Cannot replace import paths: import "path"
			return true
		case *ast.Field:
			// Cannot replace struct tags: `json:"my_tag"`
			if p.Tag == node {
				return true
			}
		}

		// Don't encrypt empty strings
		if len(node.Value) <= 2 {
			return true
		}

		unquoted, err := strconv.Unquote(node.Value)
		if err != nil {
			return true // Should not happen with valid string literals
		}

		encryptedData, key1, key2 := encryptString(unquoted)

		// Create an immediately-invoked function expression (IIFE) to decrypt the string.
		decryptExpr := createInlineDecryptor(obf, encryptedData, key1, key2)

		// Replace the string literal with the decryption expression.
		cursor.Replace(decryptExpr)

		return false // We replaced the node, no need to traverse its children.
	}, nil)

	return nil
}

// createInlineDecryptor generates an AST for an immediately-invoked anonymous function
// that performs the decryption.
func createInlineDecryptor(obf *Obfuscator, data string, k1, k2 byte) ast.Expr {
	keyVar, decryptedVar, iVar, bVar, originalByteVar := NewName(), NewName(), NewName(), NewName(), NewName()

	anonFunc := &ast.FuncLit{
		Type: &ast.FuncType{
			Params:  &ast.FieldList{},
			Results: &ast.FieldList{List: []*ast.Field{{Type: ast.NewIdent("string")}}},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				// keyVar := byte(k1) ^ byte(k2)
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(keyVar)}, Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.BinaryExpr{
						X:  &ast.CallExpr{Fun: ast.NewIdent("byte"), Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("0x%x", k1)}}},
						Op: token.XOR,
						Y:  &ast.CallExpr{Fun: ast.NewIdent("byte"), Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("0x%x", k2)}}},
					}},
				},
				// if keyVar == 0 { keyVar = 1 }
				&ast.IfStmt{
					Cond: &ast.BinaryExpr{X: ast.NewIdent(keyVar), Op: token.EQL, Y: &ast.BasicLit{Kind: token.INT, Value: "0"}},
					Body: &ast.BlockStmt{List: []ast.Stmt{&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent(keyVar)}, Tok: token.ASSIGN, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}}}}},
				},
				// keyVar ^= byte(globalKey % 256)
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(keyVar)}, Tok: token.XOR_ASSIGN,
					Rhs: []ast.Expr{&ast.CallExpr{
						Fun: ast.NewIdent("byte"),
						Args: []ast.Expr{&ast.BinaryExpr{X: ast.NewIdent(obf.WeavingKeyVarName), Op: token.REM, Y: &ast.BasicLit{Kind: token.INT, Value: "256"}}},
					}},
				},
				// decryptedVar := make([]byte, len(data))
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(decryptedVar)}, Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.CallExpr{
						Fun: ast.NewIdent("make"),
						Args: []ast.Expr{
							&ast.ArrayType{Elt: ast.NewIdent("byte")},
							&ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", data)}}},
						},
					}},
				},
				// for i, b := range []byte(data) { ... }
				&ast.RangeStmt{
					Key: ast.NewIdent(iVar), Value: ast.NewIdent(bVar), Tok: token.DEFINE,
					X: &ast.CallExpr{Fun: ast.NewIdent("[]byte"), Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", data)}}},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							// originalByteVar := b ^ keyVar
							&ast.AssignStmt{
								Lhs: []ast.Expr{ast.NewIdent(originalByteVar)}, Tok: token.DEFINE,
								Rhs: []ast.Expr{&ast.BinaryExpr{X: ast.NewIdent(bVar), Op: token.XOR, Y: ast.NewIdent(keyVar)}},
							},
							// decryptedVar[i] = originalByteVar
							&ast.AssignStmt{
								Lhs: []ast.Expr{&ast.IndexExpr{X: ast.NewIdent(decryptedVar), Index: ast.NewIdent(iVar)}}, Tok: token.ASSIGN,
								Rhs: []ast.Expr{ast.NewIdent(originalByteVar)},
							},
							// keyVar = b
							&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent(keyVar)}, Tok: token.ASSIGN, Rhs: []ast.Expr{ast.NewIdent(bVar)}},
						},
					},
				},
				// return string(decryptedVar)
				&ast.ReturnStmt{Results: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent("string"), Args: []ast.Expr{ast.NewIdent(decryptedVar)}}}},
			},
		},
	}
	return &ast.CallExpr{Fun: anonFunc}
}

func encryptString(s string) (string, byte, byte) {
	data := []byte(s)
	key1, key2 := make([]byte, 1), make([]byte, 1)
	rand.Read(key1)
	rand.Read(key2)
	key := key1[0] ^ key2[0]
	if key == 0 {
		key = 1
	}
	encrypted := make([]byte, len(data))
	for i, b := range data {
		encrypted[i] = b ^ key
		key = encrypted[i]
	}
	return string(encrypted), key1[0], key2[0]
}

var EncryptStringsFunc = EncryptStrings_v2
