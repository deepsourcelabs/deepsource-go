package build

import (
	"bytes"
	"strings"
	"testing"

	"github.com/go-test/deep"
)

func TestReadMarkdown(t *testing.T) {
	cases := []struct {
		content  string
		expected string
	}{
		{"# Sample", "<h1>Sample</h1>\n"},
		{"## Sample", "<h2>Sample</h2>\n"},
		{"`Sample`", "<p><code>Sample</code></p>\n"},
		{"```Sample```", "<p><code>Sample</code></p>\n"},
		{"[link](https://example.com)", `<p><a href="https://example.com" rel="nofollow">link</a></p>` + "\n"},
		{"![image](https://sample.org/image.png)", `<p><img src="https://sample.org/image.png" alt="image"></p>` + "\n"},
	}

	for _, tc := range cases {
		actual, err := readMarkdown(tc.content)
		if err != nil {
			t.Error(err)
		}

		if actual != tc.expected {
			t.Errorf("expected: %s, got: %s\n", tc.expected, actual)
		}
	}
}

func TestReadTOML(t *testing.T) {
	tomlNormal := `
[[issues]]

issue_code = "SA4017"
category = "bug-risk"
title = "Sprint is a pure function but its return value is ignored"
description = """
## Description
"""

[[issues]]

issue_code = "S1039"
category = "style"
title = "unnecessary use of fmt.Sprint"
description = """
# Example
"""
`

	expectedTOMLNormal := []IssueMeta{
		{
			IssueCode:   "SA4017",
			Category:    "bug-risk",
			Title:       "Sprint is a pure function but its return value is ignored",
			Description: "## Description\n",
		},
		{
			IssueCode:   "S1039",
			Category:    "style",
			Title:       "unnecessary use of fmt.Sprint",
			Description: "# Example\n",
		},
	}

	tomlBlank := ``

	var expectedTOMLBlank []IssueMeta

	tomlMissingDescription := `
[[issues]]

issue_code = "SA4017"
category = "bug-risk"
title = "Sprint is a pure function but its return value is ignored"
`

	expectedTOMLMissingDescription := []IssueMeta{
		{
			IssueCode: "SA4017",
			Category:  "bug-risk",
			Title:     "Sprint is a pure function but its return value is ignored",
		},
	}

	cases := []struct {
		description string
		tomlContent string
		expected    []IssueMeta
	}{
		{"normal TOML content with issues", tomlNormal, expectedTOMLNormal},
		{"blank TOML", tomlBlank, expectedTOMLBlank},
		{"TOML content with missing descriptions", tomlMissingDescription, expectedTOMLMissingDescription},
	}

	for _, tc := range cases {
		r := strings.NewReader(tc.tomlContent)
		var issue IssueTOML
		err := issue.Read(r)
		if err != nil {
			t.Error(err)
		}

		if diff := deep.Equal(issue.IssueMetas().Issues, tc.expected); diff != nil {
			t.Errorf("description: %s, %s", tc.description, diff)
		}
	}
}

func TestWriteTOML(t *testing.T) {
	// test buffer for writing TOML content
	var testBuffer bytes.Buffer

	expectedTOML := `issue_code = "SA4017"
category = "bug-risk"
title = "Sprint is a pure function but its return value is ignored"
description = "example"` + "\n"

	expectedTOMLMissingDescription :=
		`issue_code = "SA4017"
category = "bug-risk"
title = "Sprint is a pure function but its return value is ignored"
description = ""` + "\n"

	cases := []struct {
		description string
		writer      bytes.Buffer
		issue       IssueMeta
		expected    string
	}{
		{"must write to writer", testBuffer, IssueMeta{
			IssueCode:   "SA4017",
			Category:    "bug-risk",
			Title:       "Sprint is a pure function but its return value is ignored",
			Description: "example",
		}, expectedTOML},
		{"must write to writer in case fields are missing", testBuffer, IssueMeta{
			IssueCode: "SA4017",
			Category:  "bug-risk",
			Title:     "Sprint is a pure function but its return value is ignored",
		}, expectedTOMLMissingDescription},
	}

	for _, tc := range cases {
		err := tc.issue.Write(&tc.writer)
		if err != nil {
			t.Error(err)
		}

		// read content and reset test buffer
		content := tc.writer.String()
		defer tc.writer.Reset()

		if diff := deep.Equal(content, tc.expected); diff != nil {
			t.Errorf("description: %s, %s", tc.description, diff)
		}
	}
}

func TestParseIssues(t *testing.T) {
	cases := []struct {
		description string
		issues      []IssueMeta
		expected    []IssueMeta
	}{
		{"must parse markdown", []IssueMeta{
			{
				IssueCode:   "E001",
				Category:    "bug-risk",
				Title:       "Handle error",
				Description: "## Example",
			},
		}, []IssueMeta{
			{
				IssueCode:   "E001",
				Category:    "bug-risk",
				Title:       "Handle error",
				Description: "<h2>Example</h2>\n",
			},
		}},
		{"must wrap text in paragraph", []IssueMeta{
			{
				IssueCode:   "E001",
				Category:    "bug-risk",
				Title:       "Handle error",
				Description: "Example",
			},
		}, []IssueMeta{
			{
				IssueCode:   "E001",
				Category:    "bug-risk",
				Title:       "Handle error",
				Description: "<p>Example</p>\n",
			},
		}},
	}

	for _, tc := range cases {
		var issueMetas IssueMetas
		issueMetas.Issues = tc.issues
		actual, err := parseIssues(issueMetas)
		if err != nil {
			t.Error(err)
		}

		if diff := deep.Equal(actual.Issues, tc.expected); diff != nil {
			t.Errorf("description: %s, %s", tc.description, diff)
		}
	}
}
