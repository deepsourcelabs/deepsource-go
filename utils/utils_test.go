package utils

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deepsourcelabs/deepsource-go/types"
	"github.com/go-test/deep"
)

func TestWriteTOML(t *testing.T) {
	var testBuf bytes.Buffer

	normalTOML, err := os.ReadFile("testdata/src/issues/U001.toml")
	if err != nil {
		t.Errorf("failed to read testdata, err: %v\n", err)
	}

	type test struct {
		description string
		issue       types.Issue
		want        string
		expectErr   bool
	}

	tests := []test{
		{description: "empty issue code should return an error", issue: types.Issue{IssueCode: ""}, want: "", expectErr: true},
		{description: "normal TOML generation", issue: types.Issue{IssueCode: "U001", Category: "demo", Title: "Unused variables", Description: "# some markdown here"}, want: string(normalTOML), expectErr: false},
	}

	for _, tc := range tests {
		err := writeTOML(tc.issue, &testBuf)
		defer testBuf.Reset()
		if err != nil && !tc.expectErr {
			t.Errorf("description: %s, expected error.\n", tc.description)
		}

		got := testBuf.String()

		if got != tc.want {
			t.Errorf("description: %s, got: %v, want: %v\n", tc.description, got, tc.want)
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
		want        []types.Issue
	}

	tests := []test{
		{description: "empty issue code should return nil", content: string(emptyIssueCodeAnnotation), want: nil},
		{description: "no annotation should return nil", content: string(emptyAnnotation), want: nil},
		{description: "invalid annotation should return nil", content: string(invalidAnnotation), want: nil},
		{description: "multiple annotations should be parsed correctly", content: string(multipleAnnotation), want: []types.Issue{{IssueCode: "NU001", Category: "style", Title: "notused", Description: "## markdown"}, {IssueCode: "E001", Category: "bug-risk", Title: "handle error", Description: "## markdown"}}},
		{description: "single annotation should be parsed correctly", content: string(singleAnnotation), want: []types.Issue{{IssueCode: "EX01", Category: "example", Title: "Some random rule.", Description: "## markdown"}}},
		{description: "annotations with single line comments should also be parsed correctly", content: string(singleLineCommentAnnotation), want: []types.Issue{{IssueCode: "P001", Category: "performance", Title: "Multiple appends can be combined into a single statement", Description: "## markdown"}}},
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

		diffs := deep.Equal(got, tc.want)
		if len(diffs) != 0 {
			t.Errorf("description: %s, issues don't match\n", tc.description)
			t.Log("differences in diffs:", diffs)
		}
	}
}

func TestCodeGenerator(t *testing.T) {
	exampleContent, err := os.ReadFile("testdata/src/codegen/generated.go")
	if err != nil {
		t.Errorf("failed to read testdata, err: %v\n", err)
	}

	type test struct {
		description       string
		pluginAnalyzerMap map[string][]string
		want              string
	}

	tests := []test{
		{description: "go/ast plugin with rules", pluginAnalyzerMap: map[string][]string{
			"go-ast": {"hello", "hi"},
		}, want: string(exampleContent)},
	}

	for _, tc := range tests {
		f := codeGenerator(tc.pluginAnalyzerMap)
		got := fmt.Sprintf("%#v", f)

		diffs, equal := checkEquality(got, tc.want)

		if !equal {
			t.Errorf("description: %s, content doesn't match\n", tc.description)
			t.Log("differences in diffs:", diffs)
		}
	}
}

func TestParseAnnotation(t *testing.T) {
	type test struct {
		description string
		directory   string
		want        []types.Issue
	}

	// set temporary directory for code generation
	tempDir := os.TempDir()
	codegenPath := filepath.Join(tempDir, "generated-test.go")

	tests := []test{
		{description: "no annotation should return nil", directory: "testdata/src/annotations/empty.go", want: nil},
		{description: "multiple annotations should be parsed correctly", directory: "testdata/src/annotations/multiple.go", want: []types.Issue{{IssueCode: "NU001", Category: "style", Title: "notused", Description: "## markdown"}, {IssueCode: "E001", Category: "bug-risk", Title: "handle error", Description: "## markdown"}}},
	}

	for _, tc := range tests {
		got, err := ParseAnnotations(tc.directory, codegenPath)
		if err != nil {
			t.Error(err)
		}

		// cleanup
		defer os.Remove(codegenPath)

		diffs := deep.Equal(got, tc.want)
		if len(diffs) != 0 {
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
	}

	tests := []test{
		{description: "walk testdata", directory: "testdata/src/annotations", want: []string{"testdata/src/annotations/empty.go", "testdata/src/annotations/empty_issuecode.go", "testdata/src/annotations/invalid.go", "testdata/src/annotations/multiple.go", "testdata/src/annotations/single.go", "testdata/src/annotations/singleline_comment.go"}},
	}

	for _, tc := range tests {
		got, err := walkDir(tc.directory)
		if err != nil {
			t.Error(err)
		}

		diffs := deep.Equal(got, tc.want)
		if len(diffs) != 0 {
			t.Errorf("description: %s, file names don't match\n", tc.description)
			t.Log("differences in diffs:", diffs)
		}
	}
}

// checkEquality is a helper for checking string differences. Handles indentation differences, etc.
func checkEquality(got, want string) ([]string, bool) {
	var gotLines []string
	for _, line := range strings.Split(got, "\n") {
		trimmed := strings.TrimSpace(line)
		gotLines = append(gotLines, trimmed)
	}

	var wantLines []string
	for _, line := range strings.Split(want, "\n") {
		trimmed := strings.TrimSpace(line)
		wantLines = append(wantLines, trimmed)
	}

	diff := deep.Equal(gotLines, wantLines)
	return diff, len(diff) == 0
}
