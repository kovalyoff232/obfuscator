package obfuscator

import (
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)


// Config определяет, какие техники обфускации будут применены.
type Config struct {
	RenameIdentifiers      bool
	EncryptStrings         bool
	InsertDeadCode         bool
	ObfuscateControlFlow   bool
	ObfuscateExpressions   bool
}

// ProcessDirectory обходит директорию, читает .go файлы и применяет обфускацию.
func ProcessDirectory(inputPath, outputPath string, cfg *Config) error {
	// Очищаем и создаем выходную директорию
	if err := os.RemoveAll(outputPath); err != nil {
		return fmt.Errorf("не удалось очистить выходную директорию: %w", err)
	}
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("не удалось создать выходную директорию: %w", err)
	}

	return filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		// Создаем правильную структуру директорий в output
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

		// Применяем все слои обфускации в соответствии с конфигурацией
		if cfg.RenameIdentifiers {
			fmt.Println("  - Переименование идентификаторов...")
			RenameIdentifiers(node)
		}

		if cfg.EncryptStrings {
			fmt.Println("  - Шифрование строк...")
			if err := EncryptStrings(node); err != nil {
				return fmt.Errorf("ошибка шифрования строк в %s: %w", path, err)
			}
		}

		if cfg.InsertDeadCode {
			fmt.Println("  - Вставка мусорного кода...")
			InsertDeadCode(node)
		}

		if cfg.ObfuscateControlFlow {
			fmt.Println("  - Обфускация потока управления...")
			ObfuscateControlFlow(node)
		}

		if cfg.ObfuscateExpressions {
			fmt.Println("  - Обфускация выражений...")
			ObfuscateExpressions(node)
		}

		// Записываем измененное AST в новый файл
		fmt.Println("  - Запись результата...")
		outputFile, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("не удалось создать выходной файл %s: %w", targetPath, err)
		}
		defer outputFile.Close()

		return printer.Fprint(outputFile, fset, node)
	})
}
