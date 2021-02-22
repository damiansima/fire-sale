package engine

import (
	"github.com/influxdata/tdigest"
	"reflect"
	"testing"
	"time"
)

type fields struct {
	Name               string
	RequestCount       int64
	FailCount          int64
	SuccessCount       int64
	TimeoutCount       int64
	DefinedTimeout     time.Duration
	Td                 tdigest.TDigest
	DurationRequestSum time.Duration
}

func (f *fields) buildScenarioResult() ScenarioResult {
	return ScenarioResult{
		Name:               f.Name,
		RequestCount:       f.RequestCount,
		FailCount:          f.FailCount,
		SuccessCount:       f.SuccessCount,
		TimeoutCount:       f.TimeoutCount,
		DefinedTimeout:     f.DefinedTimeout,
		Td:                 f.Td,
		DurationRequestSum: f.DurationRequestSum,
	}
}

func TestScenarioResult_SuccessRate(t *testing.T) {

	tests := []struct {
		name   string
		fields fields
		want   float32
	}{
		{"10%", fields{RequestCount: 10, SuccessCount: 1}, float32(10)},
		{"50%", fields{RequestCount: 10, SuccessCount: 5}, float32(50)},
		{"100%", fields{RequestCount: 10, SuccessCount: 10}, float32(100)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := tt.fields.buildScenarioResult()
			if got := sr.SuccessRate(); got != tt.want {
				t.Errorf("SuccessRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScenarioResult_FailRate(t *testing.T) {
	tests := []struct {
		name   string
		fields fields
		want   float32
	}{
		{"10%", fields{RequestCount: 10, FailCount: 1}, float32(10)},
		{"50%", fields{RequestCount: 10, FailCount: 5}, float32(50)},
		{"100%", fields{RequestCount: 10, FailCount: 10}, float32(100)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := tt.fields.buildScenarioResult()
			if got := sr.FailRate(); got != tt.want {
				t.Errorf("FailRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScenarioResult_TimoutRate(t *testing.T) {

	tests := []struct {
		name   string
		fields fields
		want   float32
	}{
		{"10%", fields{RequestCount: 10, TimeoutCount: 1}, float32(10)},
		{"50%", fields{RequestCount: 10, TimeoutCount: 5}, float32(50)},
		{"100%", fields{RequestCount: 10, TimeoutCount: 10}, float32(100)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := tt.fields.buildScenarioResult()
			if got := sr.TimoutRate(); got != tt.want {
				t.Errorf("TimoutRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildReportLines(t *testing.T) {
	type args struct {
		report Report
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildReportLines(tt.args.report); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildReportLines() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildScenarioReportLines(t *testing.T) {
	type args struct {
		reportLines    []string
		scenarioResult *ScenarioResult
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildScenarioReportLines(tt.args.reportLines, tt.args.scenarioResult); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildScenarioReportLines() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_printReport(t *testing.T) {
	type args struct {
		report         Report
		reportType     string
		reportFilePath string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
