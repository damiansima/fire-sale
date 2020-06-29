package engine

import (
	"github.com/damiansima/fire-sale/util"
	"testing"
)

func TestGetDistribution(t *testing.T) {
	scenarios := []struct {
		name         string
		distribution []float32
		expected     []float32
	}{
		{"100%", []float32{0.1}, []float32{100}},
		{"50%", []float32{0.5, 0.5}, []float32{50, 100}},
		{"25%", []float32{0.25, 0.25, 0.25, 0.25}, []float32{25, 50, 75, 100}},
		{"20%", []float32{0.2, 0.2, 0.2, 0.2, 0.2}, []float32{20, 40, 60, 80, 100}},
	}
	// TODO add scenarios were we return errors

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			if actual, err := BuildBuckets(scenario.distribution); err != nil && util.Equals(actual, scenario.expected) {
				t.Errorf("SelectBucket() = %v, expected %v", actual, scenario.expected)
			}
		})
	}
}

func TestSelectBucket(t *testing.T) {
	scenarios := []struct {
		name     string
		buckets  []float32
		expected int
	}{
		{"One bucket lower edge", []float32{100}, 0},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			if actual := SelectBucket(scenario.buckets); actual != scenario.expected {
				t.Errorf("SelectBucket() = %v, expected %v", actual, scenario.expected)
			}
		})
	}
}

func TestSelectDeterministicBucket(t *testing.T) {
	scenarios := []struct {
		name         string
		value        int
		distribution []float32
		expected     int
	}{
		{"One bucket lower edge", 0, []float32{100}, 0},
		{"One bucket middle value", 35, []float32{100}, 0},
		{"One bucket max edge", 100, []float32{100}, 0},
		{"One bucket value not in bucket", 101, []float32{100}, -1},
		{"Several buckets - lower edge", 0, []float32{25, 50, 75, 100}, 0},
		{"Several buckets - middle value", 35, []float32{25, 50, 75, 100}, 1},
		{"Several buckets - max edge", 100, []float32{25, 50, 75, 100}, 3},
		{"Several buckets - value not in bucket", 101, []float32{25, 50, 75, 100}, -1},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			if actual := SelectDeterministicBucket(scenario.value, scenario.distribution); actual != scenario.expected {
				t.Errorf("SelectDeterministicBucket() = %v, expected %v", actual, scenario.expected)
			}
		})
	}
}
