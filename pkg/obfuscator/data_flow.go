package obfuscator

import (
	"fmt"
	"go/ast"
	"go/types"
	"math/rand"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

// DataFlowPass renames struct fields and global variables across an entire package.
type DataFlowPass struct{}

func (p *DataFlowPass) Apply(pkg *packages.Package) error {
	// This map stores the new name for each object that we decide to rename.
	renameMap := make(map[types.Object]string)

	// --- Pass 1: Collect objects to rename ---
	// We iterate over all definitions in the package to find things to rename.
	for ident, obj := range pkg.TypesInfo.Defs {
		// We only care about variables (includes struct fields and globals) and functions.
		if _, ok := obj.(*types.Var); !ok {
			continue
		}

		// Don't rename blank identifiers or exported fields/variables for now.
		// A more advanced version could handle exported names carefully.
		if obj.Name() == "_" || obj.Exported() {
			continue
		}
		
		// Don't rename main/init
		if obj.Name() == "main" || obj.Name() == "init" {
			continue
		}

		// Check if it's a field in a struct.
		isField := false
		if v, ok := obj.(*types.Var); ok && v.IsField() {
			isField = true
		}

		// Check if it's a package-level global variable.
		isGlobal := obj.Parent() == pkg.Types.Scope()

		if isField || isGlobal {
			// It's a candidate for renaming.
			newName := "o_df_" + strconv.Itoa(rand.Intn(10000))
			renameMap[obj] = newName
			fmt.Printf("    - Mapping %s to %s\n", ident.Name, newName)
		}
	}

	// --- Pass 2: Apply renaming ---
	// Now we walk the AST of each file and replace all uses of the objects we've mapped.
	for _, file := range pkg.Syntax {
		astutil.Apply(file, func(cursor *astutil.Cursor) bool {
			ident, ok := cursor.Node().(*ast.Ident)
			if !ok {
				return true
			}

			// Check if this identifier is a "use" of an object we need to rename.
			if obj, ok := pkg.TypesInfo.Uses[ident]; ok {
				if newName, ok := renameMap[obj]; ok {
					ident.Name = newName
				}
			}
			
			// Also check if it's a "definition" of an object we need to rename.
			if obj, ok := pkg.TypesInfo.Defs[ident]; ok {
				if newName, ok := renameMap[obj]; ok {
					ident.Name = newName
				}
			}

			return true
		}, nil)
	}

	return nil
}
