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

func (r *Result) getTotalDuration() time.Duration {
	//elapsedOverall := result.End.Sub(result.Start)
	return r.Trace.PutIdleConnTime.Sub(r.Trace.GetConnTime)
}
func (r *Result) getConnectionDuration() time.Duration {
	// elapsedNetwork := result.Trace.ConnectDoneTime.Sub(result.Trace.ConnectStartTime)
	return r.Trace.GetConnTime.Sub(r.Trace.GetConnTime)
}
func (r *Result) getRequestDuration() time.Duration {
	// elapsedRequest := result.Trace.GotFirstResponseByteTime.Sub(result.Trace.WroteRequestTime)
	requestDuration := r.Trace.GotFirstResponseByteTime.Sub(r.Trace.WroteRequestTime)
	// TODO When did we need to do this right?
	if requestDuration < 0 {
		requestDuration = -1 * requestDuration
	}
	return requestDuration
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

	// TODO we should track times of success|errors|timeout separately
	// TODO we need to change this value and do memory profile
	td := tdigest.NewWithCompression(100000)

	// TODO allow for a channel to plot data points
	for result := range results {
		if result.job.IsWarmup {
			log.Debugf("Warmp result skiping")
			continue
		}
		overallScenarioResult.RequestCount++
		requestDuration := result.getRequestDuration()

		td.Add(requestDuration.Seconds(), 1)
		overallScenarioResult.DurationRequestSum += requestDuration

		log.Debugf("The job id [%d] lasted -  status [%d] - timeout [%t] - Durations: [Connection: %s||Request: %s||Total:%s]", result.job.Id, result.Status, result.Timeout, result.getConnectionDuration(), result.getRequestDuration(), result.getTotalDuration())
		scenarioResult := getOrCreateScenarioResult(result, scenarioResults)
		updateScenarioResult(result, requestDuration, scenarioResult, &overallScenarioResult)
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

func updateScenarioResult(result Result, requestDuration time.Duration, scenarioResult *ScenarioResult, overallResult *ScenarioResult) {
	scenarioResult.RequestCount++
	scenarioResult.DurationRequestSum += requestDuration
	scenarioResult.Td.Add(requestDuration.Seconds(), 1)

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
