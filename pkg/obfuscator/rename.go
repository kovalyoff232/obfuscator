package obfuscator

import (
	"fmt"
	"go/ast"
	"math/rand"
	"time"
)

func isSafeToRename(ident *ast.Ident) bool {
	if ident == nil || ident.Obj == nil || ident.Obj.Decl == nil {
		return false
	}

	// Check if the identifier is part of a variable declaration that is now a function.
	// This is a sign that DataFlowPass has already processed it.
	if spec, ok := ident.Obj.Decl.(*ast.ValueSpec); ok {
		for _, val := range spec.Values {
			if _, isFuncLit := val.(*ast.FuncLit); isFuncLit {
				return false
			}
		}
	}

	// Do not rename exported identifiers, the main/init functions, or the blank identifier.
	return !ident.IsExported() && ident.Name != "main" && ident.Name != "init" && ident.Name != "_"
}

// RenameIdentifiers traverses the AST and renames all non-exported identifiers.
func RenameIdentifiers(node *ast.File) {
	// Seed the random number generator to ensure different names on each run.
	rand.Seed(time.Now().UnixNano())

	// A map to store the new name for each object.
	nameMap := make(map[*ast.Object]string)

	// First pass: find all declarations and assign a new name.
	ast.Inspect(node, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}

		// If this identifier is a declaration, it's a candidate for renaming.
		if ident.Obj != nil && ident.Obj.Decl == n {
			if isSafeToRename(ident) {
				// Ensure we don't accidentally generate the same name for different objects.
				if _, exists := nameMap[ident.Obj]; !exists {
					nameMap[ident.Obj] = NewName()
				}
			}
		}
		return true
	})

	// Second pass: apply the new names to all uses of the identifiers.
	ast.Inspect(node, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		if ident.Obj != nil {
			if newName, ok := nameMap[ident.Obj]; ok {
				ident.Name = newName
			}
		}
		return true
	})
	fmt.Println("  - Renaming identifiers...")
}
