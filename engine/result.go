package engine

import (
	"github.com/influxdata/tdigest"
	log "github.com/sirupsen/logrus"
	"time"
)

type Result struct {
	// TODO start and end are part of the Trace really remove it
	Start   time.Time
	End     time.Time
	Trace   Trace
	Status  int
	Timeout bool
	job     Job
}

func (r *Result) isSuccessful() bool {
	return r.job.ValidateSuccess(r.Status)
}

func ConsumeResults(results chan Result, done chan bool, report *Report) {
	overallScenarioResult := ScenarioResult{}
	scenarioResults := make(map[int]*ScenarioResult)

	go func() {
		var last int64
		for _ = range time.Tick(10 * time.Second) {
			requestPerPeriod := overallScenarioResult.RequestCount - last
			log.Infof("Requesting: [%d] RPS | [%d] request every 10s ...", requestPerPeriod/10, requestPerPeriod)
			last = overallScenarioResult.RequestCount
		}
	}()

	//TODO we need to change this value and do memory profile
	td := tdigest.NewWithCompression(100000)
	var elapsedNetworkLast time.Duration

	// TODO allow for a channel to plot data points
	for result := range results {
		if result.job.IsWarmup {
			// TODO we may want to have data about this
			log.Debugf("Warmp result skiping")
			continue
		}
		overallScenarioResult.RequestCount++
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

		overallScenarioResult.DurationRequestSum += actualServerTime
		// TODO we should not account failed request but we should account timeout
		td.Add(actualServerTime.Seconds(), 1)

		log.Tracef("The job id [%d] lasted [%s||%s||%s] status [%d] - timeout [%t]", result.job.Id, elapsedOverall, elapsedRequest, actualServerTime, result.Status, result.Timeout)
		scenarioResult := getOrCreateScenarioResult(result, scenarioResults)
		updateScenarioResult(result, actualServerTime, scenarioResult, &overallScenarioResult)
	}
	overallScenarioResult.Td = *td

	report.OverallResult = overallScenarioResult
	report.ScenarioResults = scenarioResults

	done <- true
}

func getOrCreateScenarioResult(result Result, scenarioResults map[int]*ScenarioResult) *ScenarioResult {
	scenarioResult, ok := scenarioResults[result.job.ScenarioId]
	if !ok {
		scenarioResult = &ScenarioResult{
			Name:           result.job.Name,
			DefinedTimeout: result.job.Timeout,
			RequestCount:   0,
		}
	}
	scenarioResults[result.job.ScenarioId] = scenarioResult
	return scenarioResult
}

func updateScenarioResult(result Result, actualServerTime time.Duration, scenarioResult *ScenarioResult, overallResult *ScenarioResult) {
	scenarioResult.RequestCount++
	scenarioResult.DurationRequestSum += actualServerTime
	scenarioResult.Td.Add(actualServerTime.Seconds(), 1)

	if result.Timeout {
		overallResult.TimeoutCount++
		scenarioResult.TimeoutCount++
	} else if result.isSuccessful() {
		overallResult.SuccessCount++
		scenarioResult.SuccessCount++
	} else {
		overallResult.FailCount++
		scenarioResult.FailCount++
	}
}
