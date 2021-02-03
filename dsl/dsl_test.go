package dsl

import (
	"github.com/damiansima/fire-sale/engine"
	"reflect"
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	type args struct {
		duration string
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{"Parse empty", args{""}, time.Duration(0) * time.Minute},
		{"Parse 0 without unit", args{"0"}, time.Duration(0) * time.Minute},
		{"Parse 0 s", args{"0ms"}, time.Duration(0) * time.Minute},
		{"Parse 0 s", args{"0s"}, time.Duration(0) * time.Minute},
		{"Parse 0 s", args{"0m"}, time.Duration(0) * time.Minute},
		{"Parse without unit", args{"3"}, time.Duration(3) * time.Minute},
		{"Parse ns", args{"300000000000ns"}, time.Duration(5) * time.Minute},
		{"Parse ms", args{"300000ms"}, time.Duration(5) * time.Minute},
		{"Parse m", args{"3m"}, time.Duration(3) * time.Minute},
		{"Parse h", args{"1h"}, time.Duration(60) * time.Minute},
		{"Parse h and m", args{"1h30m"}, time.Duration(90) * time.Minute},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseDuration(tt.args.duration); got != tt.want {
				t.Errorf("ParseDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapRampUp(t *testing.T) {
	type args struct {
		configuration Configuration
	}
	tests := []struct {
		name string
		args args
		want engine.RampUp
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapRampUp(tt.args.configuration); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapRampUp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapCertificates(t *testing.T) {
	type args struct {
		configuration Configuration
	}
	tests := []struct {
		name string
		args args
		want engine.Certificates
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapCertificates(tt.args.configuration); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapCertificates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapScenarios(t *testing.T) {
	type args struct {
		configuration Configuration
	}
	tests := []struct {
		name string
		args args
		want []engine.Scenario
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapScenarios(tt.args.configuration); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapScenarios() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseConfiguration(t *testing.T) {
	type args struct {
		configPath string
	}
	tests := []struct {
		name string
		args args
		want Configuration
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseConfiguration(tt.args.configPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseConfiguration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mapScenario(t *testing.T) {
	type args struct {
		scId  int
		dslSc Scenario
		host  string
	}
	tests := []struct {
		name string
		args args
		want engine.Scenario
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapScenario(tt.args.scId, tt.args.dslSc, tt.args.host); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapScenario() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildJobSuccessValidator(t *testing.T) {
	type args struct {
		status []string
	}

	tests := []struct {
		name                    string
		args                    args
		want                    func(int) bool
		testStatus              int
		functionShouldBeSuccess bool
	}{
		{"nil array should return nil", args{nil}, nil, 0, false},
		{"empty array should return nil", args{[]string{}}, nil, 0, false},
		{"{200a} array should return nil", args{[]string{"200a"}}, nil, 0, false},
		{"{0a-300} array should return nil", args{[]string{"0a-300"}}, nil, 0, false},
		{"{0-300a} array should return nil", args{[]string{"0-300a"}}, nil, 0, false},
		{"{1-0-300-} array should return nil", args{[]string{"0-300-1"}}, nil, 0, false},
		{"{-0-300-} array should return nil", args{[]string{"0-300-1"}}, nil, 0, false},
		{"{0-300} array should return nil", args{[]string{"0-300-1"}}, nil, 0, false},
		{"{0-300-1} array should return nil", args{[]string{"0-300-1"}}, nil, 0, false},
		{"{200} array should return and 200 evaluate to true", args{[]string{"200"}}, func(int) bool { return true }, 200, true},
		{"{200} array should return and 100 evaluate to false", args{[]string{"200"}}, func(int) bool { return true }, 100, false},
		{"{0-300} array should return and 0 evaluate to false", args{[]string{"0-300"}}, func(int) bool { return true }, 0, false},
		{"{0-300} array should return and 200 evaluate to true", args{[]string{"0-300"}}, func(int) bool { return true }, 200, true},
		{"{0-300} array should return and 300 evaluate to false", args{[]string{"0-300"}}, func(int) bool { return true }, 300, false},
		{"{0-300,400} array should return and 200 evaluate to true", args{[]string{"0-300", "400"}}, func(int) bool { return true }, 200, true},
		{"{0-300,400} array should return and 400 evaluate to true", args{[]string{"0-300", "400"}}, func(int) bool { return true }, 400, true},
		{"{0-300,400} array should return and 401 evaluate to false", args{[]string{"0-300", "400"}}, func(int) bool { return true }, 401, false},
		{"{0-300,400} array should return and 0 evaluate to false", args{[]string{"0-300", "400"}}, func(int) bool { return true }, 0, false},
		{"{0-300,400} array should return and 300 evaluate to false", args{[]string{"0-300", "400"}}, func(int) bool { return true }, 300, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildJobSuccessValidator(tt.args.status)
			if tt.want == nil && got != nil {
				t.Errorf("buildJobSuccessValidator() should have return a nil function")
			} else if tt.want != nil && got == nil {
				t.Errorf("buildJobSuccessValidator() should have return a function")
			} else if tt.want != nil && got != nil {
				gotIsSuccess := got(tt.testStatus)
				if gotIsSuccess != tt.functionShouldBeSuccess {
					t.Errorf("builtFunction(%d) should have return %v instead of %v", tt.testStatus, tt.functionShouldBeSuccess, gotIsSuccess)
				}
			}
		})
	}
}
