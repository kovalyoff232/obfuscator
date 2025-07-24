package obfuscator

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"math/rand"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

// ControlFlow flattens the control flow of function bodies.
func ControlFlow(f *ast.File, info *types.Info) {
	astutil.Apply(f, func(cursor *astutil.Cursor) bool {
		funcDecl, ok := cursor.Node().(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil || len(funcDecl.Body.List) == 0 {
			return true
		}

		if funcDecl.Name.Name == "main" || funcDecl.Name.Name == "init" {
			return true
		}

		newBody, err := flattenFunctionBody(funcDecl, info)
		if err != nil {
			return true
		}

		funcDecl.Body = newBody
		fmt.Printf("    - Flattened control flow for function %s\n", funcDecl.Name.Name)
		return false
	}, nil)
}

type BasicBlock struct {
	ID    int
	Stmts []ast.Stmt
}

func flattenFunctionBody(fn *ast.FuncDecl, info *types.Info) (*ast.BlockStmt, error) {
	if len(fn.Body.List) <= 1 {
		return nil, fmt.Errorf("not enough statements to flatten")
	}

	// 1. Hoist all variable declarations to the top of the function.
	hoistedVars, hoistedDecls := hoistVariables(fn.Body, info)

	// 2. Decompose the function body into basic blocks.
	blocks := decomposeToBasicBlocks(fn.Body.List)
	if len(blocks) <= 1 {
		return nil, fmt.Errorf("not enough blocks to flatten")
	}

	stateVar := ast.NewIdent("o_state_" + newName(""))
	exitState := len(blocks)
	var returnVars []*ast.Ident
	var returnStmts []ast.Stmt

	if fn.Type.Results != nil {
		for i, field := range fn.Type.Results.List {
			retVar := ast.NewIdent("o_ret_" + strconv.Itoa(i))
			returnVars = append(returnVars, retVar)
			returnStmts = append(returnStmts, &ast.DeclStmt{
				Decl: &ast.GenDecl{Tok: token.VAR, Specs: []ast.Spec{&ast.ValueSpec{Names: []*ast.Ident{retVar}, Type: field.Type}}},
			})
		}
	}

	// 3. Rewrite blocks.
	for i := range blocks {
		nextState := i + 1
		if i == len(blocks)-1 {
			nextState = exitState
		}
		blocks[i].Stmts = rewriteBlock(blocks[i].Stmts, stateVar, nextState, exitState, returnVars, hoistedVars)
	}

	rand.Shuffle(len(blocks), func(i, j int) { blocks[i], blocks[j] = blocks[j], blocks[i] })
	var cases []ast.Stmt
	for _, block := range blocks {
		cases = append(cases, &ast.CaseClause{
			List: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(block.ID)}},
			Body: block.Stmts,
		})
	}

	// 4. Assemble the new body.
	newBody := &ast.BlockStmt{}
	newBody.List = append(newBody.List, hoistedDecls...)
	newBody.List = append(newBody.List, returnStmts...)
	newBody.List = append(newBody.List, &ast.DeclStmt{
		Decl: &ast.GenDecl{Tok: token.VAR, Specs: []ast.Spec{
			&ast.ValueSpec{Names: []*ast.Ident{stateVar}, Values: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}},
		}},
	})
	newBody.List = append(newBody.List, &ast.ForStmt{
		Body: &ast.BlockStmt{List: []ast.Stmt{
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{X: stateVar, Op: token.EQL, Y: &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(exitState)}},
				Body: &ast.BlockStmt{List: []ast.Stmt{&ast.BranchStmt{Tok: token.BREAK}}},
			},
			&ast.SwitchStmt{Tag: stateVar, Body: &ast.BlockStmt{List: cases}},
		}},
	})
	if len(returnVars) > 0 {
		newBody.List = append(newBody.List, &ast.ReturnStmt{Results: Deref(returnVars)})
	}

	return newBody, nil
}

func hoistVariables(body *ast.BlockStmt, info *types.Info) (map[string]bool, []ast.Stmt) {
	vars := make(map[string]bool)
	var decls []ast.Stmt
	ast.Inspect(body, func(n ast.Node) bool {
		if assign, ok := n.(*ast.AssignStmt); ok && assign.Tok == token.DEFINE {
			for _, lhs := range assign.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					if !vars[ident.Name] {
						vars[ident.Name] = true
						var varType ast.Expr
						if info != nil && info.TypeOf(ident) != nil {
							varType = ast.NewIdent(info.TypeOf(ident).String())
						} else {
							varType = ast.NewIdent("interface{}")
						}
						decls = append(decls, &ast.DeclStmt{
							Decl: &ast.GenDecl{Tok: token.VAR, Specs: []ast.Spec{
								&ast.ValueSpec{
									Names: []*ast.Ident{ast.NewIdent(ident.Name)},
									Type:  varType,
								},
							}},
						})
					}
				}
			}
		}
		return true
	})
	return vars, decls
}

func decomposeToBasicBlocks(stmts []ast.Stmt) []BasicBlock {
	var blocks []BasicBlock
	currentBlock := BasicBlock{ID: 0}
	for _, stmt := range stmts {
		currentBlock.Stmts = append(currentBlock.Stmts, stmt)
		if isBlockTerminal(stmt) {
			blocks = append(blocks, currentBlock)
			currentBlock = BasicBlock{ID: len(blocks)}
		}
	}
	if len(currentBlock.Stmts) > 0 {
		blocks = append(blocks, currentBlock)
	}
	return blocks
}

func isBlockTerminal(stmt ast.Stmt) bool {
	switch stmt.(type) {
	case *ast.ReturnStmt, *ast.IfStmt, *ast.ForStmt, *ast.SwitchStmt:
		return true
	default:
		return false
	}
}

func rewriteBlock(stmts []ast.Stmt, stateVar *ast.Ident, nextState, exitState int, returnVars []*ast.Ident, hoistedVars map[string]bool) []ast.Stmt {
	// Convert `:=` to `=` for hoisted variables.
	ast.Inspect(&ast.BlockStmt{List: stmts}, func(n ast.Node) bool {
		if assign, ok := n.(*ast.AssignStmt); ok && assign.Tok == token.DEFINE {
			isHoisted := true
			for _, lhs := range assign.Lhs {
				if ident, ok := lhs.(*ast.Ident); !ok || !hoistedVars[ident.Name] {
					isHoisted = false
					break
				}
			}
			if isHoisted {
				assign.Tok = token.ASSIGN
			}
		}
		return true
	})

	lastStmt := stmts[len(stmts)-1]
	switch s := lastStmt.(type) {
	case *ast.ReturnStmt:
		var assignments []ast.Stmt
		for i, res := range s.Results {
			assignments = append(assignments, &ast.AssignStmt{Lhs: []ast.Expr{returnVars[i]}, Tok: token.ASSIGN, Rhs: []ast.Expr{res}})
		}
		assignments = append(assignments, &ast.AssignStmt{Lhs: []ast.Expr{stateVar}, Tok: token.ASSIGN, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(exitState)}}})
		return append(stmts[:len(stmts)-1], assignments...)
	default:
		return append(stmts, &ast.AssignStmt{Lhs: []ast.Expr{stateVar}, Tok: token.ASSIGN, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(nextState)}}})
	}
}

func Deref(vars []*ast.Ident) []ast.Expr {
	exprs := make([]ast.Expr, len(vars))
	for i, v := range vars {
		exprs[i] = v
	}
	return exprs
}
