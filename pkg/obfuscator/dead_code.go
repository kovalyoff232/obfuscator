package obfuscator

import (
	"fmt"
	"go/ast"
	"go/token"
	"math/rand"
	"strconv"
)

// InsertDeadCode обходит AST и вставляет "мертвый" код в тела функций.
func InsertDeadCode(node *ast.File) {
	ast.Inspect(node, func(n ast.Node) bool {
		block, ok := n.(*ast.BlockStmt)
		if !ok || len(block.List) == 0 {
			return true
		}

		// Проверяем, есть ли в блоке `return`.
		hasReturn := false
		for _, stmt := range block.List {
			if _, ok := stmt.(*ast.ReturnStmt); ok {
				hasReturn = true
				break
			}
		}

		// Если в блоке есть `return`, мы не можем безопасно добавить код в конец.
		// В реальном проекте потребовался бы более сложный анализ,
		// но для данного случая мы просто пропустим этот блок.
		if hasReturn {
			return true
		}

		if rand.Intn(100) < 30 { // 30% шанс
			block.List = append(block.List, createDeadIfStmt())
		}

		return true
	})
}

// createDeadIfStmt создает `if false { ... }` блок с мусорным кодом,
// который не вызывает ошибок "declared and not used".
func createDeadIfStmt() ast.Stmt {
	varName1 := fmt.Sprintf("o_dead_%d", rand.Intn(1000))
	varName2 := fmt.Sprintf("o_dead_%d", rand.Intn(1000)+1000)

	// Cоздаем `if false { ... }`
	return &ast.IfStmt{
		Cond: &ast.Ident{Name: "false"},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				// var o_dead_1 = 123
				&ast.DeclStmt{
					Decl: &ast.GenDecl{
						Tok: token.VAR,
						Specs: []ast.Spec{
							&ast.ValueSpec{
								Names:  []*ast.Ident{ast.NewIdent(varName1)},
								Values: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(rand.Intn(100))}},
							},
						},
					},
				},
				// o_dead_2 := o_dead_1 + 1
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(varName2)},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.BinaryExpr{
							X:  ast.NewIdent(varName1),
							Op: token.ADD,
							Y:  &ast.BasicLit{Kind: token.INT, Value: "1"},
						},
					},
				},
				// _ = o_dead_2 (используем переменную, чтобы компилятор был доволен)
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("_")},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{ast.NewIdent(varName2)},
				},
			},
		},
	}
}
