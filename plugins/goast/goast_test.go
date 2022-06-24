package plugins

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"testing"

	"github.com/deepsourcelabs/deepsource-go/plugins/goast/testdata/src/rules"
)

func TestBuildAST(t *testing.T) {
	type test struct {
		description string
		directory   string
		want        []*ast.File
		expectErr   bool
	}

	/////////////////////
	// prepare for tests
	/////////////////////

	// create plugin for test
	p := GoASTPlugin{Name: "go-ast"}

	// read directory and get AST
	var files []*ast.File
	fset := token.NewFileSet()
	content, err := os.ReadFile("testdata/src/trigger/trigger.go")
	if err != nil {
		t.Errorf("failed to read testdata, err: %v\n", err)
	}
	f, err := parser.ParseFile(fset, "", string(content), parser.ParseComments)
	if err != nil {
		t.Errorf("failed to read testdata, err: %v\n", err)
	}
	files = append(files, f)

	/////////////
	// run tests
	/////////////

	tests := []test{
		{description: "must generate ASTs", directory: "testdata/src/trigger", want: files, expectErr: false},
		{description: "must return for invalid directory", directory: "testdata/src/doesnotexist", want: nil, expectErr: true},
	}

	for _, tc := range tests {
		got, err := p.BuildAST(tc.directory)
		if err != nil && !tc.expectErr {
			t.Error(err)
		}

		if !reflect.DeepEqual(got, tc.want) {
			t.Error("ASTs don't match")
		}
	}
}

func TestRun(t *testing.T) {
	type test struct {
		description string
		directory   string
		expectErr   bool
	}

	/////////////////////
	// prepare for tests
	/////////////////////

	// create plugin for test
	p := GoASTPlugin{Name: "go-ast"}
	p.RegisterRule(rules.NeedFunc)

	/////////////
	// run tests
	/////////////

	tests := []test{
		{description: "must generate ASTs", directory: "testdata/src/trigger", expectErr: false},
		{description: "must return error for invalid directory", directory: "testdata/src/doesnotexist", expectErr: true},
	}

	for _, tc := range tests {
		files, err := p.BuildAST(tc.directory)
		if err != nil && !tc.expectErr {
			t.Error(err)
		}

		err = p.Run(files)
		if err != nil && !tc.expectErr {
			t.Error(err)
		}
	}
}
