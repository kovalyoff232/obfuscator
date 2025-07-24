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

// --- Pass Implementations ---

type stringEncryptionPass struct{}

func (p *stringEncryptionPass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	fmt.Println("  - Encrypting strings with weaving...")
	return EncryptStrings_v2(obf, fset, file)
}

type renamePass struct{}

func (p *renamePass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	fmt.Println("  - Renaming identifiers...")
	RenameIdentifiers(file)
	return nil
}

type deadCodePass struct{}

func (p *deadCodePass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	fmt.Println("  - Inserting junk code...")
	InsertDeadCode(file)
	return nil
}

type controlFlowPass struct{}

func (p *controlFlowPass) Apply(obf *Obfuscator, pkg *packages.Package) error {
	fmt.Println("  - Obfuscating control flow...")
	for _, file := range pkg.Syntax {
		ControlFlow(file, pkg.TypesInfo)
	}
	return nil
}

type expressionPass struct{}

func (p *expressionPass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	fmt.Println("  - Obfuscating expressions...")
	ObfuscateExpressions(file)
	return nil
}

type constantPass struct{}

func (p *constantPass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	fmt.Println("  - Obfuscating constants...")
	ObfuscateConstants(file)
	return nil
}

type antiDebugPass struct{}

func (p *antiDebugPass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	fmt.Println("  - Injecting anti-debugging checks...")
	pass := &AntiDebugPass{}
	return pass.Apply(obf, fset, file)
}

type antiVMPass struct{}

func (p *antiVMPass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	fmt.Println("  - Injecting anti-VM checks...")
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
	WeavingKeyVarName string // Name of the global var for the decryption key
}

func NewObfuscator(cfg *Config) *Obfuscator {
	var syntaxPasses []Pass
	var globalPasses []GlobalPass
	var typeAwarePasses []TypeAwarePass

	obf := &Obfuscator{
		WeavingKeyVarName: RandomIdentifier("o_wkey"),
	}

	// Weaving related passes must come first.
	// Anti-debugging will create the key, and string encryption will use it.
	if cfg.AntiDebugging {
		syntaxPasses = append(syntaxPasses, &antiDebugPass{})
	}
	if cfg.EncryptStrings {
		syntaxPasses = append(syntaxPasses, &stringEncryptionPass{})
	}

	// Other passes
	if cfg.AntiVM {
		syntaxPasses = append(syntaxPasses, &antiVMPass{})
	}
	if cfg.ObfuscateDataFlow {
		typeAwarePasses = append(typeAwarePasses, &DataFlowPass{})
	}
	if cfg.RenameIdentifiers {
		syntaxPasses = append(syntaxPasses, &renamePass{})
	}
	if cfg.ObfuscateConstants {
		syntaxPasses = append(syntaxPasses, &constantPass{})
	}
	if cfg.ObfuscateExpressions {
		syntaxPasses = append(syntaxPasses, &expressionPass{})
	}
	if cfg.InsertDeadCode {
		syntaxPasses = append(syntaxPasses, &deadCodePass{})
	}
	if cfg.ObfuscateControlFlow {
		typeAwarePasses = append(typeAwarePasses, &controlFlowPass{})
	}
	if cfg.IndirectCalls {
		globalPasses = append(globalPasses, &CallIndirectionPass{})
	}

	obf.syntaxPasses = syntaxPasses
	obf.globalPasses = globalPasses
	obf.typeAwarePasses = typeAwarePasses

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
			if err := pass.Apply(obfuscator, pkg); err != nil {
				return fmt.Errorf("error in type-aware pass for package %s: %w", pkg.Name, err)
			}
		}

		// Run syntax-only passes on each file individually.
		for i, filePath := range pkg.GoFiles {
			fileNode := pkg.Syntax[i]

			fmt.Printf("Processing file: %s\n", filePath)
			for _, pass := range obfuscator.syntaxPasses {
				if _, ok := pass.(*renamePass); ok && filepath.Base(filePath) == "main.go" {
					fmt.Println("  - Skipping identifier renaming for main.go to preserve critical variables.")
					continue
				}
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
