package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"obfuscator/pkg/obfuscator"
)

func main() {
	inputPath := flag.String("input", "", "Path to the source directory or file")
	outputPath := flag.String("output", "./obfuscated_src", "Path to the output directory for the results")
	
	rename := flag.Bool("rename", true, "Enable identifier renaming")
	encryptStrings := flag.Bool("encrypt-strings", true, "Enable string encryption")
	insertDeadCode := flag.Bool("insert-dead-code", true, "Enable dead code insertion")
	obfuscateControlFlow := flag.Bool("obfuscate-control-flow", true, "Enable control flow obfuscation")
	obfuscateExpressions := flag.Bool("obfuscate-expressions", true, "Enable expression obfuscation")
	obfuscateDataFlow := flag.Bool("obfuscate-data-flow", true, "Enable data flow obfuscation (structs, globals)")
	obfuscateConstants := flag.Bool("obfuscate-constants", true, "Enable constant obfuscation")
	antiDebugging := flag.Bool("anti-debug", true, "Enable anti-debugging checks")

	flag.Parse()

	if *inputPath == "" {
		fmt.Println("Error: input path is not specified. Use -input flag.")
		flag.Usage()
		os.Exit(1)
	}

	absInput, err := filepath.Abs(*inputPath)
	if err != nil {
		fmt.Printf("Error getting absolute path for input: %v\n", err)
		os.Exit(1)
	}

	absOutput, err := filepath.Abs(*outputPath)
	if err != nil {
		fmt.Printf("Error getting absolute path for output: %v\n", err)
		os.Exit(1)
	}

	cfg := &obfuscator.Config{
		RenameIdentifiers:		*rename,
		EncryptStrings:			*encryptStrings,
		InsertDeadCode:			*insertDeadCode,
		ObfuscateControlFlow:	*obfuscateControlFlow,
		ObfuscateExpressions:	*obfuscateExpressions,
		ObfuscateDataFlow:		*obfuscateDataFlow,
		ObfuscateConstants:		*obfuscateConstants,
		AntiDebugging:			*antiDebugging,
	}

	fmt.Printf("Starting obfuscation...\n")
	fmt.Printf("Source: %s\n", absInput)
	fmt.Printf("Output: %s\n", absOutput)
	fmt.Printf("Configuration: %+v\n", cfg)

	err = obfuscator.ProcessDirectory(absInput, absOutput, cfg)
	if err != nil {
		fmt.Printf("\nCritical error during obfuscation: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nObfuscation completed successfully.")
}