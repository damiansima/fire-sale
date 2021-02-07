package engine

import (
	"encoding/json"
	"fmt"
	"github.com/influxdata/tdigest"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

type Report struct {
	OverallResult   ScenarioResult
	ScenarioResults map[int]*ScenarioResult
}

type ScenarioResult struct {
	Name               string
	RequestCount       int64
	FailCount          int64
	SuccessCount       int64
	TimeoutCount       int64
	DefinedTimeout     time.Duration
	Td                 tdigest.TDigest
	DurationRequestSum time.Duration
}

func (sr *ScenarioResult) SuccessRate() float32 {
	return float32((sr.SuccessCount * 100) / sr.RequestCount)
}

func (sr *ScenarioResult) FailRate() float32 {
	return float32(((sr.FailCount) * 100) / sr.RequestCount)
}

func (sr *ScenarioResult) TimoutRate() float32 {
	return float32(((sr.TimeoutCount) * 100) / sr.RequestCount)
}

// TODO average time should taken from configuration with/without latency
func (sr *ScenarioResult) RequestDurationAvg() time.Duration {
	return time.Duration(sr.DurationRequestSum.Nanoseconds() / sr.RequestCount)
}

func buildReportLines(report Report) []string {
	reportLines := make([]string, 0)

	reportLines = append(reportLines, fmt.Sprintf("********************************************************"))
	reportLines = append(reportLines, fmt.Sprintf("*                      Results                         *"))
	reportLines = append(reportLines, fmt.Sprintf("********************************************************"))
	reportLines = append(reportLines, fmt.Sprintf("========================================================"))
	reportLines = append(reportLines, fmt.Sprintf("=                     Scenarios                        ="))
	reportLines = append(reportLines, fmt.Sprintf("========================================================"))
	for key, scenarioResult := range report.ScenarioResults {
		reportLines = append(reportLines, fmt.Sprintf("Scenario - %s - ID: [%d] - Timout: [%s]", scenarioResult.Name, key, scenarioResult.DefinedTimeout))
		reportLines = buildScenarioReportLines(reportLines, scenarioResult)
		reportLines = append(reportLines, fmt.Sprintf("--------------------------------------------------------"))
	}
	reportLines = append(reportLines, fmt.Sprintf("========================================================"))
	reportLines = append(reportLines, fmt.Sprintf("=                     Overall                          ="))
	reportLines = append(reportLines, fmt.Sprintf("========================================================"))
	reportLines = buildScenarioReportLines(reportLines, &report.OverallResult)
	return reportLines
}

func buildScenarioReportLines(reportLines []string, scenarioResult *ScenarioResult) []string {
	reportLines = append(reportLines, fmt.Sprintf("99th %.2fms", scenarioResult.Td.Quantile(0.99)/time.Millisecond.Seconds()))
	reportLines = append(reportLines, fmt.Sprintf("95th %.2fms", scenarioResult.Td.Quantile(0.95)/time.Millisecond.Seconds()))
	reportLines = append(reportLines, fmt.Sprintf("90th %.2fms", scenarioResult.Td.Quantile(0.9)/time.Millisecond.Seconds()))
	reportLines = append(reportLines, fmt.Sprintf("75th %.2fms", scenarioResult.Td.Quantile(0.75)/time.Millisecond.Seconds()))
	reportLines = append(reportLines, fmt.Sprintf("50th %.2fms", scenarioResult.Td.Quantile(0.5)/time.Millisecond.Seconds()))

	reportLines = append(reportLines, fmt.Sprintf("Request total [%d] ", scenarioResult.RequestCount))
	reportLines = append(reportLines, fmt.Sprintf("Request Counts: Fail [%d] - Timeout [%d] - Success [%d]  ", scenarioResult.FailCount, scenarioResult.TimeoutCount, scenarioResult.SuccessCount))
	reportLines = append(reportLines, fmt.Sprintf("Request Rates: Fail [%.2f%%] - Timeout [%.2f%%] - Success [%.2f%%]", scenarioResult.FailRate(), scenarioResult.TimoutRate(), scenarioResult.SuccessRate()))
	return reportLines
}

func printReport(report Report, reportType, reportFilePath string) {
	var jsonReport []byte
	var reportLines []string

	if reportType == "json" {
		jsonReport, _ = json.Marshal(report)
	} else {
		reportLines = buildReportLines(report)
	}

	if reportFilePath != "" {
		reportFile, _ := os.Create(reportFilePath)
		defer reportFile.Close()
		if reportType == "json" {
			reportFile.Write(jsonReport)
		} else {
			for _, line := range reportLines {
				fmt.Fprintln(reportFile, line)
			}
		}
	} else {
		if reportType == "json" {
			log.Infof(string(jsonReport))
		} else {
			for _, line := range reportLines {
				log.Info(line)
			}
		}
	}
}
