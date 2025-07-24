package obfuscator

import (
	"go/ast"
	"go/token"
	"math/rand"
	"strconv"
	"time"

	"golang.org/x/tools/go/ast/astutil"
)

// encryptBytesWithRollingKey performs encryption using a key that changes for each byte.
func encryptBytesWithRollingKey(data []byte, initialKey byte) []byte {
	result := make([]byte, len(data))
	key := initialKey
	for i, b := range data {
		encrypted := b ^ key
		result[i] = encrypted
		key = b // The next key is the original byte, creating a chain
	}
	return result
}

// EncryptStrings finds all string literals and replaces them with a function that decrypts them.
func EncryptStrings(fset *token.FileSet, file *ast.File) error {
	rand.Seed(time.Now().UnixNano())

	astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		lit, ok := cursor.Node().(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}

		if _, ok := cursor.Parent().(*ast.ImportSpec); ok {
			return true
		}
		if field, ok := cursor.Parent().(*ast.Field); ok {
			if field.Tag == lit {
				return true
			}
		}

		originalString, err := strconv.Unquote(lit.Value)
		if err != nil || originalString == "" {
			return true
		}

		// Split the key into two parts to make it harder to find in the binary
		keyPart1 := byte(rand.Intn(256))
		keyPart2 := byte(rand.Intn(256))
		initialKey := keyPart1 ^ keyPart2

		encryptedBytes := encryptBytesWithRollingKey([]byte(originalString), initialKey)

		// Replace the literal with an immediately-invoked function that decrypts it
		iife := createRollingDecryptor(encryptedBytes, keyPart1, keyPart2)
		cursor.Replace(iife)

		return true
	}, nil)

	return nil
}

// createRollingDecryptor generates the AST for an IIFE that decrypts data using a rolling key.
func createRollingDecryptor(encryptedData []byte, keyPart1, keyPart2 byte) *ast.CallExpr {
	byteSliceLit := &ast.CompositeLit{Type: &ast.ArrayType{Elt: ast.NewIdent("byte")}}
	for _, b := range encryptedData {
		byteSliceLit.Elts = append(byteSliceLit.Elts, &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(int(b))})
	}

	keyPart1Expr := &ast.CallExpr{
		Fun:  ast.NewIdent("byte"),
		Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(int(keyPart1))}},
	}
	keyPart2Expr := &ast.CallExpr{
		Fun:  ast.NewIdent("byte"),
		Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(int(keyPart2))}},
	}

	funcBody := &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent("data")}, Tok: token.DEFINE, Rhs: []ast.Expr{byteSliceLit}},
			&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent("k1")}, Tok: token.DEFINE, Rhs: []ast.Expr{keyPart1Expr}},
			&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent("k2")}, Tok: token.DEFINE, Rhs: []ast.Expr{keyPart2Expr}},
			&ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent("key")},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.BinaryExpr{X: ast.NewIdent("k1"), Op: token.XOR, Y: ast.NewIdent("k2")}},
			},
			&ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent("decrypted")},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.CallExpr{
					Fun:  ast.NewIdent("make"),
					Args: []ast.Expr{&ast.ArrayType{Elt: ast.NewIdent("byte")}, &ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{ast.NewIdent("data")}}},
				}},
			},
			&ast.RangeStmt{
				Key:   ast.NewIdent("i"),
				Value: ast.NewIdent("b"),
				Tok:   token.DEFINE,
				X:     ast.NewIdent("data"),
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.AssignStmt{
						Lhs: []ast.Expr{ast.NewIdent("originalByte")},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{&ast.BinaryExpr{X: ast.NewIdent("b"), Op: token.XOR, Y: ast.NewIdent("key")}},
					},
					&ast.AssignStmt{
						Lhs: []ast.Expr{&ast.IndexExpr{X: ast.NewIdent("decrypted"), Index: ast.NewIdent("i")}},
						Tok: token.ASSIGN,
						Rhs: []ast.Expr{ast.NewIdent("originalByte")},
					},
					&ast.AssignStmt{
						Lhs: []ast.Expr{ast.NewIdent("key")},
						Tok: token.ASSIGN,
						Rhs: []ast.Expr{ast.NewIdent("originalByte")},
					},
				}},
			},
			&ast.ReturnStmt{Results: []ast.Expr{
				&ast.CallExpr{Fun: ast.NewIdent("string"), Args: []ast.Expr{ast.NewIdent("decrypted")}},
			}},
		},
	}

	return &ast.CallExpr{
		Fun: &ast.FuncLit{
			Type: &ast.FuncType{
				Params:  &ast.FieldList{},
				Results: &ast.FieldList{List: []*ast.Field{{Type: ast.NewIdent("string")}}},
			},
			Body: funcBody,
		},
		Args: []ast.Expr{},
	}
}