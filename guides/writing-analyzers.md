# Writing Custom Analyzers

In this example, we will be writing a custom analyzer for [staticcheck](https://staticcheck.io/).

## How to a write a custom analyzer?

The flow of writing a custom analyzer using the SDK, is as follows:
- Create an analyzer (`CLIRunner`)
- Run the analyzer (`Run()`)
- Use the processor to fetch a DeepSource-compatible report (`Processor.Process()`)
- Persist the report to the local filesystem using `SaveReport`

## Setting up

Let's add this to our `main.go`:

```go
package main

import (
	"log"

	"github.com/deepsourcelabs/deepsource-go/analyzers"
	"github.com/deepsourcelabs/deepsource-go/analyzers/utils"
)

func main() {
    // create a CLI analyzer
	a := analyzers.CLIRunner{
		Name:      "staticcheck", // name of the analyzer
		Command:   "staticcheck", // main command
		Args:      []string{"-f", "text", "./..."}, // args
		Processor: &StaticCheckProcessor{}, // processor
	}

    // run the analyzer
	err := a.Run()
	if err != nil {
		log.Fatalln(err)
	}

    // process the output from staticcheck using the stdout stream
	report, err := a.Processor.Process(a.Stdout())
	if err != nil {
		log.Fatalln(err)
	}

    // save report to a JSON file
	err = utils.SaveReport(report, "issues.json", "json")
	if err != nil {
		log.Fatalln(err)
	}
}
```

## Implementing our custom processor

A processor is used for converting the result returned by the custom analyzer into a DeepSource-compatible report. The processor must implement `Process()`.

If the analyzer's output format is common in nature (unix-style, etc.), the SDK provides pre-built processors for usage.

Since `staticcheck`'s output format is not common in nature, we need to implement the processor for our `staticcheck` analyzer.

```go
type StaticCheckProcessor struct{}

// StaticCheck processor returns a DeepSource-compatible analysis report from staticcheck's results.
func (s *StaticCheckProcessor) Process(buf bytes.Buffer) (types.AnalysisReport, error) {
	var issues []types.Issue

	// trim newline from buffer output
	lines := strings.Split(buf.String(), "\n")

	for _, line := range lines {
		// trim spaces
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		// compile regular expression for parsing unix format

		// group descriptions:
		// 0: complete string
		// 1: path
		// 2: line number
		// 3: column number
		// 4: message
		exp, err := regexp.Compile("(.+):(.):(.): (.+)")
		if err != nil {
			return types.AnalysisReport{}, err
		}

		// get groups
		groups := exp.FindAllStringSubmatch(strings.TrimSuffix(line, "\n"), -1)
		if len(groups) == 0 {
			return types.AnalysisReport{}, errors.New("failed to parse output string")
		}

		// convert line and column numbers to int
		line, err := strconv.Atoi(groups[0][2])
		if err != nil {
			return types.AnalysisReport{}, err
		}

		col, err := strconv.Atoi(groups[0][3])
		if err != nil {
			return types.AnalysisReport{}, err
		}

		// compile regular expression for parsing staticcheck message

		// group descriptions:
		// 0: complete string
		// 1: partial message string
		// 2: issue code
		// 3: parentheses
		messageExp, err := regexp.Compile("(.+)[(](.+)(.+)")
		if err != nil {
			return types.AnalysisReport{}, err
		}
		messageGroups := messageExp.FindAllStringSubmatch(groups[0][4], -1)
		if len(messageGroups) == 0 {
			return types.AnalysisReport{}, errors.New("failed to parse message")
		}

		// populate issue
		issue := types.Issue{
			IssueCode: messageGroups[0][2],
			IssueText: groups[0][4],
			Location: types.Location{
				Path: groups[0][1],
				Position: types.Position{
					Begin: types.Coordinate{
						Line:   line,
						Column: col,
					},
				},
			},
		}

		issues = append(issues, issue)
	}

	// populate report
	report := types.AnalysisReport{
		Issues: issues,
	}

	// return report
	return report, nil
}
```

## Running our analyzer

Wow! We just implemented our own custom analyzer!

On running the analyzer, we must see the report saved as a JSON file:

```json
[
    {
        "code": "SA4017",
        "text": "Sprint is a pure function but its return value is ignored",
        "short_desc": "Sprint is a pure function but its return value is ignored",
        "desc": "/home/aadhav/analyzer-go-sdk/playground/sa4017.md"
    }
]
```

## Generating TOML files for issues

The developer can define all issues in a single TOML file. This file acts as a single point of truth for generating TOML files for each issue.

This is helpful for developers who wish to define custom issues for their analyzers.

For example, we have `issues.toml` as the file containing details for all issues:

```toml
[[issue]]

Code = "SA4017"
Text = "Sprint is a pure function but its return value is ignored"
ShortDescription = "Sprint is a pure function but its return value is ignored"
Description = """
## Sample
"""
```

`GenerateTOML` reads `issues.toml`, and generates TOML files for each issue, where the filename is the issue code:

```go
    // previous code

    // generate TOML files for each issue from a parent TOML file
	err = a.GenerateTOML("issues.toml", "toml")
	if err != nil {
		log.Fatalln(err)
	}
```

On inspecting `toml/SA4017.toml`, we can see the following contents:

```toml
Code = "SA4017"
Text = "Sprint is a pure function but its return value is ignored"
ShortDescription = "Sprint is a pure function but its return value is ignored"
Description = "<h2>Sample</h2>\n"
```
