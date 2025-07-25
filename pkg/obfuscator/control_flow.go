package obfuscator

import (
	"crypto/rand"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"math/big"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

// ControlFlow flattens the control flow of function bodies using a switch-based dispatcher,
// enhanced with opaque predicates to create junk code paths.
func ControlFlow(f *ast.File, info *types.Info) {
	astutil.Apply(f, func(cursor *astutil.Cursor) bool {
		funcDecl, ok := cursor.Node().(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil || len(funcDecl.Body.List) == 0 {
			return true
		}

		if funcDecl.Name.Name == "main" || funcDecl.Name.Name == "init" || len(funcDecl.Body.List) < 3 {
			return true
		}

		newBody, err := flattenFunctionBody(funcDecl, info)
		if err != nil {
			return true
		}

		funcDecl.Body = newBody
		return false
	}, nil)
}

type BasicBlock struct {
	ID    int
	Stmts []ast.Stmt
}

type hoistedVar struct {
	OriginalName string
	NewName      string
	Type         ast.Expr
}

func flattenFunctionBody(fn *ast.FuncDecl, info *types.Info) (*ast.BlockStmt, error) {
	hoistedVars, hoistedDecls := hoistAndRenameVariables(fn.Body, info)
	blocks := decomposeToBasicBlocks(fn.Body.List)
	if len(blocks) <= 1 {
		return nil, fmt.Errorf("not enough blocks to flatten")
	}

	stateVar := ast.NewIdent(NewName())
	exitState := len(blocks)
	var returnVars []*ast.Ident
	var returnStmts []ast.Stmt

	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			retVar := ast.NewIdent(NewName())
			returnVars = append(returnVars, retVar)
			returnStmts = append(returnStmts, &ast.DeclStmt{
				Decl: &ast.GenDecl{Tok: token.VAR, Specs: []ast.Spec{&ast.ValueSpec{Names: []*ast.Ident{retVar}, Type: field.Type}}},
			})
		}
	}

	for i := range blocks {
		nextState := i + 1
		if i == len(blocks)-1 {
			nextState = exitState
		}
		blocks[i].Stmts = rewriteBlock(blocks[i].Stmts, stateVar, nextState, exitState, returnVars, hoistedVars)
	}

	junkCases := createJunkCases(len(blocks), len(blocks)+5)

	// Fisher-Yates shuffle using crypto/rand
	for i := len(blocks) - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		blocks[i], blocks[j.Int64()] = blocks[j.Int64()], blocks[i]
	}

	var cases []ast.Stmt
	for _, block := range blocks {
		cases = append(cases, &ast.CaseClause{
			List: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(block.ID)}},
			Body: block.Stmts,
		})
	}
	cases = append(cases, junkCases...)

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

func createJunkCases(startID, count int) []ast.Stmt {
	var junkCases []ast.Stmt
	for i := 0; i < count; i++ {
		junkID := startID + i
		x, y := NewName(), NewName()
		var cond ast.Expr

		template := randInt(5) // Increased number of templates
		switch template {
		case 0:
			// (x*x + 1) < 0 -- always false
			cond = &ast.BinaryExpr{
				X: &ast.BinaryExpr{
					X:  &ast.BinaryExpr{X: ast.NewIdent(x), Op: token.MUL, Y: ast.NewIdent(x)},
					Op: token.ADD,
					Y:  &ast.BasicLit{Kind: token.INT, Value: "1"},
				},
				Op: token.LSS,
				Y:  &ast.BasicLit{Kind: token.INT, Value: "0"},
			}
		case 1:
			// (x - y) * (x + y) != x*x - y*y -- always false
			cond = &ast.BinaryExpr{
				X: &ast.BinaryExpr{
					X:  &ast.ParenExpr{X: &ast.BinaryExpr{X: ast.NewIdent(x), Op: token.SUB, Y: ast.NewIdent(y)}},
					Op: token.MUL,
					Y:  &ast.ParenExpr{X: &ast.BinaryExpr{X: ast.NewIdent(x), Op: token.ADD, Y: ast.NewIdent(y)}},
				},
				Op: token.NEQ,
				Y: &ast.BinaryExpr{
					X:  &ast.BinaryExpr{X: ast.NewIdent(x), Op: token.MUL, Y: ast.NewIdent(x)},
					Op: token.SUB,
					Y:  &ast.BinaryExpr{X: ast.NewIdent(y), Op: token.MUL, Y: ast.NewIdent(y)},
				},
			}
		case 2:
			// (x ^ y) ^ y != x -- always false
			cond = &ast.BinaryExpr{
				X: &ast.ParenExpr{X: &ast.BinaryExpr{
					X:  &ast.ParenExpr{X: &ast.BinaryExpr{X: ast.NewIdent(x), Op: token.XOR, Y: ast.NewIdent(y)}},
					Op: token.XOR,
					Y:  ast.NewIdent(y),
				}},
				Op: token.NEQ,
				Y:  ast.NewIdent(x),
			}
		case 3:
			// 7*x - 3*x - 4*x != 0 -- always false
			cond = &ast.BinaryExpr{
				X: &ast.BinaryExpr{
					X: &ast.BinaryExpr{
						X:  &ast.BinaryExpr{X: &ast.BasicLit{Kind: token.INT, Value: "7"}, Op: token.MUL, Y: ast.NewIdent(x)},
						Op: token.SUB,
						Y:  &ast.BinaryExpr{X: &ast.BasicLit{Kind: token.INT, Value: "3"}, Op: token.MUL, Y: ast.NewIdent(x)},
					},
					Op: token.SUB,
					Y:  &ast.BinaryExpr{X: &ast.BasicLit{Kind: token.INT, Value: "4"}, Op: token.MUL, Y: ast.NewIdent(x)},
				},
				Op: token.NEQ,
				Y:  &ast.BasicLit{Kind: token.INT, Value: "0"},
			}
		default:
			// x*y + 1 == x*y -- always false
			cond = &ast.BinaryExpr{
				X: &ast.BinaryExpr{
					X:  &ast.BinaryExpr{X: ast.NewIdent(x), Op: token.MUL, Y: ast.NewIdent(y)},
					Op: token.ADD,
					Y:  &ast.BasicLit{Kind: token.INT, Value: "1"},
				},
				Op: token.EQL,
				Y:  &ast.BinaryExpr{X: ast.NewIdent(x), Op: token.MUL, Y: ast.NewIdent(y)},
			}
		}

		junkCases = append(junkCases, &ast.CaseClause{
			List: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(junkID)}},
			Body: []ast.Stmt{
				&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent(x)}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "123"}}},
				&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent(y)}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "456"}}},
				&ast.IfStmt{
					Cond: cond,
					Body: &ast.BlockStmt{List: []ast.Stmt{
						&ast.ExprStmt{X: &ast.CallExpr{Fun: ast.NewIdent("panic"), Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: "\"unreachable\""}}}},
					}},
				},
				// Add this line to "use" the y variable
				&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent("_")}, Tok: token.ASSIGN, Rhs: []ast.Expr{ast.NewIdent(y)}},
			},
		})
	}
	return junkCases
}

