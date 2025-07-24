package obfuscator

import (
	"go/ast"
	"go/token"
	"math/rand"
	"strconv"
	"time"

	"golang.org/x/tools/go/ast/astutil"
)

const decryptFuncName = "o_d"

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

// EncryptStrings finds all string literals in the file and replaces them with a call
// to a dynamically injected decryption function. Each string is encrypted with a unique key.
func EncryptStrings(fset *token.FileSet, file *ast.File) error {
	rand.Seed(time.Now().UnixNano())
	hasEncryptedStrings := false

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

		hasEncryptedStrings = true
		randomKey := byte(rand.Intn(255) + 1) // Avoid key 0 for simplicity
		encryptedBytes := encryptBytes([]byte(originalString), randomKey)

		// If the string was a constant, we must change its declaration to var
		path, _ := astutil.PathEnclosingInterval(file, lit.Pos(), lit.End())
		if path != nil {
			for _, pnode := range path {
				if genDecl, ok := pnode.(*ast.GenDecl); ok && genDecl.Tok == token.CONST {
					genDecl.Tok = token.VAR
					break
				}
			}
		}

		// Create the AST nodes for the encrypted byte slice and the key
		byteSliceLit := &ast.CompositeLit{Type: &ast.ArrayType{Elt: ast.NewIdent("byte")}}
		for _, b := range encryptedBytes {
			byteSliceLit.Elts = append(byteSliceLit.Elts, &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(int(b))})
		}
		keyLit := &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(int(randomKey))}

		// Create the call expression to the decrypt function
		callExpr := &ast.CallExpr{
			Fun:  ast.NewIdent(decryptFuncName),
			Args: []ast.Expr{byteSliceLit, keyLit},
		}

		cursor.Replace(callExpr)
		return true
	}, nil)

	// If we have encrypted at least one string, we need to inject the decryption function
	if hasEncryptedStrings {
		decryptFunc := &ast.FuncDecl{
			Name: ast.NewIdent(decryptFuncName),
			Type: &ast.FuncType{
				Params: &ast.FieldList{List: []*ast.Field{
					{Names: []*ast.Ident{ast.NewIdent("data")}, Type: &ast.ArrayType{Elt: ast.NewIdent("byte")}},
					{Names: []*ast.Ident{ast.NewIdent("key")}, Type: ast.NewIdent("byte")},
				}},
				Results: &ast.FieldList{List: []*ast.Field{
					{Type: ast.NewIdent("string")},
				}},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
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
										X: &ast.BinaryExpr{
											X:  ast.NewIdent("b"),
											Op: token.SUB,
											Y:  ast.NewIdent("key"),
										},
										Op: token.XOR,
										Y:  ast.NewIdent("key"),
									},
								},
							},
						}},
					},
					&ast.ReturnStmt{Results: []ast.Expr{
						&ast.CallExpr{Fun: ast.NewIdent("string"), Args: []ast.Expr{ast.NewIdent("decrypted")}},
					}},
				},
			},
		}

		// Inject the decrypt function at the top level of the file
		file.Decls = append(file.Decls, decryptFunc)
	}

	return nil
}

// obfuscateDecryptFunc adds junk code to the decryption function to make it harder to analyze.
func obfuscateDecryptFunc(f *ast.FuncDecl) {
	// Example of junk code: a switch statement that does nothing.
	junkSwitch := &ast.SwitchStmt{
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.CaseClause{
					Body: []ast.Stmt{
						// Empty body, or could contain more junk
					},
				},
			},
		},
	}

	// Another example: an if statement that is always true
	junkIf := &ast.IfStmt{
		Cond: &ast.BasicLit{Kind: token.INT, Value: "1 == 1"}, // A tautology
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("_")},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"junk"`}},
				},
			},
		},
	}

	// Prepend the junk code to the function body
	f.Body.List = append([]ast.Stmt{junkSwitch, junkIf}, f.Body.List...)
}