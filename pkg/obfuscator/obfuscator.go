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

// Pass - это интерфейс для одного прохода обфускации.
type Pass interface {
	Apply(file *ast.File) error
}

// --- Реализации Pass для каждой техники ---

// stringEncryptionPass реализует шифрование строк.
type stringEncryptionPass struct{}

func (p *stringEncryptionPass) Apply(file *ast.File) error {
	fmt.Println("  - Шифрование строк...")
	return EncryptStrings(file)
}

// renamePass реализует переименование идентификаторов.
type renamePass struct{}

func (p *renamePass) Apply(file *ast.File) error {
	fmt.Println("  - Переименование идентификаторов...")
	RenameIdentifiers(file)
	return nil
}

// deadCodePass реализует вставку "мертвого" кода.
type deadCodePass struct{}

func (p *deadCodePass) Apply(file *ast.File) error {
	fmt.Println("  - Вставка мусорного кода...")
	InsertDeadCode(file)
	return nil
}

// controlFlowPass реализует обфускацию потока управления.
type controlFlowPass struct{}

func (p *controlFlowPass) Apply(file *ast.File) error {
	fmt.Println("  - Обфускация потока управления...")
	ObfuscateControlFlow(file)
	return nil
}

// expressionPass реализует обфускацию выражений.
type expressionPass struct{}

func (p *expressionPass) Apply(file *ast.File) error {
	fmt.Println("  - Обфускация выражений...")
	ObfuscateExpressions(file)
	return nil
}


// Config определяет, какие техники обфускации будут применены.
type Config struct {
	RenameIdentifiers      bool
	EncryptStrings         bool
	InsertDeadCode         bool
	ObfuscateControlFlow   bool
	ObfuscateExpressions   bool
}

// Obfuscator управляет процессом обфускации.
type Obfuscator struct {
	passes []Pass
}

// NewObfuscator создает новый экземпляр обфускатора на основе конфигурации.
func NewObfuscator(cfg *Config) *Obfuscator {
	var passes []Pass

	// Порядок важен. Переименование лучше делать первым.
	if cfg.RenameIdentifiers {
		passes = append(passes, &renamePass{})
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
	if cfg.ObfuscateExpressions {
		passes = append(passes, &expressionPass{})
	}

	return &Obfuscator{passes: passes}
}

// runOnFile применяет все настроенные проходы к одному файлу.
func (o *Obfuscator) runOnFile(node *ast.File) error {
	for _, pass := range o.passes {
		if err := pass.Apply(node); err != nil {
			return err
		}
	}
	return nil
}

// ProcessDirectory обходит директорию, читает .go файлы и применяет обфускацию.
func ProcessDirectory(inputPath, outputPath string, cfg *Config) error {
	if err := os.RemoveAll(outputPath); err != nil {
		return fmt.Errorf("не удалось очистить выходную директорию: %w", err)
	}
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("не удалось создать выходную директорию: %w", err)
	}

	obfuscator := NewObfuscator(cfg);

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

		fmt.Printf("Обрабатываю файл: %s -> %s\n", path, targetPath)

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("ошибка парсинга файла %s: %w", path, err)
		}

		if err := obfuscator.runOnFile(node); err != nil {
			return fmt.Errorf("ошибка обфускации файла %s: %w", path, err)
		}

		outputFile, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("не удалось создать выходной файл %s: %w", targetPath, err)
		}
		defer outputFile.Close()

		return printer.Fprint(outputFile, fset, node)
	})
}
