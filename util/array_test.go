package util_test

import (
	"github.com/damiansima/fire-sale/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSum(t *testing.T) {
	scenarios := []struct {
		name     string
		array    []float32
		expected float32
	}{
		{"Sum Values", []float32{0.25, -0.25, 0.25, -0.25}, 0.00},
		{"Sum Positive Values", []float32{0.25, 0.25, 0.25, 0.25}, 1.00},
		{"Sum Negative Values", []float32{-0.25, -0.25, -0.25, -0.25}, -1.00},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			assert.Equal(t, scenario.expected, util.Sum(scenario.array))
		})
	}
}

func TestEquals(t *testing.T) {
	scenarios := []struct {
		name     string
		a        []float32
		b        []float32
		expected bool
	}{
		{"Empty Arrays", []float32{}, []float32{}, true},
		{"Different Len", []float32{}, []float32{0.1}, false},
		{"Equal Arrays", []float32{0.1, 0.1}, []float32{0.1, 0.1}, true},
		{"Different Arrays", []float32{0.1, 0.1}, []float32{0.1, 0.2}, false},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			assert.Equal(t, scenario.expected, util.Equals(scenario.a, scenario.b))
		})
	}
}
