package obfuscator
import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"go/ast"
	"go/token"
	"log"
	mrand "math/rand"
	"strconv"
	"golang.org/x/tools/go/ast/astutil"
)
// StringEncryptionPass handles the inlined string encryption process using AES-CTR.
type StringEncryptionPass struct {
	metaEngine *MetamorphicEngine
}
// NewStringEncryptionPass creates a new pass instance.
func NewStringEncryptionPass() *StringEncryptionPass {
	return &StringEncryptionPass{
		metaEngine: &MetamorphicEngine{},
	}
}
// Apply finds string literals and replaces them with a metamorphic, inlined, self-decrypting block of code.
func (p *StringEncryptionPass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		node, ok := cursor.Node().(*ast.BasicLit)
		if !ok || node.Kind != token.STRING {
			return true
		}
		unquoted, err := strconv.Unquote(node.Value)
		if err != nil || len(unquoted) == 0 {
			return true
		}
		// --- Safety Checks ---
		parent := cursor.Parent()
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
		encryptedData, key, iv := encryptStringAES(unquoted)
		if encryptedData == nil {
			return true
		}
		astutil.AddImport(fset, file, "crypto/aes")
		astutil.AddImport(fset, file, "crypto/cipher")
		decryptor := p.createMetamorphicDecryptor(obf, encryptedData, key, iv)
		if mrand.Intn(100) < 30 {
			astutil.AddImport(fset, file, "crypto/aes")
			astutil.AddImport(fset, file, "crypto/cipher")
			fakeData := make([]byte, len(encryptedData))
			copy(fakeData, encryptedData)
			for i := range fakeData {
				fakeData[i] ^= byte(0x5A) 
			}
			fake := p.createMetamorphicDecryptor(obf, fakeData, key, iv)
			opaque := &ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X:  &ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{&ast.CompositeLit{Type: &ast.ArrayType{Elt: ast.NewIdent("byte")}, Elts: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}, &ast.BasicLit{Kind: token.INT, Value: "2"}, &ast.BasicLit{Kind: token.INT, Value: "3"}}}}},
					Op: token.EQL,
					Y:  &ast.BasicLit{Kind: token.INT, Value: "4"},
				},
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.ExprStmt{X: fake},
				}},
			}
			_ = opaque
		}
		cursor.Replace(decryptor)
		return false
	}, nil)
	return nil
}
// createMetamorphicDecryptor generates a varied AST for a self-contained decryption block.
func (p *StringEncryptionPass) createMetamorphicDecryptor(obf *Obfuscator, encryptedData, key, iv []byte) *ast.CallExpr {
	createByteSliceLiteral := func(data []byte) *ast.CompositeLit {
		slice := &ast.CompositeLit{Type: &ast.ArrayType{Elt: ast.NewIdent("byte")}}
		for _, b := range data {
			slice.Elts = append(slice.Elts, &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("0x%x", b)})
		}
		return slice
	}
	dataVar, keyVar, ivVar, blockVar, streamVar, resultVar, errVar, iVar :=
		NewName(), NewName(), NewName(), NewName(), NewName(), NewName(), NewName(), NewName()
	weaveIdent := ast.NewIdent(obf.WeavingKeyVarName)
	if weaveIdent.Name == "" {
		weaveIdent = ast.NewIdent("weaveKeyFallback0")
	}
	var fallbackDecl ast.Stmt = nil
	if weaveIdent.Name == "weaveKeyFallback0" {
		fallbackDecl = &ast.DeclStmt{Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names:  []*ast.Ident{weaveIdent},
					Type:   ast.NewIdent("int64"),
					Values: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
				},
			},
		}}
	}
	keyWeavingLoop := &ast.ForStmt{
		Init: &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(iVar)}, Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
		},
		Cond: &ast.BinaryExpr{
			X:  ast.NewIdent(iVar),
			Op: token.LSS,
			Y:  &ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{ast.NewIdent(keyVar)}},
		},
		Post: &ast.IncDecStmt{X: ast.NewIdent(iVar), Tok: token.INC},
		Body: &ast.BlockStmt{List: []ast.Stmt{
			// key[i] ^= byte(((uint64(weavingKey) >> uint((i%8)*8)) ^ uint64(byte(i)*31) ^ uint64(byte(len(data))*17)) & 0xff)
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.IndexExpr{X: ast.NewIdent(keyVar), Index: ast.NewIdent(iVar)}},
				Tok: token.XOR_ASSIGN,
				Rhs: []ast.Expr{
					&ast.CallExpr{Fun: ast.NewIdent("byte"), Args: []ast.Expr{
						&ast.BinaryExpr{
							X: &ast.BinaryExpr{
								X: &ast.BinaryExpr{
									X: &ast.BinaryExpr{
										X: &ast.CallExpr{
											Fun:  ast.NewIdent("uint64"),
											Args: []ast.Expr{ast.NewIdent(obf.WeavingKeyVarName)},
										},
										Op: token.SHR,
										Y: &ast.CallExpr{
											Fun: ast.NewIdent("uint"),
											Args: []ast.Expr{&ast.ParenExpr{X: &ast.BinaryExpr{
												X:  &ast.BinaryExpr{X: ast.NewIdent(iVar), Op: token.REM, Y: &ast.BasicLit{Kind: token.INT, Value: "8"}},
												Op: token.MUL,
												Y:  &ast.BasicLit{Kind: token.INT, Value: "8"},
											}}},
										},
									},
									Op: token.XOR,
									Y: &ast.CallExpr{
										Fun: ast.NewIdent("uint64"),
										Args: []ast.Expr{&ast.CallExpr{
											Fun: ast.NewIdent("byte"),
											Args: []ast.Expr{&ast.BinaryExpr{
												X:  ast.NewIdent(iVar),
												Op: token.MUL,
												Y:  &ast.BasicLit{Kind: token.INT, Value: "31"},
											}},
										}},
									},
								},
								Op: token.XOR,
								Y: &ast.CallExpr{
									Fun: ast.NewIdent("uint64"),
									Args: []ast.Expr{&ast.CallExpr{
										Fun: ast.NewIdent("byte"),
										Args: []ast.Expr{&ast.BinaryExpr{
											X:  &ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{ast.NewIdent(dataVar)}},
											Op: token.MUL,
											Y:  &ast.BasicLit{Kind: token.INT, Value: "17"},
										}},
									}},
								},
							},
							Op: token.AND,
							Y:  &ast.BasicLit{Kind: token.INT, Value: "0xff"},
						},
					}},
				},
			},
		}},
	}
	// --- Metamorphic part: shuffle declaration order ---
	declarations := []ast.Stmt{
		&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent(keyVar)}, Tok: token.DEFINE, Rhs: []ast.Expr{createByteSliceLiteral(key)}},
		&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent(ivVar)}, Tok: token.DEFINE, Rhs: []ast.Expr{createByteSliceLiteral(iv)}},
		&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent(dataVar)}, Tok: token.DEFINE, Rhs: []ast.Expr{createByteSliceLiteral(encryptedData)}},
	}
	mrand.Shuffle(len(declarations), func(i, j int) {
		declarations[i], declarations[j] = declarations[j], declarations[i]
	})
	bodyStmts := []ast.Stmt{}
	if fallbackDecl != nil {
		bodyStmts = append(bodyStmts, fallbackDecl)
	}
	bodyStmts = append(bodyStmts, declarations...)
	bodyStmts = append(bodyStmts, p.metaEngine.GenerateJunkCodeBlock()...) // Junk
	bodyStmts = append(bodyStmts, keyWeavingLoop)                          // Weave
	bodyStmts = append(bodyStmts, &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{ast.NewIdent(ivVar)}},
			Op: token.GTR,
			Y:  &ast.BasicLit{Kind: token.INT, Value: "0"},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.IndexExpr{X: ast.NewIdent(ivVar), Index: &ast.BasicLit{Kind: token.INT, Value: "0"}}},
				Tok: token.XOR_ASSIGN,
				Rhs: []ast.Expr{&ast.CallExpr{
					Fun:  ast.NewIdent("byte"),
					Args: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{ast.NewIdent(dataVar)}}},
				}},
			},
		}},
	})
	bodyStmts = append(bodyStmts, &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(blockVar), ast.NewIdent(errVar)}, Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.CallExpr{
			Fun:  &ast.SelectorExpr{X: ast.NewIdent("aes"), Sel: ast.NewIdent("NewCipher")},
			Args: []ast.Expr{ast.NewIdent(keyVar)},
		}},
	})
	bodyStmts = append(bodyStmts, &ast.IfStmt{
		Cond: &ast.BinaryExpr{X: ast.NewIdent(errVar), Op: token.NEQ, Y: ast.NewIdent("nil")},
		Body: &ast.BlockStmt{List: []ast.Stmt{
			&ast.ExprStmt{X: &ast.CallExpr{Fun: ast.NewIdent("panic"), Args: []ast.Expr{ast.NewIdent(errVar)}}},
		}},
	})
	bodyStmts = append(bodyStmts, p.metaEngine.GenerateJunkCodeBlock()...) // More junk
	bodyStmts = append(bodyStmts, &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(resultVar)}, Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.CallExpr{
			Fun:  ast.NewIdent("make"),
			Args: []ast.Expr{&ast.ArrayType{Elt: ast.NewIdent("byte")}, &ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{ast.NewIdent(dataVar)}}},
		}},
	})
	bodyStmts = append(bodyStmts, &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(streamVar)}, Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.CallExpr{
			Fun:  &ast.SelectorExpr{X: ast.NewIdent("cipher"), Sel: ast.NewIdent("NewCTR")},
			Args: []ast.Expr{ast.NewIdent(blockVar), ast.NewIdent(ivVar)},
		}},
	})
	bodyStmts = append(bodyStmts, &ast.ExprStmt{X: &ast.CallExpr{
		Fun:  &ast.SelectorExpr{X: ast.NewIdent(streamVar), Sel: ast.NewIdent("XORKeyStream")},
		Args: []ast.Expr{ast.NewIdent(resultVar), ast.NewIdent(dataVar)},
	}})
	bodyStmts = append(bodyStmts, &ast.ReturnStmt{Results: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent("string"), Args: []ast.Expr{ast.NewIdent(resultVar)}}}})
	return &ast.CallExpr{
		Fun: &ast.FuncLit{
			Type: &ast.FuncType{
				Params:  &ast.FieldList{},
				Results: &ast.FieldList{List: []*ast.Field{{Type: ast.NewIdent("string")}}},
			},
			Body: &ast.BlockStmt{List: bodyStmts},
		},
	}
}
// encryptStringAES performs AES-CTR encryption.
func encryptStringAES(s string) ([]byte, []byte, []byte) {
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
	var sum uint32
	for i := 0; i < len(s); i++ {
		sum += uint32(byte(s[i]))
	}
	for i := range key {
		key[i] ^= byte((sum>>uint(i%24))&0xFF) ^ byte(len(s)+i*7)
	}
	for i := range iv {
		iv[i] ^= byte((sum>>uint((i+3)%24))&0xFF) ^ byte(len(s)*3+i*11)
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
