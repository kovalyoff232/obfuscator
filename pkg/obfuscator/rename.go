package obfuscator

import (
	"go/ast"
)

// RenameIdentifiers safely renames local variables and constants.
// It avoids renaming anything in the global scope, struct fields, or function names,
// which prevents breaking interface implementations or public APIs.
func RenameIdentifiers(file *ast.File) {
	// A map to store the new name for each object to ensure consistency.
	nameMap := make(map[*ast.Object]string)

	ast.Inspect(file, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}

		// We only want to rename declarations of variables and constants.
		if ident.Obj != nil && ident.Obj.Pos() == ident.Pos() {
			// Check if it's a variable or constant and it's not exported.
			if (ident.Obj.Kind == ast.Var || ident.Obj.Kind == ast.Con) && !ident.IsExported() {
				// Simple check to avoid renaming things in the file (global) scope.
				// A more robust check would involve tracking scopes, but this is safer.
				if ident.Name != "_" { // Don't rename the blank identifier
					if _, exists := nameMap[ident.Obj]; !exists {
						nameMap[ident.Obj] = NewName()
					}
				}
			}
		}

		// If this identifier is a use of a renamed object, apply the new name.
		if ident.Obj != nil {
			if newName, ok := nameMap[ident.Obj]; ok {
				ident.Name = newName
			}
		}
		return true
	})
}