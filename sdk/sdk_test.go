package sdk

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/deepsourcelabs/deepsource-go/sdk/triggers"
	"github.com/deepsourcelabs/deepsource-go/sdk/types"
)

func TestStaticCheck(t *testing.T) {
	t.Run("verify staticcheck", func(t *testing.T) {
		a := CLIAnalyzer{
			Name:    "staticcheck",
			Command: "staticcheck",
			Args:    []string{"-f", "json", "./triggers/staticcheck/..."},
			ExportOpts: ExportOpts{
				Path: "triggers/staticcheck/issues.json",
				Type: "json",
			},
		}

		err := a.Run()
		if err != nil {
			t.Error(err)
		}

		// read the generated report
		reportContent, err := os.ReadFile("triggers/staticcheck/issues.json")
		if err != nil {
			t.Error(err)
		}

		var report types.AnalysisReport
		err = json.Unmarshal(reportContent, &report)
		if err != nil {
			t.Error(err)
		}

		// do a verification check for the generated report
		err = triggers.Verify(report, "triggers/staticcheck/staticcheck.go")
		if err != nil {
			t.Error(err)
		}

		// cleanup after test
		err = os.Remove("triggers/staticcheck/issues.json")
		if err != nil {
			t.Error(err)
		}
	})
}
