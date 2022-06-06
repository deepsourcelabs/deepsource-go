package sdk

import (
	"bytes"
	"log"
	"os/exec"

	"github.com/deepsourcelabs/deepsource-go/sdk/types"
	"github.com/deepsourcelabs/deepsource-go/sdk/utils"
	"github.com/deepsourcelabs/deepsource-go/sdk/utils/processors"
)

// The main analyzer interface. Analyzers must implement Run and Processor.
type Analyzer interface {
	Run() error
	Processor(result interface{}) (types.AnalysisReport, error)
}

// CLIAnalyzer is used for creating an analyzer.
type CLIAnalyzer struct {
	Name       string
	Command    string
	Args       []string
	ExportOpts ExportOpts
}

type ExportOpts struct {
	Path string
	Type string
}

// Run executes the analyzer and streams the output to the processor.
func (a *CLIAnalyzer) Run() error {
	cmd := exec.Command(a.Command, a.Args...)

	// store the process's standard output in a buffer
	var out bytes.Buffer
	cmd.Stdout = &out

	// TODO: handle exit status 1
	_ = cmd.Run()

	// fetch report from processor
	report, err := a.Processor(out.String())
	if err != nil {
		return err
	}

	// save report to file
	err = utils.SaveReport(report, a.ExportOpts.Path, a.ExportOpts.Type)
	if err != nil {
		return err
	}

	return nil
}

// Processor takes the analyzer output and generates a report.
func (a *CLIAnalyzer) Processor(result interface{}) (types.AnalysisReport, error) {
	var report types.AnalysisReport
	var err error

	// use custom processors for each major linter/analyzer
	switch a.Name {
	case "staticcheck":
		report, err = processors.StaticCheck(result)
	default:
		// if a match is not found, the user needs to implement a processor
		log.Printf("custom processor needs to be implemented for %s.\n", a.Name)
	}

	return report, err
}

// GenerateTOML helps in generating TOML files for each issue from a JSON file.
func (a *CLIAnalyzer) GenerateTOML(filename string, rootDir string) error {
	// fetch parsed issues
	issues, err := utils.ParseIssues(filename)
	if err != nil {
		return err
	}

	// generate TOML files
	err = utils.BuildTOML(issues, rootDir)
	if err != nil {
		return err
	}

	return nil
}
