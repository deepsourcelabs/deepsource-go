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

type IssueMetas struct {
	Issues []IssueMeta
}

// IssueTOML is used for decoding issues from a TOML file.
type IssueTOML struct {
	Issues []map[string]interface{}
}

// GenerateTOML helps in generating TOML files for each issue from a TOML file.
func GenerateTOML(repoRoot string) error {
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
	err = issues.BuildTOML(rootDir)
	if err != nil {
		return err
	}

	return nil
}

// FetchIssues reads a TOML file containing all issues, and returns all issues as IssueMetas.
func FetchIssues(r io.Reader) (IssueMetas, error) {
	// get issues from TOML file
	var issueTOML IssueTOML
	err := issueTOML.Read(r)
	if err != nil {
		return IssueMetas{}, err
	}
	issues := issueTOML.IssueMetas()

	// parse issues
	parsedIssues, err := parseIssues(issues)
	if err != nil {
		return IssueMetas{}, err
	}

	// sort issues (based on issue code) before returning
	sort.Slice(parsedIssues.Issues, func(i, j int) bool {
		return parsedIssues.Issues[i].IssueCode < parsedIssues.Issues[j].IssueCode
	})

	return parsedIssues, nil
}

// BuildTOML uses issues to generate TOML files to a directory.
func (i *IssueMetas) BuildTOML(rootDir string) error {
	if len(i.Issues) == 0 {
		return errors.New("no issues found")
	}

	for _, issue := range i.Issues {
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
		err = issue.Write(f)
		if err != nil {
			return err
		}
	}

	return nil
}

// Read reads content from a reader and unmarshals it to IssueTOML.
func (i *IssueTOML) Read(r io.Reader) error {
	// read content from reader
	content, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	// unmarshal TOML
	err = toml.Unmarshal(content, &i)
	if err != nil {
		return err
	}

	return nil
}

// IssueMetas returns issues from a IssueTOML struct.
func (i *IssueTOML) IssueMetas() IssueMetas {
	var issueMetas IssueMetas
	for _, issueTOML := range i.Issues {
		issueCode := ""
		category := ""
		title := ""
		description := ""

		if issueTOML["issue_code"] != nil {
			issueCode = fmt.Sprintf("%v", issueTOML["issue_code"])
		}

		if issueTOML["category"] != nil {
			category = fmt.Sprintf("%v", issueTOML["category"])
		}

		if issueTOML["title"] != nil {
			title = fmt.Sprintf("%v", issueTOML["title"])
		}

		if issueTOML["description"] != nil {
			description = fmt.Sprintf("%v", issueTOML["description"])
		}

		issueMeta := IssueMeta{
			IssueCode:   issueCode,
			Category:    category,
			Title:       title,
			Description: description,
		}

		issueMetas.Issues = append(issueMetas.Issues, issueMeta)
	}

	return issueMetas
}

// Write writes the issue data to the writer.
func (i *IssueMeta) Write(w io.Writer) error {
	if err := toml.NewEncoder(w).Encode(i); err != nil {
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
func parseIssues(issues IssueMetas) (IssueMetas, error) {
	var parsedIssues IssueMetas

	for _, issue := range issues.Issues {
		// parse and sanitize markdown content
		desc, err := readMarkdown(issue.Description)
		if err != nil {
			return IssueMetas{}, err
		}

		issue.Description = desc
		parsedIssues.Issues = append(parsedIssues.Issues, issue)
	}

	return parsedIssues, nil
}
