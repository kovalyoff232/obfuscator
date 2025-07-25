package obfuscator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/ast/astutil"
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
	StringIVsVarName         string // New variable for IVs
	stringEncryption         *StringEncryptionPass
}

func NewObfuscator(cfg *Config) *Obfuscator {
	obf := &Obfuscator{
		WeavingKeyVarName:        NewName(),
		StringDecryptionFuncName: NewName(),
		EncryptedStringsVarName:  NewName(),
		StringKeysVarName:        NewName(),
		StringIVsVarName:         NewName(), // Initialize new variable
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
	var mainPkgPath string

	for _, pkg := range pkgs {
		if pkg.Name == "main" {
			mainPkg = pkg
			if len(pkg.GoFiles) > 0 {
				mainPkgPath = filepath.Dir(pkg.GoFiles[0])
			}
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

	// After all files in the main package have been processed, inject the decryption logic into a new file.
	if mainPkg != nil && obfuscator.stringEncryption != nil && len(obfuscator.stringEncryption.EncryptedStrings) > 0 {
		fmt.Println("Injecting string decryption logic into a new file...")
		
		// Create a new file AST for the injected code.
		injectedFile := &ast.File{
			Name: ast.NewIdent("main"),
		}
		
		injectStringDecryptor(obfuscator, fset, injectedFile);

		// Determine the output path for the new file.
		relPath, err := filepath.Rel(inputPath, mainPkgPath)
		if err != nil {
			return fmt.Errorf("could not determine relative path for main package: %w", err)
		}
		injectedFilePath := filepath.Join(outputPath, relPath, "o_injected.go")

		// Write the new file.
		var buf bytes.Buffer
		if err := printer.Fprint(&buf, fset, injectedFile); err != nil {
			return fmt.Errorf("failed to print injected AST: %w", err)
		}
		if err := os.WriteFile(injectedFilePath, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("failed to write injected file %s: %w", injectedFilePath, err)
		}
		fmt.Printf("  - Injected file created at: %s\n", injectedFilePath)
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

// injectStringDecryptor adds the global variables and the AES decryption function to the file.
func injectStringDecryptor(obf *Obfuscator, fset *token.FileSet, file *ast.File) {
	astutil.AddImport(fset, file, "crypto/aes")
	astutil.AddImport(fset, file, "crypto/cipher")

	// Helper to create a `[][]byte` literal
	createByteSliceLiteral := func(data [][]byte) *ast.CompositeLit {
		slice := &ast.CompositeLit{Type: &ast.ArrayType{Elt: &ast.ArrayType{Elt: ast.NewIdent("byte")}}}
		for _, d := range data {
			byteSlice := &ast.CompositeLit{}
			for _, b := range d {
				byteSlice.Elts = append(byteSlice.Elts, &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("0x%x", b)})
			}
			slice.Elts = append(slice.Elts, byteSlice)
		}
		return slice
	}

	// 1. Create the global slice for encrypted strings
	stringSlice := &ast.CompositeLit{Type: &ast.ArrayType{Elt: ast.NewIdent("string")}}
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

	// 2. Create global slices for keys and IVs
	keysVar := &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{&ast.ValueSpec{
			Names:  []*ast.Ident{ast.NewIdent(obf.StringKeysVarName)},
			Values: []ast.Expr{createByteSliceLiteral(obf.stringEncryption.Keys)},
		}},
	}
	ivsVar := &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{&ast.ValueSpec{
			Names:  []*ast.Ident{ast.NewIdent(obf.StringIVsVarName)},
			Values: []ast.Expr{createByteSliceLiteral(obf.stringEncryption.IVs)},
		}},
	}

	// 3. Create the decryption function
	decryptionFunc := createAESDecryptorFunc(obf)

	// 4. Add the new declarations to the file
	file.Decls = append(file.Decls, stringsVar, keysVar, ivsVar, decryptionFunc)
}

// createAESDecryptorFunc builds the AST for the AES-CTR decryption function.
func createAESDecryptorFunc(obf *Obfuscator) *ast.FuncDecl {
	indexParam := "i"
	dataVar, keyVar, ivVar, blockVar, streamVar, resultVar, errVar :=
		NewName(), NewName(), NewName(), NewName(), NewName(), NewName(), NewName()

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
				// data := encryptedStrings[i]
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(dataVar)}, Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.IndexExpr{X: ast.NewIdent(obf.EncryptedStringsVarName), Index: ast.NewIdent(indexParam)}},
				},
				// key := stringKeys[i]
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(keyVar)}, Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.IndexExpr{X: ast.NewIdent(obf.StringKeysVarName), Index: ast.NewIdent(indexParam)}},
				},
				// iv := stringIVs[i]
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(ivVar)}, Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.IndexExpr{X: ast.NewIdent(obf.StringIVsVarName), Index: ast.NewIdent(indexParam)}},
				},
				// block, err := aes.NewCipher(key)
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(blockVar), ast.NewIdent(errVar)}, Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.CallExpr{
						Fun:  &ast.SelectorExpr{X: ast.NewIdent("aes"), Sel: ast.NewIdent("NewCipher")},
						Args: []ast.Expr{ast.NewIdent(keyVar)},
					}},
				},
				// if err != nil { panic(err) }
				&ast.IfStmt{
					Cond: &ast.BinaryExpr{X: ast.NewIdent(errVar), Op: token.NEQ, Y: ast.NewIdent("nil")},
					Body: &ast.BlockStmt{List: []ast.Stmt{
						&ast.ExprStmt{X: &ast.CallExpr{Fun: ast.NewIdent("panic"), Args: []ast.Expr{ast.NewIdent(errVar)}}},
					}},
				},
				// result := make([]byte, len(data))
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(resultVar)}, Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.CallExpr{
						Fun:  ast.NewIdent("make"),
						Args: []ast.Expr{&ast.ArrayType{Elt: ast.NewIdent("byte")}, &ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{ast.NewIdent(dataVar)}}},
					}},
				},
				// stream := cipher.NewCTR(block, iv)
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(streamVar)}, Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.CallExpr{
						Fun:  &ast.SelectorExpr{X: ast.NewIdent("cipher"), Sel: ast.NewIdent("NewCTR")},
						Args: []ast.Expr{ast.NewIdent(blockVar), ast.NewIdent(ivVar)},
					}},
				},
				// stream.XORKeyStream(result, []byte(data))
				&ast.ExprStmt{X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{X: ast.NewIdent(streamVar), Sel: ast.NewIdent("XORKeyStream")},
					Args: []ast.Expr{
						ast.NewIdent(resultVar),
						&ast.CallExpr{Fun: ast.NewIdent("[]byte"), Args: []ast.Expr{ast.NewIdent(dataVar)}},
					},
				}},
				// return string(result)
				&ast.ReturnStmt{Results: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent("string"), Args: []ast.Expr{ast.NewIdent(resultVar)}}}},
			},
		},
	}
}
