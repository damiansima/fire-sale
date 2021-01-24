package engine

import (
	"github.com/influxdata/tdigest"
	log "github.com/sirupsen/logrus"
	"time"
)

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

