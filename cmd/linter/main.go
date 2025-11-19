package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

const (
	// Название модуля, в котором можно вызывать os.Exit() и log.Fatal()
	AcceptablePkgName = "main"
	// Название функции, в которой можно вызывать os.Exit() и log.Fatal()
	AcceptableFuncName = "main"
)

// Анализатор проверяющий вызов panic(), os.Exit() и log.Fatal().
// os.Exit() и log.Fatal() разрешено вызывать в функции main() пакета main.
var FatalCheckAnalyzer = &analysis.Analyzer{
	Name: "fatalcheck",
	Doc:  "check for panic(), os.Exit(), log.Fatal() acceptable use",
	Run:  run,
}

func main() {
	singlechecker.Main(FatalCheckAnalyzer)
}

func run(pass *analysis.Pass) (any, error) {
	checker := &fatalChecker{
		stack: make([]ast.Node, 0, 16),
	}

	return checker.run(pass)
}

type fatalChecker struct {
	stack   []ast.Node
	pkgName string
}

func (c *fatalChecker) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		c.pkgName = file.Name.Name

		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.CallExpr:
				c.checkFuncCall(x, pass)
			}

			if node == nil {
				c.stack = c.stack[:len(c.stack)-1]
			} else {
				c.stack = append(c.stack, node)
			}
			return true
		})
	}

	return nil, nil
}

func (c *fatalChecker) checkFuncCall(x *ast.CallExpr, p *analysis.Pass) {
	switch fn := x.Fun.(type) {
	case *ast.Ident:
		if fn.Name == "panic" {
			p.Reportf(x.Pos(), "unacceptable panic call")
		}
	case *ast.SelectorExpr:
		pkg, ok := fn.X.(*ast.Ident)
		if ok {
			switch {
			case pkg.Name == "log" && fn.Sel.Name == "Fatal" && !c.isAcceptableExit():
				p.Reportf(x.Pos(), "unacceptable log.Fatal call")
			case pkg.Name == "os" && fn.Sel.Name == "Exit" && !c.isAcceptableExit():
				p.Reportf(x.Pos(), "unacceptable os.Exit call")
			}
		}
	}
}

func (c *fatalChecker) isAcceptableExit() bool {
	if c.pkgName != AcceptablePkgName {
		return false
	}
	for _, n := range c.stack {
		fn, ok := n.(*ast.FuncDecl)
		if ok {
			return fn.Name.Name == AcceptableFuncName
		}
	}
	return false
}
