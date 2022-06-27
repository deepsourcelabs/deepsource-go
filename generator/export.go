package generator

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// WriteIssues writes issues extracted from ParseAnnotations to the respective TOML files (issue_code.toml)
func WriteIssues(issues []Issue, dir string) error {
	for _, issue := range issues {
		if issue.IssueCode != "" {
			fname := fmt.Sprintf("%s.toml", issue.IssueCode)
			fpath := filepath.Join(dir, fname)

			f, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				return err
			}

			err = writeTOML(issue, f)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// generateTOMLContent generates the TOML content for an issue.
func generateTOMLContent(issue Issue) ([]byte, error) {
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
func writeTOML(issue Issue, w io.Writer) error {
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
