package generator

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseAnnotations(t *testing.T) {
	type test struct {
		description string
		directory   string
		want        Issues
	}

	// set temporary directory for code generation
	tempDir := os.TempDir()
	codegenPath := filepath.Join(tempDir, "generated-test.go")

	tests := []test{
		{description: "no annotation should return nil", directory: "testdata/src/annotations/empty.go", want: nil},
		{description: "multiple annotations should be parsed correctly", directory: "testdata/src/annotations/multiple.go", want: Issues{{IssueCode: "NU001", Category: "style", Title: "notused", Description: "## markdown"}, {IssueCode: "E001", Category: "bug-risk", Title: "handle error", Description: "## markdown"}}},
	}

	for _, tc := range tests {
		got, err := ParseAnnotations(tc.directory, codegenPath)
		if err != nil {
			t.Error(err)
		}

		// cleanup
		defer os.Remove(codegenPath)

		if diffs := cmp.Diff(got, tc.want); diffs != "" {
			t.Errorf("description: %s, issues don't match\n", tc.description)
			t.Log("differences in diffs:", diffs)
		}
	}
}

func TestWalkDir(t *testing.T) {
	type test struct {
		description string
		directory   string
		want        []string
		expectErr   bool
	}

	tests := []test{
		{description: "must walk on valid directory", directory: "testdata/src/annotations", want: []string{"testdata/src/annotations/empty.go", "testdata/src/annotations/empty_issuecode.go", "testdata/src/annotations/invalid.go", "testdata/src/annotations/multiple.go", "testdata/src/annotations/single.go", "testdata/src/annotations/singleline_comment.go"}, expectErr: false},
		{description: "must return nil for invalid directory", directory: "testdata/src/doesnotexist", want: nil, expectErr: true},
	}

	for _, tc := range tests {
		got, err := walkDir(tc.directory)
		if err != nil && !tc.expectErr {
			t.Error(err)
		}

		if diffs := cmp.Diff(got, tc.want); diffs != "" {
			t.Errorf("description: %s, file names don't match\n", tc.description)
			t.Log("differences in diffs:", diffs)
		}
	}
}

func TestTraverseAST(t *testing.T) {
	////////////////
	// read testdata
	////////////////
	emptyIssueCodeAnnotation, err := os.ReadFile("testdata/src/annotations/empty_issuecode.go")
	if err != nil {
		t.Errorf("failed to read testdata, err: %v\n", err)
	}

	emptyAnnotation, err := os.ReadFile("testdata/src/annotations/empty.go")
	if err != nil {
		t.Errorf("failed to read testdata, err: %v\n", err)
	}

	invalidAnnotation, err := os.ReadFile("testdata/src/annotations/invalid.go")
	if err != nil {
		t.Errorf("failed to read testdata, err: %v\n", err)
	}

	multipleAnnotation, err := os.ReadFile("testdata/src/annotations/multiple.go")
	if err != nil {
		t.Errorf("failed to read testdata, err: %v\n", err)
	}

	singleAnnotation, err := os.ReadFile("testdata/src/annotations/single.go")
	if err != nil {
		t.Errorf("failed to read testdata, err: %v\n", err)
	}

	singleLineCommentAnnotation, err := os.ReadFile("testdata/src/annotations/singleline_comment.go")
	if err != nil {
		t.Errorf("failed to read testdata, err: %v\n", err)
	}

	/////////////
	// run tests
	/////////////

	type test struct {
		description string
		content     string
		want        Issues
	}

	tests := []test{
		{description: "empty issue code should return nil", content: string(emptyIssueCodeAnnotation), want: nil},
		{description: "no annotation should return nil", content: string(emptyAnnotation), want: nil},
		{description: "invalid annotation should return nil", content: string(invalidAnnotation), want: nil},
		{description: "multiple annotations should be parsed correctly", content: string(multipleAnnotation), want: Issues{{IssueCode: "NU001", Category: "style", Title: "notused", Description: "## markdown"}, {IssueCode: "E001", Category: "bug-risk", Title: "handle error", Description: "## markdown"}}},
		{description: "single annotation should be parsed correctly", content: string(singleAnnotation), want: Issues{{IssueCode: "EX01", Category: "example", Title: "Some random rule.", Description: "## markdown"}}},
		{description: "annotations with single line comments should also be parsed correctly", content: string(singleLineCommentAnnotation), want: Issues{{IssueCode: "P001", Category: "performance", Title: "Multiple appends can be combined into a single statement", Description: "## markdown"}}},
	}

	for _, tc := range tests {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "", tc.content, parser.ParseComments)
		if err != nil {
			t.Error(err)
		}

		got, err := traverseAST(f)
		if err != nil {
			t.Error(err)
		}

		if diffs := cmp.Diff(got, tc.want); diffs != "" {
			t.Errorf("description: %s, issues don't match\n", tc.description)
			t.Log("differences in diffs:", diffs)
		}
	}
}
