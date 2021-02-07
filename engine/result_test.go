package engine

import (
	"reflect"
	"testing"
	"time"
)

func Test_getOrCreateScenarioResult(t *testing.T) {
	scenarioResults := map[int]*ScenarioResult{}

	firstJob := Job{ScenarioId: 0, Name: "job-1", Timeout: time.Duration(1)}
	secondJob := Job{ScenarioId: 1, Name: "job-2", Timeout: time.Duration(2)}

	firstScenarioResult := ScenarioResult{Name: firstJob.Name, DefinedTimeout: firstJob.Timeout, RequestCount: 0}
	secondScenarioResult := ScenarioResult{Name: secondJob.Name, DefinedTimeout: secondJob.Timeout, RequestCount: 0}

	type args struct {
		result          Result
		scenarioResults map[int]*ScenarioResult
	}
	tests := []struct {
		name                string
		args                args
		scenarioResultCount int
		want                *ScenarioResult
	}{
		{"create first job", args{Result{job: firstJob}, scenarioResults}, 1, &firstScenarioResult},
		{"create second job", args{Result{job: secondJob}, scenarioResults}, 2, &secondScenarioResult},
		{"get first job", args{Result{job: firstJob}, scenarioResults}, 2, &firstScenarioResult},
		{"get second job", args{Result{job: secondJob}, scenarioResults}, 2, &secondScenarioResult},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getOrCreateScenarioResult(tt.args.result, tt.args.scenarioResults)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getOrCreateScenarioResult() = %v, want %v", got, tt.want)
			}
			currentScenarioResultCount := len(scenarioResults)
			if currentScenarioResultCount != tt.scenarioResultCount {
				t.Errorf("Scenario result count is wrong got = %d, want %v", currentScenarioResultCount, tt.scenarioResultCount)
			}
		})
	}
}

func Test_updateScenarioResult(t *testing.T) {
	job := Job{ScenarioId: 0, Name: "job-1", Timeout: time.Duration(1)}
	overallScenarioResult := ScenarioResult{}
	scenarioResult := ScenarioResult{Name: job.Name, DefinedTimeout: job.Timeout, RequestCount: 0}
	type args struct {
		result           Result
		actualServerTime time.Duration
		scenarioResult   *ScenarioResult
		overallResult    *ScenarioResult
	}
	tests := []struct {
		name string
		args args
	}{
		{"success result", args{result: Result{Status: 200, Timeout: false}, actualServerTime: time.Duration(1), scenarioResult: &scenarioResult, overallResult: &overallScenarioResult}},
		{"fail result", args{result: Result{Status: 400, Timeout: false}, actualServerTime: time.Duration(1), scenarioResult: &scenarioResult, overallResult: &overallScenarioResult}},
		{"fail result", args{result: Result{Status: 0, Timeout: true}, actualServerTime: time.Duration(1), scenarioResult: &scenarioResult, overallResult: &overallScenarioResult}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overallRFailCount := overallScenarioResult.FailCount
			overallRTimeoutCount := overallScenarioResult.TimeoutCount
			overallRSuccessCount := overallScenarioResult.SuccessCount

			scenarioRFailCount := scenarioResult.FailCount
			scenarioRTimeoutCount := scenarioResult.TimeoutCount
			scenarioRSuccessCount := scenarioResult.SuccessCount

			updateScenarioResult(tt.args.result, tt.args.actualServerTime, tt.args.scenarioResult, tt.args.overallResult)
			if tt.args.result.Status > 0 && tt.args.result.Status < 300 {
				if scenarioResult.SuccessCount != scenarioRSuccessCount+1 {
					t.Errorf("Scenario result count is wrong got = %d, want %v", scenarioResult.SuccessCount, scenarioRSuccessCount+1)
				}

				if overallScenarioResult.SuccessCount != overallRSuccessCount+1 {
					t.Errorf("Overall Scenario result count is wrong got = %d, want %v", overallScenarioResult.SuccessCount, overallRSuccessCount+1)
				}
			}

			if tt.args.result.Status >= 300 {
				if scenarioResult.FailCount != scenarioRFailCount+1 {
					t.Errorf("Scenario result fail count is wrong got = %d, want %v", scenarioResult.FailCount, scenarioRFailCount+1)
				}

				if overallScenarioResult.FailCount != overallRFailCount+1 {
					t.Errorf("Overall Scenario fail result count is wrong got = %d, want %v", overallScenarioResult.FailCount, overallRFailCount+1)
				}
			}

			if tt.args.result.Timeout {
				if scenarioResult.TimeoutCount != scenarioRTimeoutCount+1 {
					t.Errorf("Scenario result timeout count is wrong got = %d, want %v", scenarioResult.TimeoutCount, scenarioRTimeoutCount+1)
				}

				if overallScenarioResult.TimeoutCount != overallRTimeoutCount+1 {
					t.Errorf("Overall Scenario timeout result count is wrong got = %d, want %v", overallScenarioResult.TimeoutCount, overallRTimeoutCount+1)
				}
			}
		})
	}
}

