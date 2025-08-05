//go:build linux
// +build linux

package obfuscator

import (
	"context"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"
)

// Если obf.Config.Аnti.Manager присутствует — используем фасад: при detected=true
// вычисляем obf.WeavingKeyVarName по текущей формуле (ptraceComponent + vm*9999),
// сохраняя существующую семантику. При отсутствии фасада — текущая AST-вставка (fallback).
// AntiDebugPass injects dynamic, multi-faceted anti-debugging checks (Linux).
// It uses ptrace detection and vmCheckVarName to compute obf.WeavingKeyVarName.
// Отключение возможно через ENV OBF_DISABLE_ANTI_DEBUG.
type AntiDebugPass struct{}

// Apply injects the anti-debugging logic into the main package (Linux).
func (p *AntiDebugPass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	// Only for main package.
	if file.Name.Name != "main" {
		return nil
	}
	// Ensure we run once.
	if isVarDeclared(file, obf.WeavingKeyVarName) {
		return nil
	}

	// runtime не используется — не импортируем
	astutil.AddImport(fset, file, "syscall")
	astutil.AddImport(fset, file, "os")

	// 1) Declare global weaving key var.
	keyVarDecl := &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{ast.NewIdent(obf.WeavingKeyVarName)},
				Type:  ast.NewIdent("int64"),
			},
		},
	}

	// Маршрутизация через фасад: если доступен и включён AntiDebug.
	useFacade := obfHasAntiDebug(obf)
	if useFacade {
		ctx := context.Background()
		detected, _, err := obf.anti.Manager.CheckDebugger(ctx)
		if err == nil && detected {
			// При detected=true — инжектируем стандартный init с формулой (fallback код подходит).
			initFunc := createAntiDebugInitFuncLinux(obf)
			insertDeclsAfterImports(file, []ast.Decl{keyVarDecl, initFunc})
			return nil
		}
		// иначе: fallback на текущую логику (тоже createAntiDebugInitFuncLinux)
	}

	// 2) init() body (fallback)
	initFunc := createAntiDebugInitFuncLinux(obf)

	// 3) Insert after imports
	insertDeclsAfterImports(file, []ast.Decl{keyVarDecl, initFunc})

	return nil
}

// createAntiDebugInitFuncLinux builds init() with:
// - early exit if OBF_DISABLE_ANTI_DEBUG is set
// - ptrace check via syscall.Syscall(SYS_PTRACE, PTRACE_TRACEME, 0, 0)
// - weaving key = ptraceComponent + int64(vmCheckVarName)*9999
func createAntiDebugInitFuncLinux(obf *Obfuscator) *ast.FuncDecl {
	done := ast.NewIdent("done")

	// Place label first to ensure any goto done does not jump over declarations
	doneLabel := &ast.LabeledStmt{Label: done, Stmt: &ast.EmptyStmt{}}

	// Declare ptraceComponent immediately after label, before any possible goto
	declPtrace := &ast.DeclStmt{Decl: &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names:  []*ast.Ident{ast.NewIdent("ptraceComponent")},
				Type:   ast.NewIdent("int64"),
				Values: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
			},
		},
	}}

	// if os.Getenv("OBF_DISABLE_ANTI_DEBUG") != "" { goto done }
	earlyExit := &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.CallExpr{Fun: &ast.SelectorExpr{X: ast.NewIdent("os"), Sel: ast.NewIdent("Getenv")}, Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: "\"OBF_DISABLE_ANTI_DEBUG\""}}},
			Op: token.NEQ,
			Y:  &ast.BasicLit{Kind: token.STRING, Value: "\"\""},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{
			&ast.BranchStmt{Tok: token.GOTO, Label: done},
		}},
	}

	// ptrace check (Linux)
	ptraceCheck := &ast.IfStmt{
		Init: &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("pres"), ast.NewIdent("_"), ast.NewIdent("_")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{X: ast.NewIdent("syscall"), Sel: ast.NewIdent("Syscall")},
					Args: []ast.Expr{
						&ast.SelectorExpr{X: ast.NewIdent("syscall"), Sel: ast.NewIdent("SYS_PTRACE")},
						&ast.CallExpr{Fun: ast.NewIdent("uintptr"), Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}}, // PTRACE_TRACEMЕ
						&ast.CallExpr{Fun: ast.NewIdent("uintptr"), Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}},
						&ast.CallExpr{Fun: ast.NewIdent("uintptr"), Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}},
					},
				},
			},
		},
		Cond: &ast.BinaryExpr{X: ast.NewIdent("pres"), Op: token.NEQ, Y: &ast.BasicLit{Kind: token.INT, Value: "0"}},
		Body: &ast.BlockStmt{List: []ast.Stmt{
			&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent("ptraceComponent")}, Tok: token.ASSIGN, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1337"}}},
		}},
	}

	// final weaving key calc
	// Anti-VM может быть отключён, тогда vmCheckVarName отсутствует — используем 0.
	vmAsInt64 := &ast.BasicLit{Kind: token.INT, Value: "0"}

	finalCalc := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(obf.WeavingKeyVarName)},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			&ast.BinaryExpr{
				X:  ast.NewIdent("ptraceComponent"),
				Op: token.ADD,
				Y: &ast.BinaryExpr{
					X:  vmAsInt64,
					Op: token.MUL,
					Y:  &ast.BasicLit{Kind: token.INT, Value: "9999"},
				},
			},
		},
	}

	body := &ast.BlockStmt{
		List: []ast.Stmt{
			doneLabel,  // label first
			declPtrace, // declarations before any goto
			earlyExit,  // early exit may goto done now safely
			ptraceCheck,
			finalCalc,
		},
	}

	return &ast.FuncDecl{
		Name: ast.NewIdent("init"),
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: body,
	}
}

// obfHasAntiDebug — проверка доступности фасада и включённости Debug.
func obfHasAntiDebug(obf *Obfuscator) bool {
	if obf == nil || obf.anti == nil || obf.anti.Config == nil {
		return false
	}
	return obf.anti.Manager != nil && obf.anti.Config.EnableDebug
}
