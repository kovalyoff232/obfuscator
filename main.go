package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"obfuscator/pkg/obfuscator"
)

func main() {
	inputPath := flag.String("input", "", "Путь к исходной директории или файлу")
	outputPath := flag.String("output", "./obfuscated_src", "Путь к директории для сохранения результатов")
	
	// Флаги для управления техниками обфускации
	rename := flag.Bool("rename", true, "Включить переименование идентификаторов")
	encryptStrings := flag.Bool("encrypt-strings", true, "Включить шифрование строк")
	insertDeadCode := flag.Bool("insert-dead-code", true, "Включить вставку мертвого кода")
	obfuscateControlFlow := flag.Bool("obfuscate-control-flow", true, "Включить обфускацию потока управления")
	obfuscateExpressions := flag.Bool("obfuscate-expressions", true, "Включить обфускацию выражений")

	flag.Parse()

	if *inputPath == "" {
		fmt.Println("Ошибка: не указан путь к исходникам. Используйте флаг -input.")
		flag.Usage()
		os.Exit(1)
	}

	absInput, err := filepath.Abs(*inputPath)
	if err != nil {
		fmt.Printf("Ошибка получения абсолютного пути для input: %v\n", err)
		os.Exit(1)
	}

	absOutput, err := filepath.Abs(*outputPath)
	if err != nil {
		fmt.Printf("Ошибка получения абсолютного пути для output: %v\n", err)
		os.Exit(1)
	}

	// Создаем конфигурацию на основе флагов
	cfg := &obfuscator.Config{
		RenameIdentifiers:      *rename,
		EncryptStrings:         *encryptStrings,
		InsertDeadCode:         *insertDeadCode,
		ObfuscateControlFlow:   *obfuscateControlFlow,
		ObfuscateExpressions:   *obfuscateExpressions,
	}

	fmt.Printf("Начинаю обфускацию...\n")
	fmt.Printf("Источник: %s\n", absInput)
	fmt.Printf("Результат: %s\n", absOutput)
	fmt.Printf("Конфигурация: %+v\n", cfg)

	err = obfuscator.ProcessDirectory(absInput, absOutput, cfg)
	if err != nil {
		fmt.Printf("\nКритическая ошибка в процессе обфускации: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nОбфускация успешно завершена.")
}