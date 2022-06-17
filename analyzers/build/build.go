package build

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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
	IssueCode   string `toml:"issue_code"`
	Category    string `toml:"category"`
	Title       string `toml:"title"`
	Description string `toml:"description"`
}

type IssuesMeta []IssueMeta

// IssueTOML is used for decoding issues from a TOML file.
type IssueTOML struct {
	Issues []map[string]interface{}
}

// GenerateTOML helps in generating TOML files for each issue from a TOML file.
func GenerateTOML() error {
	// root directory for the repository
	repoRoot := os.Getenv("REPO_ROOT")

	filename := path.Join(repoRoot, ".deepsource/analyzers/issues.toml")
	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	// fetch issues
	issues, err := FetchIssues(f)
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

// FetchIssues reads a TOML file containing all issues, and returns all issues as IssuesMeta.
func FetchIssues(r io.Reader) (IssuesMeta, error) {
	issues, err := readTOML(r)
	if err != nil {
		return nil, err
	}

	parsedIssues, err := parseIssues(issues)
	if err != nil {
		return nil, err
	}

	// sort issues (based on issue code) before returning
	sort.Slice(parsedIssues, func(i, j int) bool {
		return parsedIssues[i].IssueCode < parsedIssues[j].IssueCode
	})

	return parsedIssues, nil
}

// BuildTOML uses issues to generate TOML files to a directory.
func BuildTOML(issues IssuesMeta, rootDir string) error {
	if len(issues) == 0 {
		return errors.New("no issues found")
	}

	for _, issue := range issues {
		// The unique identifier (filename) is based on the issue code. TOML files cannot be generated for issues having an invalid/empty code.
		if issue.IssueCode == "" {
			return errors.New("invalid issue code. cannot generate toml")
		}

		// generate file path based on root directory and filename
		filename := fmt.Sprintf("%s.toml", issue.IssueCode)
		tomlPath := path.Join(rootDir, filename)

		f, err := os.Create(tomlPath)
		if err != nil {
			return err
		}

		// write to file
		err = writeTOML(f, issue)
		if err != nil {
			return err
		}
	}

	return nil
}

// readTOML reads content from a reader and returns issues.
func readTOML(r io.Reader) (IssuesMeta, error) {
	// read content from reader
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// unmarshal TOML
	var issuesTOML IssueTOML
	err = toml.Unmarshal(content, &issuesTOML)
	if err != nil {
		return nil, err
	}

	// generate issues
	var issues IssuesMeta
	for _, issueTOML := range issuesTOML.Issues {
		issueCode := ""
		category := ""
		title := ""
		description := ""

		// handle interface conversions
		if issueTOML["IssueCode"] != nil {
			issueCode = issueTOML["IssueCode"].(string)
		}

		if issueTOML["Category"] != nil {
			category = issueTOML["Category"].(string)
		}

		if issueTOML["Title"] != nil {
			title = issueTOML["Title"].(string)
		}

		if issueTOML["Description"] != nil {
			description = issueTOML["Description"].(string)
		}

		is := IssueMeta{
			IssueCode:   issueCode,
			Category:    category,
			Title:       title,
			Description: description,
		}

		issues = append(issues, is)
	}

	return issues, nil
}

// writeTOML writes issue data to the writer.
func writeTOML(w io.Writer, issue IssueMeta) error {
	if err := toml.NewEncoder(w).Encode(issue); err != nil {
		return err
	}

	return nil
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

// parseIssues returns issues after parsing and sanitizing markdown content.
func parseIssues(issues IssuesMeta) (IssuesMeta, error) {
	var parsedIssues IssuesMeta

	for _, issue := range issues {
		// parse and sanitize markdown content
		desc, err := readMarkdown(issue.Description)
		if err != nil {
			return nil, err
		}

		issue.Description = desc
		parsedIssues = append(parsedIssues, issue)
	}

	return parsedIssues, nil
}
