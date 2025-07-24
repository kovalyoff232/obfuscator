package obfuscator

import (
	"go/ast"
	"go/token"
	"math/rand"
	"strconv"
	"time"

	"golang.org/x/tools/go/ast/astutil"
)

// encryptBytes performs a simple two-step encryption on a byte slice.
func encryptBytes(data []byte, key byte) []byte {
	result := make([]byte, len(data))
	for i, b := range data {
		// Step 1: XOR with the key
		encrypted := b ^ key
		// Step 2: Add a constant value (e.g., the key itself)
		result[i] = encrypted + key
	}
	return result
}

// EncryptStrings finds all string literals in the file and replaces them with
// an inlined, immediately-invoked function that decrypts the string at runtime.
func EncryptStrings(fset *token.FileSet, file *ast.File) error {
	rand.Seed(time.Now().UnixNano())

	astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		lit, ok := cursor.Node().(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}

		// Skip import paths
		if _, ok := cursor.Parent().(*ast.ImportSpec); ok {
			return true
		}

		// Skip struct tags
		if field, ok := cursor.Parent().(*ast.Field); ok {
			if field.Tag == lit {
				return true
			}
		}

		originalString, err := strconv.Unquote(lit.Value)
		if err != nil || originalString == "" {
			return true
		}

		randomKey := byte(rand.Intn(255) + 1) // Avoid key 0 for simplicity
		encryptedBytes := encryptBytes([]byte(originalString), randomKey)

		// If the string was a constant, we must change its declaration to var
		path, _ := astutil.PathEnclosingInterval(file, lit.Pos(), lit.End())
		if path != nil {
			for _, pnode := range path {
				if genDecl, ok := pnode.(*ast.GenDecl); ok && genDecl.Tok == token.CONST {
					genDecl.Tok = token.VAR
					genDecl.Doc = nil
					break
				}
			}
		}

		// Replace the string literal with an immediately-invoked function expression (IIFE).
		// This inlines the decryption logic.
		iife := createInlineDecryptor(encryptedBytes, randomKey)
		cursor.Replace(iife)

		return true
	}, nil)

	// No need to inject a global decrypt function anymore.
	return nil
}

// createInlineDecryptor generates the AST for an immediately-invoked function
// that decrypts the given data.
func createInlineDecryptor(encryptedData []byte, key byte) *ast.CallExpr {
	// AST for the encrypted byte slice literal
	byteSliceLit := &ast.CompositeLit{Type: &ast.ArrayType{Elt: ast.NewIdent("byte")}}
	for _, b := range encryptedData {
		byteSliceLit.Elts = append(byteSliceLit.Elts, &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(int(b))})
	}

	// AST for `byte(KEY_VALUE)`
	keyExpr := &ast.CallExpr{
		Fun:  ast.NewIdent("byte"),
		Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(int(key))}},
	}

	// The body of the anonymous function
	funcBody := &ast.BlockStmt{
		List: []ast.Stmt{
			// data := []byte{...}
			&ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent("data")},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{byteSliceLit},
			},
			// key := byte(...)
			&ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent("key")},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{keyExpr},
			},
			// decrypted := make([]byte, len(data))
			&ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent("decrypted")},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.CallExpr{
					Fun: ast.NewIdent("make"),
					Args: []ast.Expr{
						&ast.ArrayType{Elt: ast.NewIdent("byte")},
						&ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{ast.NewIdent("data")}},
					},
				}},
			},
			// for i, b := range data { ... }
			&ast.RangeStmt{
				Key:   ast.NewIdent("i"),
				Value: ast.NewIdent("b"),
				Tok:   token.DEFINE,
				X:     ast.NewIdent("data"),
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.AssignStmt{
						Lhs: []ast.Expr{&ast.IndexExpr{X: ast.NewIdent("decrypted"), Index: ast.NewIdent("i")}},
						Tok: token.ASSIGN,
						Rhs: []ast.Expr{
							// (b - key) ^ key
							&ast.BinaryExpr{
								X: &ast.ParenExpr{X: &ast.BinaryExpr{
									X:  ast.NewIdent("b"),
									Op: token.SUB,
									Y:  ast.NewIdent("key"),
								}},
								Op: token.XOR,
								Y:  ast.NewIdent("key"),
							},
						},
					},
				}},
			},
			// return string(decrypted)
			&ast.ReturnStmt{Results: []ast.Expr{
				&ast.CallExpr{Fun: ast.NewIdent("string"), Args: []ast.Expr{ast.NewIdent("decrypted")}},
			}},
		},
	}

	// The anonymous function literal
	funcLit := &ast.FuncLit{
		Type: &ast.FuncType{
			Params:  &ast.FieldList{}, // No parameters
			Results: &ast.FieldList{List: []*ast.Field{{Type: ast.NewIdent("string")}}},
		},
		Body: funcBody,
	}

	// The call expression that invokes the anonymous function immediately
	return &ast.CallExpr{
		Fun:  funcLit,
		Args: []ast.Expr{}, // No arguments
	}
}