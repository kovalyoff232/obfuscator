package obfuscator

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"
)

// AntiVMPass injects checks to detect if the program is running inside a virtual machine.
// The result of the check is stored in a global variable, which can be used by other
// passes to alter the program's behavior (e.g., corrupting a decryption key).
type AntiVMPass struct{}

// Apply injects the anti-vm logic into the main package.
func (p *AntiVMPass) Apply(fset *token.FileSet, file *ast.File) error {
	// This pass should only run on the main package.
	if file.Name.Name != "main" {
		return nil
	}

	// Check if we have already run this pass.
	if isVarDeclared(file, vmCheckVarName) {
		return nil
	}

	astutil.AddImport(fset, file, "net")
	astutil.AddImport(fset, file, "strings")
	astutil.AddImport(fset, file, "io/ioutil")
	astutil.AddImport(fset, file, "path/filepath")
	astutil.AddImport(fset, file, "runtime")
	astutil.AddImport(fset, file, "io/ioutil")
	astutil.AddImport(fset, file, "path/filepath")
	astutil.AddImport(fset, file, "runtime")

	// 1. Declare the global variable for the VM check result.
	vmVarDecl := &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names:  []*ast.Ident{ast.NewIdent(vmCheckVarName)},
				Type:   ast.NewIdent("int"),
				Values: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
			},
		},
	}

	// 2. Create the init function that performs the check.
	initFunc := createVMCheckInitFunc()

	// 3. Insert the new declarations after the last import.
	insertDeclsAfterImports(file, []ast.Decl{vmVarDecl, initFunc})

	return nil
}

// A randomly generated, unique name for the variable holding the VM check result.
var vmCheckVarName = NewName()

// createVMCheckInitFunc generates the AST for an init function that checks for VM indicators.
func createVMCheckInitFunc() *ast.FuncDecl {
	// --- Check 1: MAC Address Prefixes ---
	vmPrefixes := []string{
		"00:05:69", "00:0c:29", "00:50:56", // VMware
		"08:00:27", // VirtualBox
		"00:1c:42", // Parallels
		"00:16:3e", // Xen
	}
	var prefixLits []ast.Expr
	for _, prefix := range vmPrefixes {
		prefixLits = append(prefixLits, &ast.BasicLit{Kind: token.STRING, Value: "\"" + prefix + "\""})
	}

	// --- Check 2: DMI/SMBIOS Information (Linux-specific) ---
	dmiFiles := []string{"product_name", "sys_vendor", "board_vendor", "board_name"}
	vmStrings := []string{"VMware", "VirtualBox", "QEMU", "KVM", "Xen"}
	var dmiFileLits, vmStringLits []ast.Expr
	for _, f := range dmiFiles {
		dmiFileLits = append(dmiFileLits, &ast.BasicLit{Kind: token.STRING, Value: "\"" + f + "\""})
	}
	for _, s := range vmStrings {
		vmStringLits = append(vmStringLits, &ast.BasicLit{Kind: token.STRING, Value: "\"" + s + "\""})
	}

	doneLabel := ast.NewIdent("done")

	// --- Function Body Construction ---
	initBody := &ast.BlockStmt{
		List: []ast.Stmt{
			// --- MAC Check Logic ---
			&ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent("pfx")},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.CompositeLit{Type: &ast.ArrayType{Elt: ast.NewIdent("string")}, Elts: prefixLits}},
			},
			&ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent("ifs"), ast.NewIdent("err")},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.SelectorExpr{X: ast.NewIdent("net"), Sel: ast.NewIdent("Interfaces")},}},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{X: ast.NewIdent("err"), Op: token.EQL, Y: ast.NewIdent("nil")},
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.RangeStmt{
						Key:   ast.NewIdent("_"),
						Value: ast.NewIdent("i"),
						Tok:   token.DEFINE,
						X:     ast.NewIdent("ifs"),
						Body: &ast.BlockStmt{List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{ast.NewIdent("mac")},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.SelectorExpr{X: &ast.SelectorExpr{X: ast.NewIdent("i"), Sel: ast.NewIdent("HardwareAddr")}, Sel: ast.NewIdent("String")},}},
							},
							&ast.RangeStmt{
								Key:   ast.NewIdent("_"),
								Value: ast.NewIdent("p"),
								Tok:   token.DEFINE,
								X:     ast.NewIdent("pfx"),
								Body: &ast.BlockStmt{List: []ast.Stmt{
									&ast.IfStmt{
										Cond: &ast.CallExpr{
											Fun:  &ast.SelectorExpr{X: ast.NewIdent("strings"), Sel: ast.NewIdent("HasPrefix")},
											Args: []ast.Expr{ast.NewIdent("mac"), ast.NewIdent("p")},
										},
										Body: &ast.BlockStmt{List: []ast.Stmt{
											&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent(vmCheckVarName)}, Tok: token.ASSIGN, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}}},
											&ast.BranchStmt{Tok: token.GOTO, Label: doneLabel},
										}},
									},
								}},
							},
						}},
					},
				}},
			},
			// --- DMI Check Logic (Linux only) ---
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{X: &ast.SelectorExpr{X: ast.NewIdent("runtime"), Sel: ast.NewIdent("GOOS")}, Op: token.EQL, Y: &ast.BasicLit{Kind: token.STRING, Value: "\"linux\""}},
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent("dmiPath")}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"/sys/class/dmi/id/"`}}},
					&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent("dmiFiles")}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.CompositeLit{Type: &ast.ArrayType{Elt: ast.NewIdent("string")}, Elts: dmiFileLits}}},
					&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent("vmStrings")}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.CompositeLit{Type: &ast.ArrayType{Elt: ast.NewIdent("string")}, Elts: vmStringLits}}},
					&ast.RangeStmt{
						Key:   ast.NewIdent("_"),
						Value: ast.NewIdent("f"),
						Tok:   token.DEFINE,
						X:     ast.NewIdent("dmiFiles"),
						Body: &ast.BlockStmt{List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{ast.NewIdent("content"), ast.NewIdent("_")},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{&ast.CallExpr{
									Fun: &ast.SelectorExpr{X: ast.NewIdent("ioutil"), Sel: ast.NewIdent("ReadFile")},
									Args: []ast.Expr{&ast.CallExpr{
										Fun:  &ast.SelectorExpr{X: ast.NewIdent("filepath"), Sel: ast.NewIdent("Join")},
										Args: []ast.Expr{ast.NewIdent("dmiPath"), ast.NewIdent("f")},
									}},
								}},
							},
							&ast.RangeStmt{
								Key:   ast.NewIdent("_"),
								Value: ast.NewIdent("s"),
								Tok:   token.DEFINE,
								X:     ast.NewIdent("vmStrings"),
								Body: &ast.BlockStmt{List: []ast.Stmt{
									&ast.IfStmt{
										Cond: &ast.CallExpr{
											Fun:  &ast.SelectorExpr{X: ast.NewIdent("strings"), Sel: ast.NewIdent("Contains")},
											Args: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent("string"), Args: []ast.Expr{ast.NewIdent("content")}} , ast.NewIdent("s")},
										},
										Body: &ast.BlockStmt{List: []ast.Stmt{
											&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent(vmCheckVarName)}, Tok: token.ASSIGN, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}}},
											&ast.BranchStmt{Tok: token.GOTO, Label: doneLabel},
										}},
									},
								}},
							},
						}},
					},
				}},
			},
			&ast.LabeledStmt{
				Label: doneLabel,
				Stmt:  &ast.EmptyStmt{},
			},
		},
	}

	return &ast.FuncDecl{
		Name: ast.NewIdent("init"),
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: initBody,
	}
}