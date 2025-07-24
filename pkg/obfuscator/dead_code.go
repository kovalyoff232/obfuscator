package obfuscator

import (
	"fmt"
	"go/ast"
	"go/token"
	"math/rand"
	"strconv"
)

func InsertDeadCode(node *ast.File) {
	ast.Inspect(node, func(n ast.Node) bool {
		block, ok := n.(*ast.BlockStmt)
		if !ok || len(block.List) == 0 {
			return true
		}

		hasReturn := false
		for _, stmt := range block.List {
			if _, ok := stmt.(*ast.ReturnStmt); ok {
				hasReturn = true
				break
			}
		}

		if hasReturn {
			return true
		}

		if rand.Intn(100) < 30 { // 30% chance
			block.List = append(block.List, createDeadIfStmt())
		}

		return true
	})
}

func createDeadIfStmt() ast.Stmt {
	varName1 := fmt.Sprintf("o_dead_%d", rand.Intn(1000))
	varName2 := fmt.Sprintf("o_dead_%d", rand.Intn(1000)+1000)

	return &ast.IfStmt{
		Cond: &ast.Ident{Name: "false"},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
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
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("_")},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{ast.NewIdent(varName2)},
				},
			},
		},
	}
}
