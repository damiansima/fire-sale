package engine

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestNewRandomNumberProvider(t *testing.T) {
	tests := []struct {
		name    string
		min     int
		max     int
		wantErr bool
	}{
		{"Different min max", 9, 9, true},
		{"Negative max", 0, -9, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewRandomNumberProvider(test.min, test.max)
			if (err != nil) != test.wantErr {
				t.Errorf("NewRandomNumberProvider() error = %v, wantErr %v", err, test.wantErr)
				return
			}
		})
	}
}

func TestRandomNumberProvider_ProvideStr(t *testing.T) {
	tests := []struct {
		name    string
		min     int
		max     int
		wantErr bool
	}{
		{"Same min max", 0, 9, false},
		{"Negative min", -10, 9, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p, _ := NewRandomNumberProvider(test.min, test.max)
			got := p.Provide()
			gotInt, _ := strconv.Atoi(got)
			assert.True(t, gotInt >= test.min && gotInt <= test.max)
		})
	}
}

func TestItemProvider_Provide(t *testing.T) {
	tests := []struct {
		name       string
		items      []string
		currentIdx int
		nextIdx    int
	}{
		{"Should return a", []string{"a", "b", "c"}, 0, 1},
		{"Should return c and refresh", []string{"a", "b", "c"}, 2, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := &ItemProvider{items: test.items, currentIdx: test.currentIdx}
			assert.Equal(t, p.Provide(), p.items[test.currentIdx])
			assert.Equal(t, p.Provide(), p.items[test.nextIdx])
		})
	}
}
