package obfuscator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// Pass represents a syntax-only obfuscation pass that runs on a single file.
type Pass interface {
	Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error
}

// GlobalPass represents a syntax-only obfuscation pass that needs to run on all files at once.
type GlobalPass interface {
	Apply(obf *Obfuscator, fset *token.FileSet, files map[string]*ast.File) error
}

// TypeAwarePass represents a semantic obfuscation pass that requires type info for the whole package.
type TypeAwarePass interface {
	Apply(obf *Obfuscator, pkg *packages.Package) error
}

// --- Pass Implementations (stubs for type safety) ---

type renamePass struct{}

func (p *renamePass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	RenameIdentifiers(file)
	return nil
}

type deadCodePass struct{}

func (p *deadCodePass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	InsertDeadCode(file)
	return nil
}

type controlFlowPass struct{}

func (p *controlFlowPass) Apply(obf *Obfuscator, pkg *packages.Package) error {
	for _, file := range pkg.Syntax {
		ControlFlow(file, pkg.TypesInfo)
	}
	return nil
}

type expressionPass struct{}

func (p *expressionPass) Apply(obf *Obfuscator, pkg *packages.Package) error {
	for _, file := range pkg.Syntax {
		ObfuscateExpressions(file, pkg.TypesInfo)
	}
	return nil
}

type constantPass struct{}

func (p *constantPass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	ObfuscateConstants(file)
	return nil
}

type antiDebugPass struct{}

func (p *antiDebugPass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	pass := &AntiDebugPass{}
	return pass.Apply(obf, fset, file)
}

type antiVMPass struct{}

func (p *antiVMPass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	pass := &AntiVMPass{}
	return pass.Apply(fset, file)
}

// --- Configuration and Orchestration ---

type Config struct {
	RenameIdentifiers    bool
	EncryptStrings       bool
	InsertDeadCode       bool
	ObfuscateControlFlow bool
	ObfuscateExpressions bool
	ObfuscateConstants   bool
	ObfuscateDataFlow    bool
	AntiDebugging        bool
	AntiVM               bool
	IndirectCalls        bool
}

type Obfuscator struct {
	syntaxPasses      []Pass
	globalPasses      []GlobalPass
	typeAwarePasses   []TypeAwarePass
	WeavingKeyVarName string // Name of the global var for the anti-debug key

	// For advanced string encryption
	StringDecryptionFuncName string
	EncryptedStringsVarName  string
	StringKeysVarName        string
	stringEncryption         *StringEncryptionPass
}

func NewObfuscator(cfg *Config) *Obfuscator {
	obf := &Obfuscator{
		WeavingKeyVarName:        NewName(),
		StringDecryptionFuncName: NewName(),
		EncryptedStringsVarName:  NewName(),
		StringKeysVarName:        NewName(),
	}

	// --- Pass Ordering ---
	// 1. Anti-debugging must come first to establish the weaving key.
	if cfg.AntiDebugging {
		obf.syntaxPasses = append(obf.syntaxPasses, &antiDebugPass{})
	}

	// 2. String encryption uses the weaving key.
	if cfg.EncryptStrings {
		obf.stringEncryption = NewStringEncryptionPass()
		// We don't add it to the passes list directly, it's handled specially.
	}

	// 3. Other passes
	if cfg.AntiVM {
		obf.syntaxPasses = append(obf.syntaxPasses, &antiVMPass{})
	}
	if cfg.ObfuscateDataFlow {
		obf.typeAwarePasses = append(obf.typeAwarePasses, &DataFlowPass{})
	}
	// Rename should be one of the last syntax passes to avoid interfering with other passes.
	if cfg.RenameIdentifiers {
		obf.syntaxPasses = append(obf.syntaxPasses, &renamePass{})
	}
	if cfg.ObfuscateConstants {
		obf.syntaxPasses = append(obf.syntaxPasses, &constantPass{})
	}
	if cfg.ObfuscateExpressions {
		obf.typeAwarePasses = append(obf.typeAwarePasses, &expressionPass{})
	}
	if cfg.InsertDeadCode {
		obf.syntaxPasses = append(obf.syntaxPasses, &deadCodePass{})
	}
	if cfg.ObfuscateControlFlow {
		obf.typeAwarePasses = append(obf.typeAwarePasses, &controlFlowPass{})
	}
	if cfg.IndirectCalls {
		obf.globalPasses = append(obf.globalPasses, &CallIndirectionPass{})
	}

	return obf
}

