package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/deepsourcelabs/deepsource-go/analyzers/types"
	"github.com/microcosm-cc/bluemonday"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// IssueMeta represents the issue present in a TOML file.
type IssueMeta struct {
	Code             string `json:"code"`
	Text             string `json:"text"`
	ShortDescription string `json:"short_desc"`
	Description      string `json:"desc"`
}

// IssueTOML is used for decoding issues from a TOML file.
type IssueTOML struct {
	Issue []map[string]interface{}
}

// SaveReport saves the analysis report to the local filesystem.
func SaveReport(report types.AnalysisReport, filename string, exportType string) error {
	var err error

	switch exportType {
	case "json":
		err = exportJSON(report, filename)
	default:
		return errors.New("export type not supported. supported types include: json")
	}

	return err
}

// exportJSON is a helper utility for saving the analysis report in a JSON format.
func exportJSON(report types.AnalysisReport, filename string) error {
	data, err := json.MarshalIndent(report, "", "	")
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(filename, data, 0644); err != nil {
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

	for _, issueTOML := range issuesTOML.Issue {
		is := IssueMeta{
			Code:             issueTOML["Code"].(string),
			Text:             issueTOML["Text"].(string),
			ShortDescription: issueTOML["ShortDescription"].(string),
			Description:      issueTOML["Description"].(string),
		}

		issues = append(issues, is)
	}

	for _, issue := range issues {
		// parse markdown content
		desc, err := readMarkdown(issue.Description)
		if err != nil {
			return nil, err
		}

		issue.Description = desc
		parsedIssues = append(parsedIssues, issue)
	}

	// sort issues (based on issue code) before returning
	sort.Slice(parsedIssues, func(i, j int) bool {
		return parsedIssues[i].Code < parsedIssues[j].Code
	})

	return parsedIssues, nil
}

// readMarkdown is a helper utility used for parsing markdown content.
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
	for _, issue := range issues {
		// The unique identifier (filename) is based on the issue code. TOML files cannot be generated for issues having an invalid/empty code.
		if issue.Code == "" {
			return errors.New("invalid issue code. cannot generate toml")
		}

		// if rootDir doesn't exist, create one
		if _, err := os.Stat(rootDir); err != nil {
			err = os.Mkdir(rootDir, 0700)
			if err != nil {
				return err
			}
		}

		// generate file path based on root directory and filename
		filename := fmt.Sprintf("%s.toml", issue.Code)
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

// ParsedIssue represents an issue parsed using tree-sitter.
type ParsedIssue struct {
	IssueCode string
	Line      int
}

// Verify compares the generated report and parsed issues using tree-sitter.
func Verify(report types.AnalysisReport, filename string) error {
	parser := sitter.NewParser()

	// get language
	lang, err := getLanguage(filename)
	if err != nil {
		return err
	}
	parser.SetLanguage(lang)

	// read report
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// generate tree
	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, content)
	if err != nil {
		return err
	}

	// create a query for fetching comments
	queryStr := "(comment) @comment"
	query, err := sitter.NewQuery([]byte(queryStr), lang)
	if err != nil {
		return err
	}

	// execute query on root node
	qc := sitter.NewQueryCursor()
	n := tree.RootNode()
	qc.Exec(query, n)
	defer qc.Close()

	var parsedIssues []ParsedIssue

	// iterate over matches
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}

		for _, c := range m.Captures {
			// get node content
			node := c.Node
			nodeContent := node.Content(content)

			// check if the comment contains raise annotation
			if strings.Contains(nodeContent, "raise") {
				// find match using expression
				exp := regexp.MustCompile(`.+ raise: `)
				submatches := exp.FindStringSubmatch(nodeContent)

				if len(submatches) != 0 {
					substrings := exp.Split(nodeContent, -1)
					if len(substrings) > 1 {
						issueCodes := strings.Split(substrings[1], ",")
						// add issue to parsedIssues
						for _, issueCode := range issueCodes {
							parsedIssue := ParsedIssue{IssueCode: strings.TrimSpace(issueCode), Line: int(node.StartPoint().Row) + 1}
							parsedIssues = append(parsedIssues, parsedIssue)
						}
					}
				}
			}
		}
	}

	// if number of issues don't match, exit early.
	if len(parsedIssues) != len(report.Issues) {
		return errors.New("mismatch between the number of reported issues and parsed issues")
	}

	// compare the report's issues and parsed issues
	match := compareReport(parsedIssues, report)
	if !match {
		return errors.New("mismatch between parsed issue and report issue")
	}

	return nil
}

// getLanguage is a helper for fetching a tree-sitter language based on the file's extension.
func getLanguage(filename string) (*sitter.Language, error) {
	extension := filepath.Ext(filename)

	switch extension {
	case ".go":
		return golang.GetLanguage(), nil
	default:
		return nil, errors.New("language not supported")
	}
}

// compareReport is a helper which checks if the parsed issues are identical to the issues present in the report.
func compareReport(parsedIssues []ParsedIssue, report types.AnalysisReport) bool {
	// sort report and parsedIssues by line number
	sort.Slice(parsedIssues, func(i, j int) bool {
		return parsedIssues[i].Line < parsedIssues[j].Line
	})

	sort.Slice(report.Issues, func(i, j int) bool {
		return report.Issues[i].Location.Position.Begin.Line < report.Issues[j].Location.Position.Begin.Line
	})

	for i, issue := range report.Issues {
		if (parsedIssues[i].Line != issue.Location.Position.Begin.Line) && (parsedIssues[i].IssueCode != issue.IssueCode) {
			return false
		}
	}

	return true
}
