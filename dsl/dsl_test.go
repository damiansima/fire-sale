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