func ProcessDirectory(inputPath, outputPath string, cfg *Config) error {
	if err := os.RemoveAll(outputPath); err != nil {
		return fmt.Errorf("failed to clean output directory: %w", err)
	}
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	obfuscator := NewObfuscator(cfg)
	fset := token.NewFileSet()

	loadCfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Fset:  fset,
		Dir:   inputPath,
	}
	pkgs, err := packages.Load(loadCfg, "./...")
	if err != nil {
		return fmt.Errorf("failed to load package: %w", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return fmt.Errorf("errors occurred while loading packages")
	}

	var mainPkg *packages.Package
	for _, pkg := range pkgs {
		if pkg.Name == "main" {
			mainPkg = pkg
		}
		fmt.Printf("Processing package: %s\n", pkg.PkgPath)

		// Run type-aware passes that operate on the whole package at once.
		for _, pass := range obfuscator.typeAwarePasses {
			if err := pass.Apply(obfuscator, pkg); err != nil {
				return fmt.Errorf("error in type-aware pass for package %s: %w", pkg.Name, err)
			}
		}

		// Run syntax-only passes on each file individually.
		for i, filePath := range pkg.GoFiles {
			fileNode := pkg.Syntax[i]
			fmt.Printf("  - File: %s\n", filePath)

			// Special handling for string encryption
			if obfuscator.stringEncryption != nil {
				if err := obfuscator.stringEncryption.Apply(obfuscator, fset, fileNode); err != nil {
					return fmt.Errorf("error in string encryption pass for file %s: %w", filePath, err)
				}
			}

			for _, pass := range obfuscator.syntaxPasses {
				if err := pass.Apply(obfuscator, fset, fileNode); err != nil {
					return fmt.Errorf("error in syntax pass for file %s: %w", filePath, err)
				}
			}
		}

		// Run global passes that operate on all files at once.
		fileMap := make(map[string]*ast.File)
		for i, filePath := range pkg.GoFiles {
			fileMap[filePath] = pkg.Syntax[i]
		}
		for _, pass := range obfuscator.globalPasses {
			if err := pass.Apply(obfuscator, fset, fileMap); err != nil {
				return fmt.Errorf("error in global pass for package %s: %w", pkg.Name, err)
			}
		}
	}

	// After all files in the main package have been processed, inject the decryption logic.
	if mainPkg != nil && obfuscator.stringEncryption != nil && len(obfuscator.stringEncryption.EncryptedStrings) > 0 {
		fmt.Println("Injecting string decryption logic into main package...")
		mainFile := mainPkg.Syntax[0] // Inject into the first file of the main package.
		injectStringDecryptor(obfuscator, fset, mainFile)
	}

	// Write all modified files to the output directory.
	for _, pkg := range pkgs {
		for i, filePath := range pkg.GoFiles {
			fileNode := pkg.Syntax[i]
			relPath, err := filepath.Rel(inputPath, filePath)
			if err != nil {
				return err
			}
			targetPath := filepath.Join(outputPath, relPath)
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}

			var buf bytes.Buffer
			if err := printer.Fprint(&buf, fset, fileNode); err != nil {
				return fmt.Errorf("failed to print AST for %s: %w", filePath, err)
			}

			if err := os.WriteFile(targetPath, buf.Bytes(), 0644); err != nil {
				return fmt.Errorf("failed to write output file %s: %w", targetPath, err)
			}
		}
	}

	return nil
}

