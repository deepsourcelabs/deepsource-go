# Writing Custom Analyzers

In this example, we will be writing a custom analyzer for [staticcheck](https://staticcheck.io/).

## How to a write a custom analyzer?

The flow of writing a custom analyzer using the SDK, is as follows:
- Create an analyzer (`CLIRunner`)
- Run the analyzer (`Run()`)
- Use the processor to fetch a DeepSource-compatible report (`Processor.Process()`)
- Persist the report to the local filesystem using `SaveReport`

## Getting Started

### Setting up our analyzer

The analyzer should contain the following:
- `Name`: Name of the analyzer
- `Command`: The main command for the CLI tool (for example, `staticcheck`, etc.)
- `Args`: Arguments for `Command`.
- `Processor`: Processor used for parsing the output of the CLI analyzer

The analyzer can be executed using `Run()`, which executes the CLI (`Command` along with its `Args`). The output of the CLI is stored to its respective buffers. (`stdout` and `stderr`; accessible through `a.Stdout()` and `a.Stderr()`)

Here is the code for the analyzer:

```go
package main

import (
	"fmt"
	"log"

	"github.com/deepsourcelabs/deepsource-go/analyzers"
	"github.com/deepsourcelabs/deepsource-go/analyzers/processors"
)

func main() {
	a := analyzers.CLIRunner{
		Name:      "staticcheck",
		Command:   "staticcheck",
		Args:      []string{"-f", "text", "./..."},
		Processor: &processor, // <=== will be implemented later
	}

	err := a.Run()
	if err != nil {
		log.Fatalln(err)
	}
}
```

### Using processors

A processor is used for converting the result returned by the custom analyzer into a DeepSource-compatible report. The processor must implement `Process()`.

Let's use the built-in `RegexProcessor`. The pattern used by `RegexProcessor` should have the following groups:

- `filename`
- `line`
- `column`
- `message`
- `issue_code`

`RegexProcessor` uses these named groups to populate issues.

```go
package main

import (
	"fmt"
	"log"

	"github.com/deepsourcelabs/deepsource-go/analyzers"
	"github.com/deepsourcelabs/deepsource-go/analyzers/processors"
)

func main() {
	processor := processors.RegexProcessor{
		Pattern: `(?P<filename>.+):(?P<line>\d+):(?P<column>\d+): (?P<message>.+)\((?P<issue_code>\w+)\)`,
	}

	a := analyzers.CLIRunner{
		Name:      "staticcheck",
		Command:   "staticcheck",
		Args:      []string{"-f", "text", "./..."},
		Processor: &processor,
	}

	err := a.Run()
	if err != nil {
		log.Fatalln(err)
	}

	report, err := a.Processor.Process(a.Stdout())
	if err != nil {
		log.Fatalln(err)
	}
}
```

### Saving the report

For persisting the report fetched from our processor, we can use `SaveReport`.

> **Note**:
>
> `SaveReport` requires `TOOLBOX_PATH` to be set in the environment variables.

The report is then saved to `$TOOLBOX_PATH/analysis_report.json`.

```go

	(previous code)

	err = a.SaveReport(report)
	if err != nil {
		log.Fatalln(err)
	}
```

## Running our analyzer

On running the analyzer, we must see the report saved as a JSON file (under `$TOOLBOX_PATH/analysis_report.json`):

```json
{
	"issues": [
		{
			"issue_code": "U1000",
			"issue_text": "func trigger is unused ",
			"location": {
				"path": "analyzers/testdata/src/staticcheck/staticcheck.go",
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
		},
		{
			"issue_code": "SA4017",
			"issue_text": "Sprint is a pure function but its return value is ignored ",
			"location": {
				"path": "analyzers/testdata/src/staticcheck/staticcheck.go",
				"position": {
					"begin": {
						"line": 6,
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
			"issue_code": "S1039",
			"issue_text": "unnecessary use of fmt.Sprint ",
			"location": {
				"path": "analyzers/testdata/src/staticcheck/staticcheck.go",
				"position": {
					"begin": {
						"line": 6,
						"column": 2
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

## Generating TOML files for issues

The developer can define all issues in a single TOML file. This file acts as a single point of truth for generating TOML files for each issue.

This is helpful for developers who wish to define custom issues for their analyzers.

For example, we have `issues.toml` as the file containing details for all issues:

```toml
[[issues]]

issue_code = "SA4017"
category = "bug-risk"
title = "Sprint is a pure function but its return value is ignored"
description = """
Pure functions do not change the passed value but return a new value that is meant to be used. This issue is raised when values returned by pure functions are discarded.
"""
```

> **Note**:
>
> `GenerateTOML` requires `REPO_ROOT` to be set in the environment variables.

`GenerateTOML` reads `$REPO_ROOT/.deepsource/analyzers/issues.toml`, and generates TOML files for each issue, where the filename is the issue code.

The TOML files are generated at `$REPO_ROOT/.deepsource/analyzers/issues/<IssueCode>.toml`.

```go
    // previous code

    // generate TOML files for each issue from a parent TOML file
	err = GenerateTOML()
	if err != nil {
		log.Fatalln(err)
	}
```

On inspecting `$REPO_ROOT/.deepsource/analyzers/issues/SA4017.toml`, we can see the following contents:

```toml
issue_code = "SA4017"
category = "bug-risk"
title = "Sprint is a pure function but its return value is ignored"
description = "Pure functions do not change the passed value but return a new value that is meant to be used. This issue is raised when values returned by pure functions are discarded."
```
