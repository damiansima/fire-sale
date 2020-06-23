package engine

import (
	"fmt"
	"math/rand"
	"time"
)

func GetDistribution(distribution []float32) []float32 {
	dist := make([]float32, len(distribution))
	dist[0] = distribution[0] * 100
	for i := 1; i < len(distribution); i++ {
		dist[i] = dist[i-1] + distribution[i]*100
	}
	return dist
}

func SelectBucket(distribution []float32) int {
	random := random(0, 100)
	fmt.Printf("Random %d", random)

	bucket := 0
	for float32(random) > distribution[bucket] {
		bucket++
	}
	return bucket
}

func random(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}
