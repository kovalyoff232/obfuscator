package obfuscator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math/rand"
	"strconv"
	"time"

	"golang.org/x/tools/go/ast/astutil"
)

const decryptFuncName = "o_d"

// Новый шаблон. Ключ вычисляется внутри из двух частей.
const decryptFuncTpl = `
func %s(data []byte) string {
	p1 := byte(%d)
	p2 := byte(%d)
	key := p1 ^ p2
	decrypted := make([]byte, len(data))
	for i, b := range data {
		decrypted[i] = b ^ key
	}
	return string(decrypted)
}
`

func xorEncrypt(data []byte, key byte) []byte {
	result := make([]byte, len(data))
	for i, b := range data {
		result[i] = b ^ key
	}
	return result
}

// EncryptStrings шифрует все строковые литералы в файле.
func EncryptStrings(file *ast.File) error {
	rand.Seed(time.Now().UnixNano())
	
	// Генерируем "рецепт" ключа
	part1 := byte(rand.Intn(256))
	part2 := byte(rand.Intn(256))
	key := part1 ^ part2
	
	hasEncryptedStrings := false

	// Ищем и заменяем строки
	astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		lit, ok := cursor.Node().(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		if _, ok := cursor.Parent().(*ast.ImportSpec); ok {
			return true // Не трогаем импорты
		}

		originalString, err := strconv.Unquote(lit.Value)
		if err != nil || originalString == "" {
			return true
		}

		hasEncryptedStrings = true
		encryptedBytes := xorEncrypt([]byte(originalString), key)

		// Если мы внутри const, меняем его на var
		path, _ := astutil.PathEnclosingInterval(file, lit.Pos(), lit.End())
		if path != nil {
			for _, pnode := range path {
				if genDecl, ok := pnode.(*ast.GenDecl); ok && genDecl.Tok == token.CONST {
					genDecl.Tok = token.VAR
					break
				}
			}
		}

		// Создаем вызов функции-дешифратора (теперь без ключа)
		byteSliceLit := &ast.CompositeLit{Type: &ast.ArrayType{Elt: ast.NewIdent("byte")}};
		for _, b := range encryptedBytes {
			byteSliceLit.Elts = append(byteSliceLit.Elts, &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(int(b))})
		}
		callExpr := &ast.CallExpr{
			Fun: ast.NewIdent(decryptFuncName),
			Args: []ast.Expr{byteSliceLit},
		}
		cursor.Replace(callExpr)
		return true
	}, nil)

	// Если нужно, внедряем функцию-дешифратор с "рецептом"
	if hasEncryptedStrings {
		fset := token.NewFileSet()
		// Вставляем части ключа в шаблон
		fullSrc := "package " + file.Name.Name + "\n" + fmt.Sprintf(decryptFuncTpl, decryptFuncName, part1, part2)
		decrypterFile, err := parser.ParseFile(fset, "", fullSrc, 0)
		if err != nil {
			return fmt.Errorf("не удалось распарсить шаблон дешифратора: %w", err)
		}
		decryptFunc := decrypterFile.Decls[0].(*ast.FuncDecl)
		file.Decls = append(file.Decls, decryptFunc)
	}

	return nil
}