package engine

import (
	"encoding/json"
	"fmt"
	"github.com/influxdata/tdigest"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

type ScenarioResult struct {
	Name               string
	RequestCount       int64
	FailCount          int64
	SuccessCount       int64
	TimeoutCount       int64
	Td                 tdigest.TDigest
	DurationRequestSum time.Duration
}

func (sr *ScenarioResult) SuccessRate() float32 {
	return float32((sr.SuccessCount * 100) / sr.RequestCount)
}

func (sr *ScenarioResult) FailRate() float32 {
	return float32(((sr.TimeoutCount + sr.FailCount) * 100) / sr.RequestCount)
}

func (sr *ScenarioResult) TimoutRate() float32 {
	return float32(((sr.TimeoutCount) * 100) / sr.RequestCount)
}

// TODO average time should taken from configuration with/without latency
func (sr *ScenarioResult) RequestDurationAvg() time.Duration {
	return time.Duration(sr.DurationRequestSum.Nanoseconds() / sr.RequestCount)
}

type Report struct {
	OverallResult   ScenarioResult
	ScenarioResults map[int]ScenarioResult
}

// TODO this needs to be moved
func ConsumeResults(results chan Result, done chan bool, reportType, reportFilePath string) {
	overallResult := ScenarioResult{}
	scenarioResults := make(map[int]ScenarioResult)

	// TODO THIS SHOULD BE ACCUMULATED FOR REPORT PURPOSES
	var last int64
	go func() {
		for _ = range time.Tick(10 * time.Second) {
			requestPerPeriod := overallResult.RequestCount - last
			log.Infof("Request per 10 second [%d] | per 1 second [%d]...", requestPerPeriod, requestPerPeriod/10)
			last = overallResult.RequestCount
		}
	}()

	//TODO we need to change this value and do memory profile
	td := tdigest.NewWithCompression(100000)
	var elapsedNetworkLast time.Duration

	// TODO allow for a channel to plot data points
	for result := range results {
		overallResult.RequestCount++
		elapsedOverall := result.End.Sub(result.Start)
		elapsedNetwork := result.Trace.ConnectDoneTime.Sub(result.Trace.ConnectStartTime)
		elapsedRequest := result.Trace.GotFirstResponseByteTime.Sub(result.Trace.WroteRequestTime)
		//TODO WE ASSUME NETWORK AS LATENCY MAY BE KILL IT?
		if elapsedNetwork != 0 {
			elapsedNetworkLast = elapsedNetwork
			log.Tracef("change network time")
		}
		// TODO this measurement should be able to turn on and off
		actualServerTime := elapsedRequest - elapsedNetworkLast
		if actualServerTime < 0 {
			actualServerTime = -1 * actualServerTime
		}

		overallResult.DurationRequestSum += actualServerTime
		// TODO we should not account failed request but we should account timeout
		td.Add(actualServerTime.Seconds(), 1)

		log.Tracef("The job id [%d] lasted [%s||%s||%s] status [%d] - timeout [%t]", result.job.Id, elapsedOverall, elapsedRequest, actualServerTime, result.Status, result.Timeout)

		reportResult := buildScenarioResult(result, actualServerTime, scenarioResults, &overallResult)
		scenarioResults[result.job.ScenarioId] = reportResult
	}
	overallResult.Td = *td

	report := Report{OverallResult: overallResult, ScenarioResults: scenarioResults}
	printResults(report, reportType, reportFilePath)

	done <- true
}

func buildScenarioResult(result Result, actualServerTime time.Duration, scenarioResults map[int]ScenarioResult, overallResult *ScenarioResult) ScenarioResult {
	scenarioResult, ok := scenarioResults[result.job.ScenarioId]
	if !ok {
		scenarioResult = ScenarioResult{
			Name:         result.job.Name,
			RequestCount: 0,
		}
	}
	scenarioResult.RequestCount++
	scenarioResult.DurationRequestSum += actualServerTime
	scenarioResult.Td.Add(actualServerTime.Seconds(), 1)

	if result.Timeout {
		overallResult.TimeoutCount++
		scenarioResult.TimeoutCount++
	} else if result.Status > 0 && result.Status < 300 {
		overallResult.SuccessCount++
		scenarioResult.SuccessCount++
	} else {
		//TODO BUG: Fail percentage is not accurate
		overallResult.FailCount++
		scenarioResult.FailCount++
	}
	return scenarioResult
}

// TODO this needs to me moved to a report module
func printResults(report Report, reportType, reportFilePath string) {
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

func buildReportLines(report Report) []string {
	reportLines := make([]string, 0)

	reportLines = append(reportLines, fmt.Sprintf("********************************************************"))
	reportLines = append(reportLines, fmt.Sprintf("*                      Results                         *"))
	reportLines = append(reportLines, fmt.Sprintf("********************************************************"))
	reportLines = append(reportLines, fmt.Sprintf("========================================================"))
	reportLines = append(reportLines, fmt.Sprintf("=                     Scenarios                        ="))
	reportLines = append(reportLines, fmt.Sprintf("========================================================"))
	for key, scenarioResult := range report.ScenarioResults {
		reportLines = append(reportLines, fmt.Sprintf("Scenario - %s - ID: [%d]", scenarioResult.Name, key))
		reportLines = buildScenarioReportLines(reportLines, scenarioResult)
		reportLines = append(reportLines, fmt.Sprintf("--------------------------------------------------------"))
	}
	reportLines = append(reportLines, fmt.Sprintf("========================================================"))
	reportLines = append(reportLines, fmt.Sprintf("=                     Overall                          ="))
	reportLines = append(reportLines, fmt.Sprintf("========================================================"))
	reportLines = buildScenarioReportLines(reportLines, report.OverallResult)
	reportLines = append(reportLines, fmt.Sprintf("Timeout [%d] - Fail [%d] - Success [%d]  ", report.OverallResult.TimeoutCount, report.OverallResult.FailCount, report.OverallResult.SuccessCount))
	return reportLines
}

func buildScenarioReportLines(reportLines []string, scenarioResult ScenarioResult) []string {
	reportLines = append(reportLines, fmt.Sprintf("Success [%f%%] - Fail [%f%%]", scenarioResult.SuccessRate(), scenarioResult.FailRate()))
	reportLines = append(reportLines, fmt.Sprintf("Request average [%s] ", scenarioResult.RequestDurationAvg()))
	reportLines = append(reportLines, fmt.Sprintf("Request total [%d] ", scenarioResult.RequestCount))

	reportLines = append(reportLines, fmt.Sprintf("99th %fms", scenarioResult.Td.Quantile(0.99)/time.Millisecond.Seconds()))
	reportLines = append(reportLines, fmt.Sprintf("90th %fms", scenarioResult.Td.Quantile(0.9)/time.Millisecond.Seconds()))
	reportLines = append(reportLines, fmt.Sprintf("75th %fms", scenarioResult.Td.Quantile(0.75)/time.Millisecond.Seconds()))
	reportLines = append(reportLines, fmt.Sprintf("50th %fms", scenarioResult.Td.Quantile(0.5)/time.Millisecond.Seconds()))
	return reportLines
}
