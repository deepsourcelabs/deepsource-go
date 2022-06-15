# Writing a CSS Analyzer

In this example, we will be writing a custom analyzer for [csslint](https://github.com/CSSLint/csslint).

## Pre-requisites

This guide assumes that you are familiar with [writing custom analyzers for DeepSource](writing-analyzers.md).

## Getting Started

We will be using csslint's `--format=compact` for getting results. Since the compact format can be easily parsed through regular expressions, let's use the built-in `RegexProcessor`.

`csslint` doesn't natively support issue codes. Here is an example:

```
/home/testdir/file1.css: line 1, col 1, Warning - Rule is empty. (empty-rules)
/home/testdir/file1.css: line 4, col 2, Warning - Expected (<color>) but found '"blue" size'. (known-properties)
/home/testdir/file1.css: line 5, col 6, Error - Expected RBRACE at line 5, col 6. (errors)
```

In order to fulfill our requirements, let's make use of the `IssueCodeGenerator` provided by `RegexProcessor`.

`IssueCodeGenerator` is used when an analyzer doesn't support issue codes. It takes the content of the `issue_code` named group (from `Pattern`) and returns the issue code.

> **Note**:
>
> If `IssueCodeGenerator` is not implemented, it fallbacks to using the content as the issue code.

```go
func issueCodeGenerator(content string) string {
	issueMap := map[string]string{
		"empty-rules":      "E001",
		"errors":           "E002",
		"known-properties": "K001",
	}

	if issueMap[content] == "" {
		return "U001"
	}

	return issueMap[content]
}
```

Here is the complete code:

```go
package main

import (
	"fmt"
	"log"

	"github.com/deepsourcelabs/deepsource-go/analyzers"
	"github.com/deepsourcelabs/deepsource-go/analyzers/processors"
)

func main() {
	rp := processors.RegexProcessor{
		Pattern:            `(?P<filename>.+): line (?P<line>\d+), col (?P<column>\d+), (?P<message>.+) \((?P<issue_code>.+)\)`,
		IssueCodeGenerator: issueCodeGenerator,
	}

	a := analyzers.CLIRunner{
		Name:      "csslint",
		Command:   "csslint",
		Args:      []string{"--format=compact", "."},
		Processor: &rp,
	}

	err := a.Run()
	if err != nil {
		log.Fatalln(err)
	}

	report, err := a.Processor.Process(a.Stdout())
	if err != nil {
		log.Fatalln(err)
	}

	err = a.SaveReport(report)
	if err != nil {
		log.Fatalln(err)
	}
}

func issueCodeGenerator(content string) string {
	issueMap := map[string]string{
		"empty-rules":      "E001",
		"errors":           "E002",
		"known-properties": "K001",
	}

	if issueMap[content] == "" {
		return "U001"
	}

	return issueMap[content]
}
```

## Running our analyzer

Let's run the analyzer. We must see the report saved as a JSON file (under `$TOOLBOX_PATH/analysis_report.json`):

```json
{
	"issues": [
		{
			"issue_code": "E001",
			"issue_text": "Warning - Rule is empty.",
			"location": {
				"path": "/home/testdir/file1.css",
				"position": {
					"begin": {
						"line": 1,
						"column": 1
					},
					"end": {
						"line": 0,
						"column": 0
					}
				}
			}
		},
		{
			"issue_code": "K001",
			"issue_text": "Warning - Expected (\u003ccolor\u003e) but found '\"blue\" size'.",
			"location": {
				"path": "/home/testdir/file1.css",
				"position": {
					"begin": {
						"line": 4,
						"column": 2
					},
					"end": {
						"line": 0,
						"column": 0
					}
				}
			}
		},
		{
			"issue_code": "E002",
			"issue_text": "Error - Expected RBRACE at line 5, col 6.",
			"location": {
				"path": "/home/testdir/file1.css",
				"position": {
					"begin": {
						"line": 5,
						"column": 6
					},
					"end": {
						"line": 0,
						"column": 0
					}
				}
			}
		}
	],
	"errors": null,
	"extra_data": null
}
```
