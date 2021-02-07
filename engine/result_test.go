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
