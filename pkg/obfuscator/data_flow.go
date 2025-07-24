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

// DataFlowPass renames struct fields, global variables, and shuffles struct layouts.
type DataFlowPass struct{}

func (p *DataFlowPass) Apply(pkg *packages.Package) error {
	if err := p.renameGlobalsAndFields(pkg); err != nil {
		return fmt.Errorf("failed to rename globals and fields: %w", err)
	}

	if err := p.shuffleStructs(pkg); err != nil {
		return fmt.Errorf("failed to shuffle structs: %w", err)
	}

	return nil
}

// shuffleStructs finds all struct definitions and modifies their layout.
func (p *DataFlowPass) shuffleStructs(pkg *packages.Package) error {
	for _, file := range pkg.Syntax {
		astutil.Apply(file, func(cursor *astutil.Cursor) bool {
			structType, ok := cursor.Node().(*ast.StructType)
			if !ok || structType.Fields == nil || len(structType.Fields.List) == 0 {
				return true
			}

			// --- 1. Add dummy fields ---
			// Add 1 to 2 dummy fields to increase noise.
			numDummyFields := rand.Intn(2) + 1
			for i := 0; i < numDummyFields; i++ {
				dummyField := &ast.Field{
					Names: []*ast.Ident{ast.NewIdent("o_dummy_" + strconv.Itoa(rand.Intn(10000)))},
					Type:  ast.NewIdent("int"), // A common, simple type.
				}
				structType.Fields.List = append(structType.Fields.List, dummyField)
			}

			// --- 2. Shuffle all fields ---
			// This is safe because modern Go code almost always uses keyed literals
			// (e.g., MyStruct{Field: value}), which are not affected by order.
			// Unkeyed literals (MyStruct{value}) would break, but they are rare
			// and discouraged.
			rand.Shuffle(len(structType.Fields.List), func(i, j int) {
				structType.Fields.List[i], structType.Fields.List[j] = structType.Fields.List[j], structType.Fields.List[i]
			})

			fmt.Printf("    - Shuffled and added dummy fields to a struct in file %s\n", file.Name)

			// We've modified this struct, no need to traverse its children further.
			return false
		}, nil)
	}
	return nil
}

// renameGlobalsAndFields renames struct fields and global variables across an entire package.
func (p *DataFlowPass) renameGlobalsAndFields(pkg *packages.Package) error {
	renameMap := make(map[types.Object]string)

	// --- Pass 1: Collect objects to rename ---
	for ident, obj := range pkg.TypesInfo.Defs {
		if _, ok := obj.(*types.Var); !ok {
			continue
		}

		if obj.Name() == "_" || obj.Exported() || obj.Name() == "main" || obj.Name() == "init" {
			continue
		}

		isField := false
		if v, ok := obj.(*types.Var); ok && v.IsField() {
			isField = true
		}

		isGlobal := obj.Parent() == pkg.Types.Scope()

		if isField || isGlobal {
			newName := "o_df_" + strconv.Itoa(rand.Intn(10000))
			renameMap[obj] = newName
			fmt.Printf("    - Mapping %s to %s\n", ident.Name, newName)
		}
	}

	// --- Pass 2: Apply renaming ---
	for _, file := range pkg.Syntax {
		astutil.Apply(file, func(cursor *astutil.Cursor) bool {
			ident, ok := cursor.Node().(*ast.Ident)
			if !ok {
				return true
			}

			if obj, ok := pkg.TypesInfo.Uses[ident]; ok {
				if newName, ok := renameMap[obj]; ok {
					ident.Name = newName
				}
			}

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