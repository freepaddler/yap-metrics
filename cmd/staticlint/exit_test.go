package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"testing"
)

const src1 = `package main

import (
	//"os"
	osx "os"
)

func main() {
	osx.Exit(1)
}
`

func Test_Exit(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src1, 0)
	if err != nil {
		panic(err)
	}
	//ast.Print(fset, f)
	var mainPackage, mainFunc bool
	var osImport string
	ast.Inspect(f, func(node ast.Node) bool {
		switch x := node.(type) {
		case *ast.ImportSpec:
			fmt.Print("import spec: ")
			printer.Fprint(os.Stdout, fset, x)
			fmt.Println()
			if x.Path.Value == "\"os\"" {
				osImport = "os"
				fmt.Println("it is os import")
				if x.Name != nil && x.Name.Name != "" {
					osImport = x.Name.Name
				}
				fmt.Println("osImport is:", osImport)
			}
		case *ast.File:
			if x.Name.Name == "main" {
				mainPackage = true
			} else {
				mainPackage = false
			}
		case *ast.FuncDecl:
			if x.Name.Name == "main" {
				mainFunc = true
			} else {
				mainFunc = false
			}
		case *ast.CallExpr:
			//fmt.Println("CallExpr:", fset.Position(x.Pos()))
			//ast.Print(fset, x)
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				if pfx, ok := sel.X.(*ast.Ident); ok {
					if pfx.Name == osImport && sel.Sel.Name == "Exit" && mainFunc && mainPackage {
						fmt.Println(fset.Position(x.Pos()), "bad exit is here")
					}
				}
			}

		}
		return true
	})
}
