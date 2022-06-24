package utils

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/go-test/deep"
)

func TestWriteTOML(t *testing.T) {
	var testBuf bytes.Buffer

	normalTOML, err := os.ReadFile("testdata/src/issues/U001.toml")
	if err != nil {
		t.Error("failed to read testdata.")
	}

	type test struct {
		description string
		result      map[string]string
		want        string
		expectErr   bool
	}

	tests := []test{
		{description: "empty issue code should return an error", result: map[string]string{"issue_code": ""}, want: "", expectErr: true},
		{description: "normal TOML generation", result: map[string]string{"issue_code": "U001", "title": "Unused variables", "category": "demo", "description": "# some markdown here"}, want: string(normalTOML), expectErr: false},
	}

	for _, tc := range tests {
		err := writeTOML(tc.result, &testBuf)
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

}

func TestCodeGenerator(t *testing.T) {
	exampleContent, err := os.ReadFile("testdata/src/codegen/example.go")
	if err != nil {
		t.Error("failed to read testdata.")
	}

	type test struct {
		description       string
		pluginAnalyzerMap map[string][]string
		want              string
	}

	tests := []test{
		{description: "some plugins", pluginAnalyzerMap: map[string][]string{
			"go-ast": {"hello", "hi"},
		}, want: string(exampleContent)},
	}

	for _, tc := range tests {
		f := codeGenerator(tc.pluginAnalyzerMap)
		got := fmt.Sprintf("%#v", f)

		diffs, result := checkEquality(got, tc.want)

		if !result {
			t.Errorf("description: %s, content doesn't match\n", tc.description)
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
		{description: "walk testdata", directory: "testdata/src/", want: []string{"testdata/src/codegen/example.go", "testdata/src/issues/U001.toml"}},
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
