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

		encryptedData, key, iv := encryptStringAES(unquoted)
		if encryptedData == nil {
			return true
		}

		astutil.AddImport(fset, file, "crypto/aes")
		astutil.AddImport(fset, file, "crypto/cipher")

		decryptor := p.createMetamorphicDecryptor(encryptedData, key, iv)
		cursor.Replace(decryptor)

		return false
	}, nil)

	return nil
}

// createMetamorphicDecryptor generates a varied AST for a self-contained decryption block.
func (p *StringEncryptionPass) createMetamorphicDecryptor(encryptedData, key, iv []byte) *ast.CallExpr {
	createByteSliceLiteral := func(data []byte) *ast.CompositeLit {
		slice := &ast.CompositeLit{Type: &ast.ArrayType{Elt: ast.NewIdent("byte")}}
		for _, b := range data {
			slice.Elts = append(slice.Elts, &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("0x%x", b)})
		}
		return slice
	}

	dataVar, keyVar, ivVar, blockVar, streamVar, resultVar, errVar :=
		NewName(), NewName(), NewName(), NewName(), NewName(), NewName(), NewName()

	// --- Metamorphic part: shuffle declaration order ---
	declarations := []ast.Stmt{
		&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent(keyVar)}, Tok: token.DEFINE, Rhs: []ast.Expr{createByteSliceLiteral(key)}},
		&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent(ivVar)}, Tok: token.DEFINE, Rhs: []ast.Expr{createByteSliceLiteral(iv)}},
		&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent(dataVar)}, Tok: token.DEFINE, Rhs: []ast.Expr{createByteSliceLiteral(encryptedData)}},
	}
	mrand.Shuffle(len(declarations), func(i, j int) {
		declarations[i], declarations[j] = declarations[j], declarations[i]
	})

	// --- Metamorphic part: build the body with junk code ---
	bodyStmts := []ast.Stmt{}
	bodyStmts = append(bodyStmts, declarations...)
	bodyStmts = append(bodyStmts, p.metaEngine.GenerateJunkCodeBlock()...) // Junk code
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
	bodyStmts = append(bodyStmts, p.metaEngine.GenerateJunkCodeBlock()...) // More junk code
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
