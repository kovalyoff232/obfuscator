package obfuscator

import (
	"fmt"
	"go/ast"
	"go/token"
	"math/rand"
	"strconv"
	"time"

	"golang.org/x/tools/go/ast/astutil"
)

const decryptFuncName = "o_d"

// xorEncrypt остается без изменений.
func xorEncrypt(data []byte, key byte) []byte {
	result := make([]byte, len(data))
	for i, b := range data {
		result[i] = b ^ key
	}
	return result
}

// EncryptStrings - новая, полиморфная версия.
func EncryptStrings(file *ast.File) error {
	rand.Seed(time.Now().UnixNano())

	// 1. Генерируем "осколки" ключа и их объявления.
	numParts := 3 // Количество частей, из которых будет состоять ключ.
	keyParts := make([]byte, numParts)
	keyPartNames := make([]string, numParts)
	var keyPartDecls []ast.Decl

	for i := 0; i < numParts; i++ {
		keyParts[i] = byte(rand.Intn(256))
		keyPartNames[i] = fmt.Sprintf("o_k_%d", i)
		
		decl := &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names:  []*ast.Ident{ast.NewIdent(keyPartNames[i])},
					Values: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(int(keyParts[i]))}},
				},
			},
		}
		keyPartDecls = append(keyPartDecls, decl)
	}

	// 2. Генерируем полиморфную формулу и вычисляем итоговый ключ.
	operators := []token.Token{token.ADD, token.SUB, token.XOR}
	
	// Собираем формулу в виде AST
	formula := ast.Expr(ast.NewIdent(keyPartNames[0]))
	finalKey := keyParts[0]

	for i := 1; i < numParts; i++ {
		op := operators[rand.Intn(len(operators))]
		formula = &ast.BinaryExpr{
			X:  formula,
			Op: op,
			Y:  ast.NewIdent(keyPartNames[i]),
		}
		switch op {
		case token.ADD:
			finalKey += keyParts[i]
		case token.SUB:
			finalKey -= keyParts[i]
		case token.XOR:
			finalKey ^= keyParts[i]
		}
	}

	// 3. Ищем и заменяем строки, используя `finalKey`.
	hasEncryptedStrings := false
	astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		lit, ok := cursor.Node().(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING { return true }
		if _, ok := cursor.Parent().(*ast.ImportSpec); ok { return true }

		originalString, err := strconv.Unquote(lit.Value)
		if err != nil || originalString == "" { return true }

		hasEncryptedStrings = true
		encryptedBytes := xorEncrypt([]byte(originalString), finalKey)

		path, _ := astutil.PathEnclosingInterval(file, lit.Pos(), lit.End())
		if path != nil {
			for _, pnode := range path {
				if genDecl, ok := pnode.(*ast.GenDecl); ok && genDecl.Tok == token.CONST {
					genDecl.Tok = token.VAR
					break
				}
			}
		}

		byteSliceLit := &ast.CompositeLit{Type: &ast.ArrayType{Elt: ast.NewIdent("byte")}}
		for _, b := range encryptedBytes {
			byteSliceLit.Elts = append(byteSliceLit.Elts, &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(int(b))})
		}
		callExpr := &ast.CallExpr{
			Fun:  ast.NewIdent(decryptFuncName),
			Args: []ast.Expr{byteSliceLit},
		}
		cursor.Replace(callExpr)
		return true
	}, nil)

	// 4. Если нужно, внедряем глобальные переменные и новую функцию-дешифратор.
	if hasEncryptedStrings {
		// Создаем функцию-дешифратор программно
		decryptFunc := &ast.FuncDecl{
			Name: ast.NewIdent(decryptFuncName),
			Type: &ast.FuncType{
				Params: &ast.FieldList{List: []*ast.Field{
					{Names: []*ast.Ident{ast.NewIdent("data")}, Type: &ast.ArrayType{Elt: ast.NewIdent("byte")}},
				}},
				Results: &ast.FieldList{List: []*ast.Field{
					{Type: ast.NewIdent("string")},
				}},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					// key := (o_k_0 + o_k_1) ^ o_k_2
					&ast.AssignStmt{
						Lhs: []ast.Expr{ast.NewIdent("key")},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{&ast.CallExpr{
							Fun:  ast.NewIdent("byte"),
							Args: []ast.Expr{&ast.ParenExpr{X: formula}},
						}},
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
								Rhs: []ast.Expr{&ast.BinaryExpr{X: ast.NewIdent("b"), Op: token.XOR, Y: ast.NewIdent("key")}},
							},
						}},
					},
					// return string(decrypted)
					&ast.ReturnStmt{Results: []ast.Expr{
						&ast.CallExpr{Fun: ast.NewIdent("string"), Args: []ast.Expr{ast.NewIdent("decrypted")}},
					}},
				},
			},
		}
		
		// Вставляем глобальные переменные и функцию в начало файла (после импортов).
		declsToInsert := append(keyPartDecls, decryptFunc)
		file.Decls = append(file.Decls[:len(file.Imports)], append(declsToInsert, file.Decls[len(file.Imports):]...)...)
	}

	return nil
}
