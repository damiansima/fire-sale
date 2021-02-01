package engine

import (
	"testing"
)

func Test_balanceScenarioDistribution(t *testing.T) {
	type args struct {
		scenarios []Scenario
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"1 Scenario 0.00 distribution", args{[]Scenario{{Distribution: 0}}}, false},
		{"1 Scenario 1.00 distribution", args{[]Scenario{{Distribution: 1}}}, false},
		{"2 Scenarios 0.0 & 0.00 distribution", args{[]Scenario{{Distribution: 0}, {Distribution: 0}}}, false},
		{"2 Scenarios 0.5 & 0.00 distribution", args{[]Scenario{{Distribution: 0.5}, {Distribution: 0}}}, false},
		{"2 Scenarios 0.7 & 0.00 distribution", args{[]Scenario{{Distribution: 0.7}, {Distribution: 0}}}, false},
		{"3 Scenarios 0.0 & 0.00 & 0.00 distribution", args{[]Scenario{{Distribution: 0}, {Distribution: 0}, {Distribution: 0}}}, false},
		{"3 Scenarios 0.0 & 0.5 & 0.00 distribution", args{[]Scenario{{Distribution: 0}, {Distribution: 0.5}, {Distribution: 0}}}, false},
		{"Empty Scenarios", args{[]Scenario{}}, true},
		{"2 Scenarios 1 & 0.00 distribution", args{[]Scenario{{Distribution: 1.0}, {Distribution: 0}}}, true},
		{"2 Scenarios 0.7 & 0.7 distribution", args{[]Scenario{{Distribution: 0.7}, {Distribution: 0.7}}}, true},
		{"3 Scenarios 0.0 & 0.7 & 0.7 distribution", args{[]Scenario{{Distribution: 0.0}, {Distribution: 0.7}, {Distribution: 0.7}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balancedScenarios, err := balanceScenarioDistribution(tt.args.scenarios)
			if tt.wantErr && err == nil {
				t.Errorf("balanceScenarioDistribution() should have failderror = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				distributionSum := float32(0)
				for _, s := range balancedScenarios {
					distributionSum += s.Distribution
				}
				if distributionSum != 1 {
					t.Errorf("Scenarios distribution must add up to 1 and it adds up to %0.2f", distributionSum)
				}
			}
		})
	}
}