func hoistAndRenameVariables(body *ast.BlockStmt, info *types.Info) (map[string]*hoistedVar, []ast.Stmt) {
	vars := make(map[string]*hoistedVar)
	var decls []ast.Stmt

	registerVar := func(ident *ast.Ident) {
		if _, exists := vars[ident.Name]; !exists {
			var varType ast.Expr
			if info != nil && info.TypeOf(ident) != nil {
				typeString := info.TypeOf(ident).String()
				if _, err := parser.ParseExpr(typeString); err == nil {
					varType = ast.NewIdent(typeString)
				} else {
					varType = ast.NewIdent("interface{}")
				}
			} else {
				varType = ast.NewIdent("interface{}")
			}

			newVar := &hoistedVar{
				OriginalName: ident.Name,
				NewName:      NewName(),
				Type:         varType,
			}
			vars[ident.Name] = newVar
			decls = append(decls, &ast.DeclStmt{
				Decl: &ast.GenDecl{Tok: token.VAR, Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{ast.NewIdent(newVar.NewName)},
						Type:  newVar.Type,
					},
				}},
			})
		}
	}

	astutil.Apply(body, func(cursor *astutil.Cursor) bool {
		switch n := cursor.Node().(type) {
		case *ast.GenDecl:
			if n.Tok == token.VAR {
				var assignments []ast.Stmt
				for _, spec := range n.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range vs.Names {
							registerVar(name)
						}
						if len(vs.Values) > 0 {
							var lhs []ast.Expr
							for _, name := range vs.Names {
								lhs = append(lhs, name)
							}
							assignments = append(assignments, &ast.AssignStmt{
								Lhs: lhs,
								Tok: token.ASSIGN,
								Rhs: vs.Values,
							})
						}
					}
				}
				if len(assignments) > 0 {
					if len(assignments) == 1 {
						cursor.Replace(assignments[0])
					}
				} else {
					cursor.Delete()
				}
			}
		case *ast.AssignStmt:
			if n.Tok == token.DEFINE {
				for _, lhs := range n.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok {
						registerVar(ident)
					}
				}
				n.Tok = token.ASSIGN
			}
		}
		return true
	}, nil)

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
	case *ast.ReturnStmt, *ast.IfStmt, *ast.ForStmt, *ast.SwitchStmt, *ast.RangeStmt:
		return true
	default:
		return false
	}
}

func rewriteBlock(stmts []ast.Stmt, stateVar *ast.Ident, nextState, exitState int, returnVars []*ast.Ident, hoistedVars map[string]*hoistedVar) []ast.Stmt {
	astutil.Apply(&ast.BlockStmt{List: stmts}, func(cursor *astutil.Cursor) bool {
		ident, ok := cursor.Node().(*ast.Ident)
		if ok {
			if hv, exists := hoistedVars[ident.Name]; exists {
				ident.Name = hv.NewName
			}
		}
		return true
	}, nil)

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
