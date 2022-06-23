package plugins

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/deepsourcelabs/deepsource-go/types"
)

type GoASTRuleType func(n ast.Node) ([]types.Diagnostic, error)

type GoASTPlugin struct {
	Name  string
	rules []GoASTRuleType
}

func (p *GoASTPlugin) String() string {
	return p.Name
}

func (p *GoASTPlugin) BuildAST(dir string) ([]*ast.File, error) {
	var files []*ast.File

	fset := token.NewFileSet()
	err := filepath.WalkDir(dir, func(path string, fileInfo fs.DirEntry, err error) error {
		if !fileInfo.IsDir() {
			content, err := os.ReadFile(filepath.Join(dir, fileInfo.Name()))
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

func (p *GoASTPlugin) Run(files []*ast.File, strict bool) error {
	for _, file := range files {
		ast.Inspect(file, func(n ast.Node) bool {
			for _, rule := range p.rules {
				diag, err := rule(n)
				if err != nil {
					return false
				}

				if len(diag) != 0 {
					log.Println("diagnostic:", diag)
				}
			}

			return true
		})
	}
	return nil
}

func (p *GoASTPlugin) RegisterRule(rule GoASTRuleType) {
	p.rules = append(p.rules, rule)
}
