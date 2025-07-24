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
	Apply(fset *token.FileSet, file *ast.File) error
}

// TypeAwarePass represents a semantic obfuscation pass that requires type info for the whole package.
type TypeAwarePass interface {
	Apply(pkg *packages.Package) error
}

// --- Pass Implementations ---

type stringEncryptionPass struct{}

func (p *stringEncryptionPass) Apply(fset *token.FileSet, file *ast.File) error {
	fmt.Println("  - Encrypting strings...")
	return EncryptStrings(fset, file)
}

type renamePass struct{}

func (p *renamePass) Apply(fset *token.FileSet, file *ast.File) error {
	fmt.Println("  - Renaming identifiers...")
	RenameIdentifiers(file)
	return nil
}

type deadCodePass struct{}

func (p *deadCodePass) Apply(fset *token.FileSet, file *ast.File) error {
	fmt.Println("  - Inserting junk code...")
	InsertDeadCode(file)
	return nil
}

type controlFlowPass struct{}

func (p *controlFlowPass) Apply(fset *token.FileSet, file *ast.File) error {
	fmt.Println("  - Obfuscating control flow...")
	ObfuscateControlFlow(file)
	return nil
}

type expressionPass struct{}

func (p *expressionPass) Apply(fset *token.FileSet, file *ast.File) error {
	fmt.Println("  - Obfuscating expressions...")
	ObfuscateExpressions(file)
	return nil
}

// --- New Type-Aware Pass ---

// (This is now defined in data_flow.go)

// --- Configuration and Orchestration ---

type Config struct {
	RenameIdentifiers    bool
	EncryptStrings       bool
	InsertDeadCode       bool
	ObfuscateControlFlow bool
	ObfuscateExpressions bool
	ObfuscateDataFlow    bool // New flag
}

type Obfuscator struct {
	syntaxPasses     []Pass
	typeAwarePasses  []TypeAwarePass
}

func NewObfuscator(cfg *Config) *Obfuscator {
	var syntaxPasses []Pass
	var typeAwarePasses []TypeAwarePass

	// Order is important here. Renaming should generally go first.
	if cfg.RenameIdentifiers {
		syntaxPasses = append(syntaxPasses, &renamePass{})
	}
	// Data flow obfuscation should run before other things that might break type analysis.
	if cfg.ObfuscateDataFlow {
		typeAwarePasses = append(typeAwarePasses, &DataFlowPass{})
	}
	if cfg.ObfuscateExpressions {
		syntaxPasses = append(syntaxPasses, &expressionPass{})
	}
	if cfg.EncryptStrings {
		syntaxPasses = append(syntaxPasses, &stringEncryptionPass{})
	}
	if cfg.InsertDeadCode {
		syntaxPasses = append(syntaxPasses, &deadCodePass{})
	}
	if cfg.ObfuscateControlFlow {
		syntaxPasses = append(syntaxPasses, &controlFlowPass{})
	}

	return &Obfuscator{
		syntaxPasses:    syntaxPasses,
		typeAwarePasses: typeAwarePasses,
	}
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

	// Load and type-check the package.
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
		fmt.Printf("Processing package: %s\n", pkg.Name)

		// Run type-aware passes that operate on the whole package at once.
		for _, pass := range obfuscator.typeAwarePasses {
			if err := pass.Apply(pkg); err != nil {
				return fmt.Errorf("error in type-aware pass for package %s: %w", pkg.Name, err)
			}
		}

		// Run syntax-only passes on each file individually.
		for i, filePath := range pkg.GoFiles {
			fileNode := pkg.Syntax[i]

			fmt.Printf("Processing file: %s\n", filePath)
			for _, pass := range obfuscator.syntaxPasses {
				if err := pass.Apply(fset, fileNode); err != nil {
					return fmt.Errorf("error in syntax pass for file %s: %w", filePath, err)
				}
			}
		}

		// Write the modified files to the output directory.
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