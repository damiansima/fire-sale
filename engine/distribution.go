package engine

import (
	log "github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

// Generate an slices where each item is the value bellow which is the Distribution of numbers from 0 to 100
// TODO change name, change name to the variable Distribution too it is a probability Distribution
func GetDistribution(distribution []float32) []float32 {
	dist := make([]float32, len(distribution))
	dist[0] = distribution[0] * 100
	for i := 1; i < len(distribution); i++ {
		dist[i] = dist[i-1] + distribution[i]*100
	}
	log.Debugf("Building Distributions: Distribution %v -- Buckets %v", distribution, dist)
	return dist
}

// TODO change name, change name to the variable Distribution too
func SelectBucket(distribution []float32) int {
	random := random(0, 100)

	bucket := 0
	for float32(random) > distribution[bucket] {
		bucket++
	}
	log.Debugf("Selecting bucket %d", bucket)
	return bucket
}

func random(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}