// injectStringDecryptor adds the global variables and the decryption function to the file.
func injectStringDecryptor(obf *Obfuscator, fset *token.FileSet, file *ast.File) {
	// 1. Create the global slice for encrypted strings
	stringSlice := &ast.CompositeLit{Type: &ast.ArrayType{Elt: ast.NewIdent("string")}};
	for _, s := range obf.stringEncryption.EncryptedStrings {
		stringSlice.Elts = append(stringSlice.Elts, &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", s)})
	}
	stringsVar := &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{&ast.ValueSpec{
			Names:  []*ast.Ident{ast.NewIdent(obf.EncryptedStringsVarName)},
			Values: []ast.Expr{stringSlice},
		}},
	}

	// 2. Create the global slice for decryption keys
	keySlice := &ast.CompositeLit{Type: &ast.ArrayType{Elt: ast.NewIdent("byte")}};
	for _, k := range obf.stringEncryption.StringKeys {
		keySlice.Elts = append(keySlice.Elts, &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("0x%x", k)})
	}
	keysVar := &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{&ast.ValueSpec{
			Names:  []*ast.Ident{ast.NewIdent(obf.StringKeysVarName)},
			Values: []ast.Expr{keySlice},
		}},
	}

	// 3. Create the decryption function
	decryptionFunc := createDecryptorFunc(obf)

	// 4. Add the new declarations to the top of the file (after imports).
	decls := []ast.Decl{stringsVar, keysVar, decryptionFunc}

	// Insert after the last import
	lastImport := -1
	for i, decl := range file.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.IMPORT {
			lastImport = i
		}
	}
	if lastImport != -1 {
		file.Decls = append(file.Decls[:lastImport+1], append(decls, file.Decls[lastImport+1:]...)...)
	} else {
		// No imports, add at the beginning
		file.Decls = append(decls, file.Decls...)
	}
}

// createDecryptorFunc builds the AST for the global string decryption function.
// This function is intentionally convoluted to make it harder to analyze.
func createDecryptorFunc(obf *Obfuscator) *ast.FuncDecl {
	indexParam := "i"
	dataVar, keyVar, resultVar, iVar, bVar := NewName(), NewName(), NewName(), NewName(), NewName()

	return &ast.FuncDecl{
		Name: ast.NewIdent(obf.StringDecryptionFuncName),
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: []*ast.Field{
				{Names: []*ast.Ident{ast.NewIdent(indexParam)}, Type: ast.NewIdent("int")},
			}},
			Results: &ast.FieldList{List: []*ast.Field{
				{Type: ast.NewIdent("string")},
			}},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(dataVar)}, Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.IndexExpr{X: ast.NewIdent(obf.EncryptedStringsVarName), Index: ast.NewIdent(indexParam)}},
				},
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(keyVar)}, Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.IndexExpr{X: ast.NewIdent(obf.StringKeysVarName), Index: ast.NewIdent(indexParam)}},
				},
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(keyVar)}, Tok: token.XOR_ASSIGN,
					Rhs: []ast.Expr{&ast.CallExpr{
						Fun:  ast.NewIdent("byte"),
						Args: []ast.Expr{&ast.BinaryExpr{X: ast.NewIdent(obf.WeavingKeyVarName), Op: token.REM, Y: &ast.BasicLit{Kind: token.INT, Value: "256"}}},
					}},
				},
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(resultVar)}, Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.CallExpr{
						Fun: ast.NewIdent("make"),
						Args: []ast.Expr{
							&ast.ArrayType{Elt: ast.NewIdent("byte")},
							&ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{ast.NewIdent(dataVar)}},
						},
					}},
				},
				&ast.RangeStmt{
					Key: ast.NewIdent(iVar), Value: ast.NewIdent(bVar), Tok: token.DEFINE,
					X: &ast.CallExpr{Fun: ast.NewIdent("[]byte"), Args: []ast.Expr{ast.NewIdent(dataVar)}},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{&ast.IndexExpr{X: ast.NewIdent(resultVar), Index: ast.NewIdent(iVar)}}, Tok: token.ASSIGN,
								Rhs: []ast.Expr{&ast.BinaryExpr{X: ast.NewIdent(bVar), Op: token.XOR, Y: ast.NewIdent(keyVar)}},
							},
						},
					},
				},
				&ast.ReturnStmt{Results: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent("string"), Args: []ast.Expr{ast.NewIdent(resultVar)}}}},
			},
		},
	}
}
