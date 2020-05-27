package engine

import (
	"github.com/storozhukBM/verifier"
	"strconv"
)

type ValueProvider interface {
	Provide() string
}

type RandomNumberProvider struct {
	Min int
	Max int
}

func NewRandomNumberProvider(min, max int) (*RandomNumberProvider, error) {
	verify := verifier.New()
	verify.That(min < max, "min should be < than max")

	if verify.GetError() != nil {
		return nil, verify.GetError()
	}

	return &RandomNumberProvider{Min: min, Max: max}, nil
}

func (p *RandomNumberProvider) Provide() string {
	return strconv.Itoa(GenerateRandomValue(p.Min, p.Max))
}

type ItemProvider struct {
	items      []string
	currentIdx int
}

func NewItemProvider(items []string) *ItemProvider {
	return &ItemProvider{items: items}
}
func (p *ItemProvider) Provide() string {
	item := p.items[p.currentIdx]
	p.currentIdx++
	if p.currentIdx == len(p.items) {
		p.currentIdx = 0
	}
	return item
}
