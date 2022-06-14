package analyzers

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path"

	"github.com/deepsourcelabs/deepsource-go/analyzers/types"
)

type Processor interface {
	Process(bytes.Buffer) (types.AnalysisReport, error)
}

// CLIRunner is used for creating an analyzer.
type CLIRunner struct {
	Name             string
	Command          string
	Args             []string
	AllowedExitCodes []int
	Processor        Processor
	stdout           *bytes.Buffer
	stderr           *bytes.Buffer
	exitCode         int
}

// Stdout returns the stdout buffer.
func (a *CLIRunner) Stdout() bytes.Buffer {
	return *a.stdout
}

// Stderr returns the stderr buffer.
func (a *CLIRunner) Stderr() bytes.Buffer {
	return *a.stderr
}

// Run executes the analyzer and streams the output to the processor.
func (a *CLIRunner) Run() error {
	outBuf, errBuf, exitCode, err := runCmd(a.Command, a.Args, a.AllowedExitCodes)
	if err != nil {
		return err
	}

	a.stdout = &outBuf
	a.stderr = &errBuf
	a.exitCode = exitCode

	return nil
}

// runCmd returns the stdout and stderr streams, along with an exit code and error after running the command.
func runCmd(command string, args []string, allowedExitCodes []int) (bytes.Buffer, bytes.Buffer, int, error) {
	cmd := exec.Command(command, args...)

	// store stdout and stderr in buffers
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Start()
	if err != nil {
		return bytes.Buffer{}, bytes.Buffer{}, -1, err
	}

	// wait for the command to exit
	err = cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			// if exit code is allowed, return the buffers with no error
			for _, v := range allowedExitCodes {
				if v == exitCode {
					return outBuf, errBuf, exitCode, nil
				}
			}
		} else {
			// in case of errors, exit code is -1
			return outBuf, errBuf, -1, err
		}
	}

	// default exit code is 0
	return outBuf, errBuf, 0, nil
}

// SaveReport saves the analysis report to the local filesystem.
func (*CLIRunner) SaveReport(report types.AnalysisReport) error {
	// get toolbox path from environment variable
	toolboxPath := os.Getenv("TOOLBOX_PATH")

	// set report location
	filename := path.Join(toolboxPath, "analysis_report.json")

	data, err := json.MarshalIndent(report, "", "	")
	if err != nil {
		return err
	}

	if err = os.WriteFile(filename, data, 0644); err != nil {
		return err
	}

	return nil
}
