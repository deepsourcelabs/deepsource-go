package utils

import (
	"bytes"
	"os"
	"testing"
)

func TestWriteTOML(t *testing.T) {
	var testBuf bytes.Buffer

	normalTOML, err := os.ReadFile("testdata/U001.toml")
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
