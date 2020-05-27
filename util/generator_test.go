package util

import (
	"math"
	"strconv"
	"testing"
)

func TestRandInRange(t *testing.T) {
	type args struct {
		min int
		max int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"All positives", args{1, 10}, 0},
		{"Min is 0", args{0, 10}, 0},
		{"Max int", args{0, math.MaxInt32}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RandInRange(tt.args.min, tt.args.max)
			if got < tt.args.min || got >= tt.args.max {
				t.Errorf("RandInRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRandInList(t *testing.T) {
	type args struct {
		items []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Get Random item from list", args{[]string{"0", "1", "2", "3"}}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RandInList(tt.args.items)
			atoi, _ := strconv.Atoi(got)
			if got != tt.args.items[atoi] {
				t.Errorf("RandInList() = %v, want %v", got, got)
			}
		})
	}
}
