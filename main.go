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
	
	flag.Parse()

	if *inputPath == "" {
		fmt.Println("Ошибка: не указан путь к исходникам. Используйте флаг -input.")
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

	fmt.Printf("Начинаю обфускацию...\n")
	fmt.Printf("Источник: %s\n", absInput)
	fmt.Printf("Результат: %s\n", absOutput)

	err = obfuscator.ProcessDirectory(absInput, absOutput)
	if err != nil {
		fmt.Printf("\nКритическая ошибка в процессе обфускации: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nОбфускация успешно завершена.")
}