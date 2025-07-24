package obfuscator

import (
	"crypto/rand"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

// EncryptStrings finds all string literals in the file, encrypts them,
// and replaces them with a call to a runtime decryption function.
// It uses the WeavingKeyVarName from the Obfuscator instance to tie the
// decryption to the anti-debugging checks.
func EncryptStrings(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	var stringLiterals []*ast.BasicLit
	var mainPkg bool
	if file.Name.Name == "main" {
		mainPkg = true
	}

	astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		node, ok := cursor.Node().(*ast.BasicLit)
		if ok && node.Kind == token.STRING {
			// Don't encrypt import paths
			if _, ok := cursor.Parent().(*ast.ImportSpec); ok {
				return true
			}
			stringLiterals = append(stringLiterals, node)
		}
		return true
	}, nil)

	if len(stringLiterals) == 0 {
		return nil
	}

	decryptFuncName := RandomIdentifier("o_decrypt")

	// Only add the decrypt function and key to the main package
	if mainPkg {
		addDecryptor(obf, file, decryptFuncName)
	}

	// Replace all string literals with calls to the decrypt function
	for _, lit := range stringLiterals {
		encryptedData, key1, key2 := encryptString(lit.Value)
		callExpr := &ast.CallExpr{
			Fun: ast.NewIdent(decryptFuncName),
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", encryptedData)},
				&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", key1)},
				&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", key2)},
			},
		}

		// To replace the node, we need to find its parent and the field it's in.
		// This is a bit complex but necessary for a correct replacement.
		path, _ := astutil.PathEnclosingInterval(file, lit.Pos(), lit.End())
		if path == nil {
			continue
		}
		// The first element is the literal itself, the second is its parent.
		if len(path) < 2 {
			continue
		}
		astutil.Apply(path[1], func(c *astutil.Cursor) bool {
			if c.Node() == lit {
				c.Replace(callExpr)
				return false // Stop after replacing
			}
			return true
		}, nil)
	}

	return nil
}

func encryptString(s string) (string, byte, byte) {
	unquoted, err := strconv.Unquote(s)
	if err != nil {
		unquoted = s // Fallback for non-standard strings
	}
	data := []byte(unquoted)

	key1 := make([]byte, 1)
	rand.Read(key1)
	key2 := make([]byte, 1)
	rand.Read(key2)

	key := key1[0] ^ key2[0]
	if key == 0 { // Avoid null key
		key = 1
	}

	encrypted := make([]byte, len(data))
	for i, b := range data {
		encrypted[i] = b ^ key
		key = encrypted[i] // Chaining
	}
	return string(encrypted), key1[0], key2[0]
}

// addDecryptor injects the runtime decryption function into the file.
// The generated function uses the global weaving key.
func addDecryptor(obf *Obfuscator, file *ast.File, funcName string) {
	// The Go code for the decryptor function as a string
	decryptorCode := fmt.Sprintf(`
func %s(data string, k1, k2 byte) string {
    key := k1 ^ k2
    if key == 0 {
        key = 1
    }

    // Weave the anti-debug check into the key
    key ^= byte(%s %% 256)

    decrypted := make([]byte, len(data))
    for i, b := range data {
        originalByte := b ^ key
        decrypted[i] = originalByte
        key = b // Use the encrypted byte for chaining
    }
    return string(decrypted)
}`, funcName, obf.WeavingKeyVarName)

	// We need to parse this string into an AST fragment
	fset := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fset, "", "package main\n"+decryptorCode, parser.ParseComments)
	if err != nil {
		panic(fmt.Sprintf("failed to parse decryptor code: %v", err)) // Should not happen
	}

	// Find the function declaration in the parsed file
	var decryptorDecl ast.Decl
	for _, decl := range parsedFile.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == funcName {
			decryptorDecl = fn
			break
		}
	}

	if decryptorDecl != nil {
		// Add the decryptor function to the target file's declarations
		file.Decls = append(file.Decls, decryptorDecl)
	}
}



