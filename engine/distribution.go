package engine

import (
	"github.com/damiansima/fire-sale/util"
	"github.com/storozhukBM/verifier"
	"math/rand"
	"time"
)

// Provided a slice containing values between 0.0:1.0 it generates buckets containing values between 0:100
func BuildBuckets(distribution []float32) ([]float32, error) {
	verify := verifier.New()
	verify.That(distribution != nil, "Scenario distribution can't be nil")
	verify.That(len(distribution) > 0, "Scenario distribution must not be empty")
	verify.That(util.Sum(distribution) == 1, "Scenario distribution must add up 1. Current [%g]", util.Sum(distribution))
	if verify.GetError() != nil {
		return nil, verify.GetError()
	}

	buckets := make([]float32, len(distribution))
	buckets[0] = distribution[0] * 100
	for i := 1; i < len(distribution); i++ {
		buckets[i] = buckets[i-1] + distribution[i]*100
	}
	return buckets, nil
}

// It generates a value between 0:100 and it looks for the bucket containing the value generated
// It returns the idx representing a bucket in of the bucket containing the value, -1 if the value was not found
func SelectBucket(buckets []float32) int {
	randomValue := GenerateRandomValue(0, 100)
	return SelectBucketContaining(randomValue, buckets)
}

// It looks for the bucket containing the value sent as parameter
// It returns the idx of the bucket containing the value, -1 if the value was not found
func SelectBucketContaining(value int, buckets []float32) int {
	bucket := 0
	for bucket < len(buckets) && float32(value) > buckets[bucket] {
		bucket++
	}
	if bucket >= len(buckets) {
		bucket = -1
	}
	return bucket
}

func GenerateRandomValue(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}
