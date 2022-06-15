package build

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"

	"github.com/BurntSushi/toml"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// IssueMeta represents the issue present in a TOML file.
type IssueMeta struct {
	IssueCode   string `toml:"code"`
	Category    string `toml:"category"`
	Title       string `toml:"title"`
	Description string `toml:"description"`
}

// IssueTOML is used for decoding issues from a TOML file.
type IssueTOML struct {
	Issues []map[string]interface{}
}

// GenerateTOML helps in generating TOML files for each issue from a TOML file.
func GenerateTOML() error {
	// root directory for the repository
	repoRoot := os.Getenv("REPO_ROOT")

	// fetch parsed issues
	filename := path.Join(repoRoot, ".deepsource/analyzers/issues.toml")
	issues, err := ParseIssues(filename)
	if err != nil {
		return err
	}

	// generate TOML files
	rootDir := path.Join(repoRoot, ".deepsource/analyzers/issues")
	err = BuildTOML(issues, rootDir)
	if err != nil {
		return err
	}

	return nil
}

// ParseIssues reads a TOML file containing all issues, and returns all issues as []IssueMeta.
func ParseIssues(filename string) ([]IssueMeta, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var issues []IssueMeta
	var parsedIssues []IssueMeta

	var issuesTOML IssueTOML
	err = toml.Unmarshal(content, &issuesTOML)
	if err != nil {
		return nil, err
	}

	for _, issueTOML := range issuesTOML.Issues {
		is := IssueMeta{
			IssueCode:   issueTOML["issue_code"].(string),
			Category:    issueTOML["category"].(string),
			Title:       issueTOML["title"].(string),
			Description: issueTOML["description"].(string),
		}

		issues = append(issues, is)
	}

	for _, issue := range issues {
		// parse and sanitize markdown content
		desc, err := readMarkdown(issue.Description)
		if err != nil {
			return nil, err
		}

		issue.Description = desc
		parsedIssues = append(parsedIssues, issue)
	}

	// sort issues (based on issue code) before returning
	sort.Slice(parsedIssues, func(i, j int) bool {
		return parsedIssues[i].IssueCode < parsedIssues[j].IssueCode
	})

	return parsedIssues, nil
}

// readMarkdown is a helper utility used for parsing and sanitizing markdown content.
func readMarkdown(content string) (string, error) {
	// use the Github-flavored Markdown extension
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(content), &buf); err != nil {
		return "", err
	}

	// sanitize markdown body
	body := buf.String()
	p := bluemonday.UGCPolicy()
	sanitizedBody := p.Sanitize(body)

	return sanitizedBody, nil
}

// BuildTOML uses issues to generate TOML files to a directory.
func BuildTOML(issues []IssueMeta, rootDir string) error {
	if len(issues) == 0 {
		return errors.New("no issues found")
	}

	for _, issue := range issues {
		// The unique identifier (filename) is based on the issue code. TOML files cannot be generated for issues having an invalid/empty code.
		if issue.IssueCode == "" {
			return errors.New("invalid issue code. cannot generate toml")
		}

		// if rootDir doesn't exist, create one
		if _, err := os.Stat(rootDir); err != nil {
			err := os.Mkdir(rootDir, 0700)
			if err != nil {
				return err
			}
		}

		// generate file path based on root directory and filename
		filename := fmt.Sprintf("%s.toml", issue.IssueCode)
		tomlPath := path.Join(rootDir, filename)

		f, err := os.Create(tomlPath)
		if err != nil {
			return err
		}

		if err := toml.NewEncoder(f).Encode(issue); err != nil {
			return err
		}

		if err := f.Close(); err != nil {
			return err
		}
	}

	return nil
}
