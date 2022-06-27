package generator

import (
	"bytes"
	"os"
	"testing"
)

func TestWriteTOML(t *testing.T) {
	var testBuf bytes.Buffer

	normalTOML, err := os.ReadFile("testdata/src/issues/U001.toml")
	if err != nil {
		t.Errorf("failed to read testdata, err: %v\n", err)
	}

	type test struct {
		description string
		issue       Issue
		want        string
		expectErr   bool
	}

	tests := []test{
		{description: "empty issue code should return an error", issue: Issue{IssueCode: ""}, want: "", expectErr: true},
		{description: "normal TOML generation", issue: Issue{IssueCode: "U001", Category: "demo", Title: "Unused variables", Description: "# some markdown here"}, want: string(normalTOML), expectErr: false},
	}

	for _, tc := range tests {
		err := tc.issue.writeTOML(&testBuf)
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
