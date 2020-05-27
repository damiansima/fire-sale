package engine

import (
	"github.com/influxdata/tdigest"
	log "github.com/sirupsen/logrus"
	"time"
)

type ReportResult struct {
	Name               string
	RequestCount       int64
	FailCount          int64
	SuccessCount       int64
	TimeoutCount       int64
	Td                 tdigest.TDigest
	DurationSum        time.Duration
	DurationRequestSum time.Duration
}

// TODO this needs to be moved
func ConsumeResults(results chan Result, done chan bool) {
	var durationSum time.Duration
	var durationRequestSum time.Duration
	var count int64

	var failCount int64
	var successCount int64
	var timeoutCount int64

	scenarioResults := make(map[int]ReportResult)

	// TODO THIS SHOULD BE ACCUMULATED FOR REPORT PURPOSES
	var last int64
	go func() {
		for _ = range time.Tick(10 * time.Second) {
			requestPerPeriod := count - last
			log.Infof("Request per 10 second [%d] | per 1 second [%d]...", requestPerPeriod, requestPerPeriod/10)
			last = count
		}
	}()

	//TODO we need to change this value and do memory profile
	td := tdigest.NewWithCompression(100000)
	var elapsedNetworkLast time.Duration

	// TODO allow for a channel to plot data points
	for result := range results {
		count++
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
		// TODO we should not account failed request but we should account timeout
		durationSum += elapsedOverall
		durationRequestSum += actualServerTime
		td.Add(actualServerTime.Seconds(), 1)

		log.Tracef("The job id [%d] lasted [%s||%s||%s] status [%d] - timeout [%t]", result.job.Id, elapsedOverall, elapsedRequest, actualServerTime, result.Status, result.Timeout)

		reportResult, ok := scenarioResults[result.job.ScenarioId]
		if !ok {
			reportResult = ReportResult{
				Name:         result.job.Name,
				RequestCount: 0,
			}
		}

		reportResult.RequestCount++
		reportResult.DurationSum += elapsedOverall
		reportResult.DurationRequestSum += actualServerTime
		reportResult.Td.Add(actualServerTime.Seconds(), 1)

		if result.Timeout {
			timeoutCount++
			reportResult.TimeoutCount++
		} else if result.Status > 0 && result.Status < 300 {
			successCount++
			reportResult.SuccessCount++
		} else {
			failCount++
			reportResult.FailCount++
		}
		scenarioResults[result.job.ScenarioId] = reportResult
	}

	overallResult := ReportResult{
		Name:               "Overal",
		RequestCount:       count,
		FailCount:          failCount, // TODO BUG: Fail percentage is not accurate
		SuccessCount:       successCount,
		TimeoutCount:       timeoutCount,
		Td:                 *td,
		DurationSum:        durationSum,
		DurationRequestSum: durationRequestSum,
	}

	printResults(overallResult, scenarioResults)

	done <- true
}

// TODO this needs to me moved to a report module
func printResults(overallResult ReportResult, scenarioResults map[int]ReportResult) {
	log.Infof("********************************************************")
	log.Infof("*                      Results                         *")
	log.Infof("********************************************************")

	log.Infof("========================================================")
	log.Infof("=                     Scenarios                        =")
	log.Infof("========================================================")
	for key, reportResult := range scenarioResults {
		log.Infof("Scenario - %s - ID: [%d]", reportResult.Name, key)
		log.Infof("Success [%f%%] - Fail [%f%%]", float32((reportResult.SuccessCount*100)/reportResult.RequestCount), float32(((reportResult.TimeoutCount+reportResult.FailCount)*100)/reportResult.RequestCount))
		log.Infof("Request average [%s] ", time.Duration(reportResult.DurationRequestSum.Nanoseconds()/reportResult.RequestCount))
		log.Infof("Request total [%d] average [%s] ", reportResult.RequestCount, time.Duration(reportResult.DurationSum.Nanoseconds()/reportResult.RequestCount))
		log.Infof("99th %fms", reportResult.Td.Quantile(0.99)/time.Millisecond.Seconds())
		log.Infof("90th %fms", reportResult.Td.Quantile(0.9)/time.Millisecond.Seconds())
		log.Infof("75th %fms", reportResult.Td.Quantile(0.75)/time.Millisecond.Seconds())
		log.Infof("50th %fms", reportResult.Td.Quantile(0.5)/time.Millisecond.Seconds())
		log.Infof("--------------------------------------------------------")
	}

	log.Infof("========================================================")
	log.Infof("=                     Overall                          =")
	log.Infof("========================================================")

	log.Infof("Success [%f%%] - Fail [%f%%]", float32((overallResult.SuccessCount*100)/overallResult.RequestCount), float32(((overallResult.TimeoutCount+overallResult.FailCount)*100)/overallResult.RequestCount))
	// TODO average time should taken from configuration with/without latency
	log.Infof("Request average [%s] ", time.Duration(overallResult.DurationRequestSum.Nanoseconds()/overallResult.RequestCount))
	log.Infof("Request total [%d] average [%s] ", overallResult.RequestCount, time.Duration(overallResult.DurationSum.Nanoseconds()/overallResult.RequestCount))
	// TODO we may need to change this lib at least inject it
	log.Infof("99th %fms", overallResult.Td.Quantile(0.99)/time.Millisecond.Seconds())
	log.Infof("90th %fms", overallResult.Td.Quantile(0.9)/time.Millisecond.Seconds())
	log.Infof("75th %fms", overallResult.Td.Quantile(0.75)/time.Millisecond.Seconds())
	log.Infof("50th %fms", overallResult.Td.Quantile(0.5)/time.Millisecond.Seconds())
	log.Infof("Timeout [%d] - Fail [%d] - Success [%d]  ", overallResult.TimeoutCount, overallResult.FailCount, overallResult.SuccessCount)
}
