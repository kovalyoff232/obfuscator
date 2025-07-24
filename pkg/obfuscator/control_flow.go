package obfuscator

import (
	"fmt"
	"go/ast"
	"go/token"
	"math/rand"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

// ControlFlow flattens the control flow of function bodies using different techniques.
func ControlFlow(f *ast.File) {
	astutil.Apply(f, func(cursor *astutil.Cursor) bool {
		funcDecl, ok := cursor.Node().(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil || len(funcDecl.Body.List) < 3 { // Need at least 3 statements for meaningful flattening
			return true
		}

		// A more sophisticated analysis would be needed to handle returns correctly.
		// For now, we skip functions with returns to avoid "missing return" errors.
		if hasReturn(funcDecl.Body) {
			return true
		}

		var flattenedBody *ast.BlockStmt
		// Randomly choose a flattening technique
		if rand.Intn(2) == 0 {
			flattenedBody = flattenWithSwitch(funcDecl.Body)
		} else {
			flattenedBody = flattenWithGoto(funcDecl.Body)
		}

		funcDecl.Body = flattenedBody
		return false
	}, nil)
}

// flattenWithSwitch uses a for-loop and a switch statement.
func flattenWithSwitch(body *ast.BlockStmt) *ast.BlockStmt {
	ctrlVarName := "o_switch_flow_" + strconv.Itoa(rand.Intn(10000))
	ctrlVar := ast.NewIdent(ctrlVarName)
	numStmts := len(body.List)

	var cases []ast.Stmt
	for i, stmt := range body.List {
		nextState := i + 1
		if nextState >= numStmts {
			nextState = -1 // Exit condition
		}

		caseBody := []ast.Stmt{
			stmt,
			&ast.AssignStmt{
				Lhs: []ast.Expr{ctrlVar},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(nextState)}},
			},
		}

		clause := &ast.CaseClause{
			List: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(i)}},
			Body: caseBody,
		}
		cases = append(cases, clause)
	}

	// Shuffle the cases to make the control flow non-obvious.
	rand.Shuffle(len(cases), func(i, j int) {
		cases[i], cases[j] = cases[j], cases[i]
	})

	return &ast.BlockStmt{
		List: []ast.Stmt{
			// var o_switch_flow_XXX = 0
			&ast.DeclStmt{Decl: &ast.GenDecl{Tok: token.VAR, Specs: []ast.Spec{
				&ast.ValueSpec{Names: []*ast.Ident{ctrlVar}, Values: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}},
			}}},
			// for { ... }
			&ast.ForStmt{Body: &ast.BlockStmt{List: []ast.Stmt{
				// if o_switch_flow_XXX < 0 { break }
				&ast.IfStmt{
					Cond: &ast.BinaryExpr{X: ctrlVar, Op: token.LSS, Y: &ast.BasicLit{Kind: token.INT, Value: "0"}},
					Body: &ast.BlockStmt{List: []ast.Stmt{&ast.BranchStmt{Tok: token.BREAK}}},
				},
				// switch o_switch_flow_XXX { ... }
				&ast.SwitchStmt{Tag: ctrlVar, Body: &ast.BlockStmt{List: cases}},
			}}},
		},
	}
}

// flattenWithGoto uses labels and goto statements.
func flattenWithGoto(body *ast.BlockStmt) *ast.BlockStmt {
	ctrlVarName := "o_goto_flow_" + strconv.Itoa(rand.Intn(10000))
	ctrlVar := ast.NewIdent(ctrlVarName)
	numStmts := len(body.List)
	order := rand.Perm(numStmts) // Random execution order of blocks

	var newBodyStmts []ast.Stmt

	// Initial state setup
	// var o_goto_flow_XXX = START_STATE
	startState := order[0]
	initStmt := &ast.DeclStmt{Decl: &ast.GenDecl{Tok: token.VAR, Specs: []ast.Spec{
		&ast.ValueSpec{Names: []*ast.Ident{ctrlVar}, Values: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(startState)}}},
	}}}
	newBodyStmts = append(newBodyStmts, initStmt)

	// Dispatcher using if/goto
	// dispatch_loop:
	//   if ctrlVar == 0 { goto block_0 }
	//   ...
	dispatchLabel := &ast.LabeledStmt{Label: ast.NewIdent("dispatch_loop"), Stmt: &ast.EmptyStmt{}}
	newBodyStmts = append(newBodyStmts, dispatchLabel)

	for i := 0; i < numStmts; i++ {
		ifStmt := &ast.IfStmt{
			Cond: &ast.BinaryExpr{X: ctrlVar, Op: token.EQL, Y: &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(i)}},
			Body: &ast.BlockStmt{List: []ast.Stmt{&ast.BranchStmt{Tok: token.GOTO, Label: ast.NewIdent(fmt.Sprintf("block_%d", i))}}},
		}
		newBodyStmts = append(newBodyStmts, ifStmt)
	}
	// goto end_loop (if no condition met, which shouldn't happen)
	newBodyStmts = append(newBodyStmts, &ast.BranchStmt{Tok: token.GOTO, Label: ast.NewIdent("end_loop")})

	// Statement blocks
	// block_N:
	//   original_statement_N
	//   ctrlVar = next_state
	//   goto dispatch_loop
	nextStateMap := make(map[int]int)
	for i := 0; i < numStmts-1; i++ {
		nextStateMap[order[i]] = order[i+1]
	}
	nextStateMap[order[numStmts-1]] = -1 // Final state

	for i := 0; i < numStmts; i++ {
		blockStmts := []ast.Stmt{
			body.List[i],
			&ast.AssignStmt{
				Lhs: []ast.Expr{ctrlVar},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(nextStateMap[i])}},
			},
			&ast.BranchStmt{Tok: token.GOTO, Label: ast.NewIdent("dispatch_loop")},
		}
		newBodyStmts = append(newBodyStmts, &ast.LabeledStmt{
			Label: ast.NewIdent(fmt.Sprintf("block_%d", i)),
			Stmt:  &ast.BlockStmt{List: blockStmts},
		})
	}

	// End label
	newBodyStmts = append(newBodyStmts, &ast.LabeledStmt{Label: ast.NewIdent("end_loop"), Stmt: &ast.EmptyStmt{}})

	return &ast.BlockStmt{List: newBodyStmts}
}

// hasReturn checks if a block contains a return statement.
func hasReturn(body *ast.BlockStmt) bool {
	has := false
	ast.Inspect(body, func(n ast.Node) bool {
		if _, ok := n.(*ast.ReturnStmt); ok {
			has = true
		}
		// Stop searching if we found one
		return !has
	})
	return has
}