// EncryptStrings_v2 is the primary function, replacing the original one.
func EncryptStrings_v2(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	// ... (same logic as EncryptStrings to find literals)
	var stringLiterals []*ast.BasicLit
	astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		node, ok := cursor.Node().(*ast.BasicLit)
		if ok && node.Kind == token.STRING {
			if _, ok := cursor.Parent().(*ast.ImportSpec); ok {
				return true
			}
			// Avoid encrypting the build tag
			if node.Value == `"//go:build linux"` {
				return true
			}
			stringLiterals = append(stringLiterals, node)
		}
		return true
	}, nil)

	if len(stringLiterals) == 0 {
		return nil
	}

	decryptFuncName := RandomIdentifier("o_decrypt")

	if file.Name.Name == "main" {
		addDecryptor_v2(obf, file, decryptFuncName)
	}

	for _, lit := range stringLiterals {
		unquoted, _ := strconv.Unquote(lit.Value)
		encryptedData, key1, key2 := encryptString(unquoted)

		// Replace the literal with a call to the decryptor
		call := &ast.CallExpr{
			Fun: ast.NewIdent(decryptFuncName),
			Args: []ast.Expr{
				&ast.CallExpr{
					Fun:  ast.NewIdent("string"),
					Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", encryptedData)}},
				},
				&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("0x%x", key1)},
				&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("0x%x", key2)},
			},
		}
		path, _ := astutil.PathEnclosingInterval(file, lit.Pos(), lit.End())
		if path == nil {
			continue
		}
		// The first element is the literal itself, the second is its parent.
		if len(path) < 2 {
			continue
		}
		astutil.Apply(path[1], func(c *astutil.Cursor) bool {
			if c.Node() == lit {
				c.Replace(call)
				return false // Stop after replacing
			}
			return true
		}, nil)
	}

	return nil
}

// addDecryptor_v2 creates the decryptor function AST directly.
func addDecryptor_v2(obf *Obfuscator, file *ast.File, funcName string) {
	// Manually create the AST for the decryption function
	fn := &ast.FuncDecl{
		Name: ast.NewIdent(funcName),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{Names: []*ast.Ident{ast.NewIdent("data")}, Type: ast.NewIdent("string")},
					{Names: []*ast.Ident{ast.NewIdent("k1")}, Type: ast.NewIdent("byte")},
					{Names: []*ast.Ident{ast.NewIdent("k2")}, Type: ast.NewIdent("byte")},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{Type: ast.NewIdent("string")},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				// key := k1 ^ k2
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("key")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.BinaryExpr{X: ast.NewIdent("k1"), Op: token.XOR, Y: ast.NewIdent("k2")}},
				},
				// if key == 0 { key = 1 }
				&ast.IfStmt{
					Cond: &ast.BinaryExpr{X: ast.NewIdent("key"), Op: token.EQL, Y: &ast.BasicLit{Kind: token.INT, Value: "0"}},
					Body: &ast.BlockStmt{List: []ast.Stmt{
						&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent("key")}, Tok: token.ASSIGN, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}}},
					}},
				},
				// key ^= byte(WEAVING_KEY_VAR % 256)
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("key")},
					Tok: token.XOR_ASSIGN,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: ast.NewIdent("byte"),
							Args: []ast.Expr{
								&ast.BinaryExpr{
									X:  ast.NewIdent(obf.WeavingKeyVarName),
									Op: token.REM,
									Y:  &ast.BasicLit{Kind: token.INT, Value: "256"},
								},
							},
						},
					},
				},
				// decrypted := make([]byte, len(data))
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("decrypted")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: ast.NewIdent("make"),
							Args: []ast.Expr{
								&ast.ArrayType{Elt: ast.NewIdent("byte")},
								&ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{ast.NewIdent("data")}},
							},
						},
					},
				},
				// for i, b := range data { ... }
				&ast.RangeStmt{
					Key:   ast.NewIdent("i"),
					Value: ast.NewIdent("b"),
					Tok:   token.DEFINE,
					X:     &ast.CallExpr{Fun: ast.NewIdent("[]byte"), Args: []ast.Expr{ast.NewIdent("data")}},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							// originalByte := b ^ key
							&ast.AssignStmt{
								Lhs: []ast.Expr{ast.NewIdent("originalByte")},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{
									&ast.BinaryExpr{
										X:  ast.NewIdent("b"),
										Op: token.XOR,
										Y:  ast.NewIdent("key"),
									},
								},
							},
							// decrypted[i] = originalByte
							&ast.AssignStmt{
								Lhs: []ast.Expr{&ast.IndexExpr{X: ast.NewIdent("decrypted"), Index: ast.NewIdent("i")}},
								Tok: token.ASSIGN,
								Rhs: []ast.Expr{ast.NewIdent("originalByte")},
							},
							// key = b
							&ast.AssignStmt{
								Lhs: []ast.Expr{ast.NewIdent("key")},
								Tok: token.ASSIGN,
								Rhs: []ast.Expr{ast.NewIdent("b")},
							},
						},
					},
				},
				// return string(decrypted)
				&ast.ReturnStmt{
					Results: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent("string"), Args: []ast.Expr{ast.NewIdent("decrypted")}}},
				},
			},
		},
	}
	file.Decls = append(file.Decls, fn)
}

// We need to replace the original EncryptStrings with the new one in the pass.
// The original implementation is kept for reference.
var _ = EncryptStrings

// The actual pass will call this function.
var EncryptStringsFunc = EncryptStrings_v2
