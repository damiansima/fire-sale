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
