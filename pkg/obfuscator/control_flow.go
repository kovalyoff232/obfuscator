package obfuscator

import (
	"go/ast"
	"go/token"
	"math/rand"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

func ObfuscateControlFlow(node *ast.File) {
	astutil.Apply(node, func(cursor *astutil.Cursor) bool {
		block, ok := cursor.Node().(*ast.BlockStmt)
		if !ok || len(block.List) < 1 {
			return true
		}

		if parent := cursor.Parent(); parent != nil {
			if _, ok := parent.(*ast.CaseClause); ok {
				return true
			}
			if sw, ok := parent.(*ast.SwitchStmt); ok && sw.Body == block {
				return true
			}
		}

		if rand.Intn(100) < 30 { // 30% chance
			newStmts := createOpaqueSwitch(block.List)
			block.List = newStmts
		}

		return true
	}, nil)
}

func createOpaqueSwitch(stmts []ast.Stmt) []ast.Stmt {
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
					Body: stmts,
				},
			},
		},
	}

	return []ast.Stmt{initCtrlVar, switchStmt}
}

