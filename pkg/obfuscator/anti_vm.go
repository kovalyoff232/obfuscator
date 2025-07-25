package obfuscator

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"
)

// AntiVMPass injects checks to detect if the code is running inside a virtual machine.
type AntiVMPass struct{}

func (p *AntiVMPass) Apply(fset *token.FileSet, file *ast.File) error {
	if file.Name.Name != "main" {
		return nil
	}

	// This is a weak check to prevent multiple injections. A proper implementation
	// would require a more robust mechanism to track applied passes.
	for _, decl := range file.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok && len(f.Body.List) > 2 {
			if as, ok := f.Body.List[2].(*ast.RangeStmt); ok {
				if _, ok := as.X.(*ast.Ident); ok {
					// Likely our function, let's skip. This is very heuristic.
					return nil
				}
			}
		}
	}

	astutil.AddImport(fset, file, "net")
	astutil.AddImport(fset, file, "os")
	astutil.AddImport(fset, file, "strings")

	checkFuncName := NewName()
	// 1. Create the MAC address check function
	checkFunc := &ast.FuncDecl{
		Name: ast.NewIdent(checkFuncName),
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				// interfaces, err := net.Interfaces()
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("interfaces"), ast.NewIdent("err")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.CallExpr{
						Fun: &ast.SelectorExpr{X: ast.NewIdent("net"), Sel: ast.NewIdent("Interfaces")},
					}},
				},
				// if err != nil { return }
				&ast.IfStmt{
					Cond: &ast.BinaryExpr{X: ast.NewIdent("err"), Op: token.NEQ, Y: ast.NewIdent("nil")},
					Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{}}},
				},
				// var vmPrefixes = []string{"00:05:69", "00:0c:29", "00:1c:14", "00:50:56", "08:00:27"}
				&ast.DeclStmt{
					Decl: &ast.GenDecl{
						Tok: token.VAR,
						Specs: []ast.Spec{
							&ast.ValueSpec{
								Names: []*ast.Ident{ast.NewIdent("vmPrefixes")},
								Type:  &ast.ArrayType{Elt: ast.NewIdent("string")},
								Values: []ast.Expr{
									&ast.CompositeLit{
										Type: &ast.ArrayType{Elt: ast.NewIdent("string")},
										Elts: []ast.Expr{
											&ast.BasicLit{Kind: token.STRING, Value: `"00:05:69"`}, // VMware
											&ast.BasicLit{Kind: token.STRING, Value: `"00:0c:29"`}, // VMware
											&ast.BasicLit{Kind: token.STRING, Value: `"00:1c:14"`}, // VMware
											&ast.BasicLit{Kind: token.STRING, Value: `"00:50:56"`}, // VMware
											&ast.BasicLit{Kind: token.STRING, Value: `"08:00:27"`}, // VirtualBox
										},
									},
								},
							},
						},
					},
				},
				// for _, i := range interfaces { ... }
				&ast.RangeStmt{
					Key:   ast.NewIdent("_"),
					Value: ast.NewIdent("i"),
					Tok:   token.DEFINE,
					X:     ast.NewIdent("interfaces"),
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							// for _, prefix := range vmPrefixes { ... }
							&ast.RangeStmt{
								Key:   ast.NewIdent("_"),
								Value: ast.NewIdent("prefix"),
								Tok:   token.DEFINE,
								X:     ast.NewIdent("vmPrefixes"),
								Body: &ast.BlockStmt{
									List: []ast.Stmt{
										// if strings.HasPrefix(i.HardwareAddr.String(), prefix) { os.Exit(1) }
										&ast.IfStmt{
											Cond: &ast.CallExpr{
												Fun: &ast.SelectorExpr{X: ast.NewIdent("strings"), Sel: ast.NewIdent("HasPrefix")},
												Args: []ast.Expr{
													&ast.CallExpr{
														Fun: &ast.SelectorExpr{
															X:   &ast.SelectorExpr{X: ast.NewIdent("i"), Sel: ast.NewIdent("HardwareAddr")},
															Sel: ast.NewIdent("String"),
														},
													},
													ast.NewIdent("prefix"),
												},
											},
											Body: &ast.BlockStmt{
												List: []ast.Stmt{
													&ast.ExprStmt{
														X: &ast.CallExpr{
															Fun:  &ast.SelectorExpr{X: ast.NewIdent("os"), Sel: ast.NewIdent("Exit")},
															Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// 2. Create or find an init function to call the check.
	var initFunc *ast.FuncDecl
	for _, decl := range file.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok && f.Name.Name == "init" {
			initFunc = f
			break
		}
	}

	// If no init function exists, create one.
	if initFunc == nil {
		initFunc = &ast.FuncDecl{
			Name: ast.NewIdent("init"),
			Type: &ast.FuncType{Params: &ast.FieldList{}},
			Body: &ast.BlockStmt{},
		}
		// Add the new init function to the file's declarations
		file.Decls = append(file.Decls, initFunc)
	}

	// 3. Add the call to the check function at the beginning of the init function's body.
	callStmt := &ast.ExprStmt{X: &ast.CallExpr{Fun: ast.NewIdent(checkFuncName)}}
	initFunc.Body.List = append([]ast.Stmt{callStmt}, initFunc.Body.List...)

	// 4. Add the check function itself to the file's declarations.
	file.Decls = append(file.Decls, checkFunc)

	return nil
}
