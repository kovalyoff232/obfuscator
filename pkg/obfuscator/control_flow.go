package obfuscator

import (
	"go/ast"
	"go/token"
	"math/rand"
	"strconv"
)

// ObfuscateControlFlow обходит AST и заменяет блоки кода на `switch`.
func ObfuscateControlFlow(node *ast.File) {
	ast.Inspect(node, func(n ast.Node) bool {
		block, ok := n.(*ast.BlockStmt)
		if !ok || len(block.List) < 2 {
			return true
		}

		if rand.Intn(100) < 50 { // Увеличим шанс для тестирования
			newBlock := hoistVarsAndCreateOpaqueSwitch(block)
			*block = *newBlock
		}

		return true
	})
}

func hoistVarsAndCreateOpaqueSwitch(block *ast.BlockStmt) *ast.BlockStmt {
	varsToHoist := make(map[string]bool)
	var newBodyStmts []ast.Stmt

	// 1. Собираем все имена переменных, которые объявляются через `:=`.
	ast.Inspect(block, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok || assign.Tok != token.DEFINE {
			return true
		}
		for _, lhs := range assign.Lhs {
			if ident, ok := lhs.(*ast.Ident); ok {
				varsToHoist[ident.Name] = true
			}
		}
		return true
	})

	if len(varsToHoist) == 0 {
		return block // Нечего "поднимать".
	}

	// 2. Создаем объявления `var` для всех найденных переменных.
	var varSpecs []ast.Spec
	for varName := range varsToHoist {
		varSpecs = append(varSpecs, &ast.ValueSpec{
			Names: []*ast.Ident{ast.NewIdent(varName)},
			Type:  &ast.InterfaceType{Methods: &ast.FieldList{}},
		})
	}
	varDecl := &ast.DeclStmt{Decl: &ast.GenDecl{Tok: token.VAR, Specs: varSpecs}}

	// 3. Заменяем все `:=` на `=` в блоке.
	for _, stmt := range block.List {
		if assign, ok := stmt.(*ast.AssignStmt); ok && assign.Tok == token.DEFINE {
			assign.Tok = token.ASSIGN
		}
		newBodyStmts = append(newBodyStmts, stmt)
	}

	// 4. Создаем `switch`.
	ctrlVarName := "o_ctrl_" + strconv.Itoa(rand.Intn(1000))
	initCtrlVar := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(ctrlVarName)},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
	}
	switchStmt := &ast.SwitchStmt{
		Tag: ast.NewIdent(ctrlVarName),
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.CaseClause{
					List: nil, // default
					Body: newBodyStmts,
				},
			},
		},
	}

	// 5. Собираем итоговый блок.
	finalStmts := []ast.Stmt{varDecl, initCtrlVar, switchStmt}
	return &ast.BlockStmt{List: finalStmts}
}
