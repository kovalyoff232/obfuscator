package obfuscator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type Pass interface {
	Apply(file *ast.File) error
}
type stringEncryptionPass struct{}

func (p *stringEncryptionPass) Apply(file *ast.File) error {
	fmt.Println("  - Encrypting strings...")
	return EncryptStrings(file)
}

type renamePass struct{}

func (p *renamePass) Apply(file *ast.File) error {
	fmt.Println("  - Renaming identifiers...")
	RenameIdentifiers(file)
	return nil
}

type deadCodePass struct{}

func (p *deadCodePass) Apply(file *ast.File) error {
	fmt.Println("  - Inserting junk code...")
	InsertDeadCode(file)
	return nil
}

type controlFlowPass struct{}

func (p *controlFlowPass) Apply(file *ast.File) error {
	fmt.Println("  - Obfuscating control flow...")
	ObfuscateControlFlow(file)
	return nil
}

type expressionPass struct{}

func (p *expressionPass) Apply(file *ast.File) error {
	fmt.Println("  - Obfuscating expressions...")
	ObfuscateExpressions(file)
	return nil
}

type Config struct {
	RenameIdentifiers      bool
	EncryptStrings         bool
	InsertDeadCode         bool
	ObfuscateControlFlow   bool
	ObfuscateExpressions   bool
}

type Obfuscator struct {
	passes []Pass
}

func NewObfuscator(cfg *Config) *Obfuscator {
	var passes []Pass

	if cfg.RenameIdentifiers {
		passes = append(passes, &renamePass{})
	}

	if cfg.ObfuscateExpressions {
		passes = append(passes, &expressionPass{})
	}
	if cfg.EncryptStrings {
		passes = append(passes, &stringEncryptionPass{})
	}
	if cfg.InsertDeadCode {
		passes = append(passes, &deadCodePass{})
	}
	if cfg.ObfuscateControlFlow {
		passes = append(passes, &controlFlowPass{})
	}

	return &Obfuscator{passes: passes}
}

func (o *Obfuscator) runOnFile(node *ast.File) error {
	for _, pass := range o.passes {
		if err := pass.Apply(node); err != nil {
			return err
		}
	}
	return nil
}

func ProcessDirectory(inputPath, outputPath string, cfg *Config) error {
	if err := os.RemoveAll(outputPath); err != nil {
		return fmt.Errorf("failed to clean output directory: %w", err)
	}
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	obfuscator := NewObfuscator(cfg)

	return filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		relPath, err := filepath.Rel(inputPath, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(outputPath, relPath)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		fmt.Printf("Processing file: %s -> %s\n", path, targetPath)

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("error parsing file %s: %w", path, err)
		}

		if err := obfuscator.runOnFile(node); err != nil {
			return fmt.Errorf("error obfuscating file %s: %w", path, err)
		}

		outputFile, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", targetPath, err)
		}
		defer outputFile.Close()

		return printer.Fprint(outputFile, fset, node)
	})
}