func TestConsumeResults(t *testing.T) {
	tests := []struct {
		name    string
		results []Result
		report  Report
	}{
		{"consume one result success", []Result{Result{Status: 200, Timeout: false, job: Job{ScenarioId: 0}}}, Report{}},
		{"consume one result fail", []Result{Result{Status: 400, Timeout: false, job: Job{ScenarioId: 0}}}, Report{}},
		{"consume one result timeout", []Result{Result{Status: 0, Timeout: true, job: Job{ScenarioId: 0}}}, Report{}},
		{"consume one result success job warmup", []Result{Result{Status: 0, Timeout: true, job: Job{IsWarmup: true}}}, Report{}},
		{"consume two results 2 success", []Result{Result{Status: 200, Timeout: false, job: Job{ScenarioId: 0}}, {Status: 200, Timeout: false, job: Job{ScenarioId: 1}}}, Report{}},
		{"consume two results 1 success 1 failure", []Result{Result{Status: 200, Timeout: false, job: Job{ScenarioId: 0}}, {Status: 401, Timeout: false, job: Job{ScenarioId: 1}}}, Report{}},
		{"consume two results 1 success 1 timeout", []Result{Result{Status: 200, Timeout: false, job: Job{ScenarioId: 0}}, {Status: 0, Timeout: true, job: Job{ScenarioId: 1}}}, Report{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan bool)
			results := make(chan Result, 1000)

			var resultsCount int
			var successCount int64
			var failCount int64
			var timeoutCount int64
			for _, r := range tt.results {
				if r.job.IsWarmup {
					continue
				}
				resultsCount++
				if r.Timeout {
					timeoutCount++
				} else if 0 < r.Status && r.Status < 300 {
					successCount++
				} else {
					failCount++
				}

				results <- r
			}

			go ConsumeResults(results, done, &tt.report)

			close(results)
			<-done

			if len(tt.report.ScenarioResults) != resultsCount {
				t.Errorf("Report scenarion result count is wrong got = %d, want = %d", len(tt.report.ScenarioResults), resultsCount)
			}

			if int(tt.report.OverallResult.RequestCount) != resultsCount {
				t.Errorf("Overall request count is wrong got = %d, want = %d", tt.report.OverallResult.RequestCount, resultsCount)
			}

			if tt.report.OverallResult.SuccessCount != successCount {
				t.Errorf("Overall sucess count is wrong got = %d, want = %d", tt.report.OverallResult.SuccessCount, successCount)
			}

			if tt.report.OverallResult.FailCount != failCount {
				t.Errorf("Overall fail count is wrong got = %d, want = %d", tt.report.OverallResult.FailCount, failCount)
			}

			if tt.report.OverallResult.TimeoutCount != timeoutCount {
				t.Errorf("Overall timeout count is wrong got = %d, want = %d", tt.report.OverallResult.TimeoutCount, timeoutCount)
			}
		})
	}
}
