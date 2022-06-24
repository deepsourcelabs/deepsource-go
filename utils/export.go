package utils

import (
	"errors"
	"io"
	"os"

	"github.com/deepsourcelabs/deepsource-go/types"
	"github.com/pelletier/go-toml/v2"
)

// WriteIssues writes issues extracted from ParseAnnotations to the respective TOML files (issue_code.toml)
func WriteIssues(issues []types.Issue, dir string) error {
	for _, result := range issues {
		if result.IssueCode != "" {
			fname := dir + result.IssueCode + ".toml"

			f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				return err
			}

			err = writeTOML(result, f)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// generateTOMLContent generates the TOML content for an issue using the result map.
func generateTOMLContent(issue types.Issue) ([]byte, error) {
	// only generate content if the issue code is not empty
	if issue.IssueCode != "" {
		content, err := toml.Marshal(issue)
		if err != nil {
			return nil, err
		}

		return content, err
	}

	// return an error if the issue code is empty
	return nil, errors.New("issue code is empty")
}

// writeTOML writes the TOML content for an issue to a TOML file.
func writeTOML(issue types.Issue, w io.Writer) error {
	content, err := generateTOMLContent(issue)
	if err != nil {
		return err
	}

	_, err = w.Write(content)
	if err != nil {
		return err
	}

	return nil
}
