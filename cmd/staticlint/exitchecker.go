package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// ExitCheckAnalyzer checks that there are no direct os.Exit calls in func main() of package main
var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for os.Exit in main func of main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		var mainPackage, mainFunc bool
		var osImport string
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.File:
				if x.Name.Name == "main" {
					mainPackage = true
				} else {
					mainPackage = false
				}
			case *ast.ImportSpec:
				if x.Path.Value == "\"os\"" {
					osImport = "os"
					if x.Name != nil && x.Name.Name != "" {
						osImport = x.Name.Name
					}
				}
			case *ast.FuncDecl:
				if x.Name.Name == "main" {
					mainFunc = true
				} else {
					mainFunc = false
				}
			case *ast.DeferStmt:
				return false
			case *ast.CallExpr:
				if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
					if pfx, ok := sel.X.(*ast.Ident); ok {
						if pfx.Name == osImport && sel.Sel.Name == "Exit" && mainFunc && mainPackage {
							pass.Reportf(x.Pos(), "direct call of os.Exit is not allowed")
						}
					}
				}

			}
			return true
		})
	}
	return nil, nil
}
