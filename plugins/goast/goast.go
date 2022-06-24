package plugins

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/deepsourcelabs/deepsource-go/types"
)

// GoASTRuleType defines the signature of a rule for the go/ast plugin.
type GoASTRuleType func(n ast.Node) ([]types.Diagnostic, error)

type GoASTPlugin struct {
	Name        string
	rules       []GoASTRuleType
	Diagnostics []types.Diagnostic
}

// String returns the string representation of the plugin
func (p *GoASTPlugin) String() string {
	return p.Name
}

// BuildAST generates the AST by parsing source code from the directory.
func (*GoASTPlugin) BuildAST(directory string) ([]*ast.File, error) {
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
func (p *GoASTPlugin) Run(files []*ast.File) error {
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

// RegisterRule register a rule for the go/ast plugin.
func (p *GoASTPlugin) RegisterRule(rule GoASTRuleType) {
	p.rules = append(p.rules, rule)
}
