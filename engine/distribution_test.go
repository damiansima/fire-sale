package engine_test

import (
	"github.com/damiansima/fire-sale/engine"
	"github.com/damiansima/fire-sale/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetDistribution(t *testing.T) {
	scenarios := []struct {
		name         string
		distribution []float32
		expected     []float32
		shouldFail   bool
	}{
		{"100%", []float32{1}, []float32{100}, false},
		{"50%", []float32{0.5, 0.5}, []float32{50, 100}, false},
		{"25%", []float32{0.25, 0.25, 0.25, 0.25}, []float32{25, 50, 75, 100}, false},
		{"20%", []float32{0.2, 0.2, 0.2, 0.2, 0.2}, []float32{20, 40, 60, 80, 100}, false},
		{"Nil distribution should fail", nil, []float32{}, true},
		{"Empty distribution should fail", []float32{}, []float32{}, true},
		{"Over 1 distribution should fail", []float32{0.5, 0.5, 0.5}, []float32{}, true},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			actual, err := engine.BuildBuckets(scenario.distribution)
			if scenario.shouldFail {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.True(t, util.Equals(scenario.expected, actual))

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
			assert.Equal(t, scenario.expected, engine.SelectBucket(scenario.buckets))
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
			assert.Equal(t, scenario.expected, engine.SelectDeterministicBucket(scenario.value, scenario.distribution))
		})
	}
}
