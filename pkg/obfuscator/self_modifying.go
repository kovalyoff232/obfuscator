package obfuscator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"
)

// SelfModifyingPass transforms suitable functions into a self-modifying form.
type SelfModifyingPass struct{}

func (p *SelfModifyingPass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		fn, ok := cursor.Node().(*ast.FuncDecl)
		// For this PoC, we target a specific function. A real implementation
		// would have heuristics to find good candidates.
		if !ok || fn.Name.Name != "getOperationMessage" {
			return true
		}

		// 1. In a real scenario, we'd compile fn.Body to machine code.
		// For now, we'll just use a placeholder string.
		originalCode := `fmt.Println("This is the original, now 'encrypted' function body.")`
		encryptedCode := []byte(originalCode) // Placeholder for actual encryption

		// 2. Create the new function body (the "stub").
		newBody, err := p.createStubBody(fn, encryptedCode)
		if err != nil {
			// Cannot transform this function, so we skip it.
			return true
		}

		fn.Body = newBody

		// Add necessary imports for the new body
		astutil.AddImport(fset, file, "fmt")
		astutil.AddImport(fset, file, "syscall")
		astutil.AddImport(fset, file, "unsafe")

		// We've modified the function, no need to traverse its children.
		return false
	}, nil)

	return nil
}

// createStubBody generates the new function body that contains the
// self-modifying logic.
func (p *SelfModifyingPass) createStubBody(fn *ast.FuncDecl, encryptedCode []byte) (*ast.BlockStmt, error) {
	// This is a simplified template. A real implementation would be much more complex.
	template := `
package main

func temp() {
	// 1. Placeholder for the encrypted original function body.
	encryptedFunc := %#v

	// 2. Allocate executable memory using mmap.
	// PROT_READ | PROT_WRITE | PROT_EXEC = 7
	// MAP_ANON | MAP_PRIVATE = 0x22
	mem, err := syscall.Mmap(-1, 0, len(encryptedFunc), 7, 0x22)
	if err != nil {
		panic("mmap failed")
	}

	// 3. Copy the 'decrypted' code into the executable memory.
	copy(mem, encryptedFunc)

	// 4. Create a function pointer to the executable memory.
	// This part is highly platform-specific and simplified here.
	// The signature must match the original function.
	type fnType func(string) string
	executableFunc := *(*fnType)(unsafe.Pointer(&mem))
	_ = executableFunc // Avoid "declared and not used" error

	// 5. Call the function from the new memory location.
	// We are not actually calling it in this PoC to avoid complexity.
	fmt.Println("Function has been dynamically prepared and would be executed now.")

	// 6. Return a default value matching the original function's signature.
	return "self-modified"
}`

	// Create the code for the new body
	sourceCode := fmt.Sprintf(template, encryptedCode)

	// Parse the source code into a full AST file.
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "temp.go", sourceCode, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source code: %w", err)
	}

	// Find the function declaration in the parsed file.
	for _, decl := range file.Decls {
		if fnDecl, ok := decl.(*ast.FuncDecl); ok {
			return fnDecl.Body, nil
		}
	}

	return nil, fmt.Errorf("failed to find function body in parsed source")
}