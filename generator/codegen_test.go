package generator

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCodeGenerator(t *testing.T) {
	exampleContent, err := os.ReadFile("testdata/src/codegen/generated.go")
	if err != nil {
		t.Errorf("failed to read testdata, err: %v\n", err)
	}

	type test struct {
		description     string
		analyzerRuleMap map[string][]string
		want            string
	}

	tests := []test{
		{description: "go/ast analyzer with rules", analyzerRuleMap: map[string][]string{
			"go-ast": {"NeedFunc"},
		}, want: string(exampleContent)},
	}

	for _, tc := range tests {
		f := codeGenerator(tc.analyzerRuleMap)
		got := fmt.Sprintf("%#v", f)

		diffs, equal := checkEquality(got, tc.want)

		if !equal {
			t.Errorf("description: %s, content doesn't match\n", tc.description)
			t.Log("differences in diffs:", diffs)
		}
	}
}

// checkEquality is a helper for checking string differences. Handles indentation differences, etc.
func checkEquality(got, want string) (string, bool) {
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

	diffs := cmp.Diff(gotLines, wantLines)
	return diffs, diffs == ""
}
