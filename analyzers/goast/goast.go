package analyzers

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/deepsourcelabs/deepsource-go/analyzers"
)

// GoASTRuleType defines the signature of a rule for the go/ast analyzer.
type GoASTRuleType func(n ast.Node) ([]analyzers.Diagnostic, error)

type GoASTAnalyzer struct {
	Name        string
	rules       []GoASTRuleType
	Diagnostics []analyzers.Diagnostic
}

// String returns the string representation of the analyzer
func (p *GoASTAnalyzer) String() string {
	return p.Name
}

// buildAST generates the AST by parsing source code from the directory.
func buildAST(directory string) ([]*ast.File, error) {
	var files []*ast.File

	fset := token.NewFileSet()

	// check if directory exists before walking
	if _, err := os.Stat(directory); err != nil {
		return nil, err
	}

	err := filepath.WalkDir(directory, func(path string, fileInfo fs.DirEntry, err error) error {
		if !fileInfo.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			f, err := parser.ParseFile(fset, "", string(content), parser.ParseComments)
			if err != nil {
				return err
			}

			files = append(files, f)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

// Run uses the AST to apply rules and record diagnostics.
func (p *GoASTAnalyzer) Run(directory string) error {
	files, err := buildAST(directory)
	if err != nil {
		return err
	}

	for _, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			for _, rule := range p.rules {
				diagnostic, err := rule(node)
				if err != nil {
					return false
				}
				p.Diagnostics = append(p.Diagnostics, diagnostic...)
			}

			return true
		})
	}

	return nil
}

// RegisterRule registers a rule for the go/ast analyzer.
func (p *GoASTAnalyzer) RegisterRule(rule GoASTRuleType) {
	p.rules = append(p.rules, rule)
}
