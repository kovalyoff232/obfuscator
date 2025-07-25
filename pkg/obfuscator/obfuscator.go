package obfuscator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"math/rand"
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

// AddMetamorphicCode walks the AST and inserts junk code into function bodies.
func AddMetamorphicCode(file *ast.File) {
	engine := &MetamorphicEngine{}
	ast.Inspect(file, func(n ast.Node) bool {
		// We only care about function bodies
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil || len(fn.Body.List) == 0 {
			return true
		}

		// Don't add junk to tiny functions
		if len(fn.Body.List) < 2 {
			return true
		}

		// Insert junk code at a random position
		if rand.Intn(100) < 30 { // 30% chance to add junk
			junk := engine.GenerateJunkCodeBlock()
			insertionPoint := rand.Intn(len(fn.Body.List))

			// Prepend to avoid issues with slice indexing
			fn.Body.List = append(fn.Body.List[:insertionPoint], append(junk, fn.Body.List[insertionPoint:]...)...)
		}

		return true
	})
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

type metamorphicPass struct{}

func (p *metamorphicPass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	AddMetamorphicCode(file)
	return nil
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
	WeaveIntegrity       bool
	AddMetamorphicCode   bool
}

type Obfuscator struct {
	syntaxPasses      []Pass
	globalPasses      []GlobalPass
	typeAwarePasses   []TypeAwarePass
	WeavingKeyVarName string // Name of the global var for the anti-debug key
	stringEncryption  *StringEncryptionPass
	integrityWeaver   *IntegrityWeavingPass
}

func NewObfuscator(cfg *Config) *Obfuscator {
	obf := &Obfuscator{
		WeavingKeyVarName: NewName(),
	}

	// --- Pass Ordering ---
	if cfg.AntiDebugging {
		obf.syntaxPasses = append(obf.syntaxPasses, &antiDebugPass{})
	}
	if cfg.EncryptStrings {
		obf.stringEncryption = NewStringEncryptionPass()
	}
	if cfg.AntiVM {
		obf.syntaxPasses = append(obf.syntaxPasses, &antiVMPass{})
	}
	if cfg.ObfuscateDataFlow {
		obf.typeAwarePasses = append(obf.typeAwarePasses, &DataFlowPass{})
	}
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
	if cfg.AddMetamorphicCode {
		obf.syntaxPasses = append(obf.syntaxPasses, &metamorphicPass{})
	}
	if cfg.ObfuscateControlFlow {
		obf.typeAwarePasses = append(obf.typeAwarePasses, &controlFlowPass{})
	}
	if cfg.IndirectCalls {
		obf.globalPasses = append(obf.globalPasses, &CallIndirectionPass{})
	}
	// Integrity weaving must be the absolute last pass.
	if cfg.WeaveIntegrity {
		obf.integrityWeaver = NewIntegrityWeavingPass()
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

	for _, pkg := range pkgs {
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

		// Run the final integrity weaving pass if enabled.
		if obfuscator.integrityWeaver != nil {
			if err := obfuscator.integrityWeaver.Apply(obfuscator, fset, fileMap); err != nil {
				return fmt.Errorf("error in integrity weaving pass for package %s: %w", pkg.Name, err)
			}
		}
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
