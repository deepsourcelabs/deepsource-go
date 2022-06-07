package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/deepsourcelabs/deepsource-go/sdk/types"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

type IssueMeta struct {
	Code             string `json:"code"`
	Text             string `json:"text"`
	ShortDescription string `json:"short_desc"`
	Description      string `json:"desc"`
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

// ParseIssues reads a JSON file containing all issues, and returns all issues.
func ParseIssues(filename string) ([]IssueMeta, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var issues []IssueMeta
	var parsedIssues []IssueMeta

	err = json.Unmarshal(content, &issues)
	if err != nil {
		return nil, err
	}

	for _, issue := range issues {
		// read description from a markdown file
		desc, err := readMarkdown(issue.Description)
		if err != nil {
			return nil, err
		}

		issue.Description = desc
		parsedIssues = append(parsedIssues, issue)
	}

	return parsedIssues, nil
}

// readMarkdown is a helper utility used for parsing a markdown file.
func readMarkdown(filename string) (string, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	// use the Github-flavored Markdown extension
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
	)

	var buf bytes.Buffer
	if err := md.Convert(content, &buf); err != nil {
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
