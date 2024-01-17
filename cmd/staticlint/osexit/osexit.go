package osexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer configuration
var Analyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "checks of calling os.Exit in main package main func",
	Run:  run,
}

// run runs the analyzer
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			ast.Inspect(file, func(node ast.Node) bool {
				switch x := node.(type) {
				case *ast.FuncDecl:
					if x.Name.String() != "main" {
						return false
					}
				case *ast.CallExpr:
					if selexpr, ok := x.Fun.(*ast.SelectorExpr); ok {
						if ident, ok := selexpr.X.(*ast.Ident); ok {
							if ident.Name == "os" && selexpr.Sel.Name == "Exit" {
								pass.Reportf(ident.NamePos, "calling os.Exit in main package main func")
							}
						}
					}
				}

				return true
			})
		}
	}
	return nil, nil
}
