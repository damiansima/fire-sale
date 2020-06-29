package engine

import (
	"github.com/damiansima/fire-sale/util"
	"github.com/storozhukBM/verifier"
	"math/rand"
	"time"
)

// Provided a slice contanig values between 0.0:1.0 it generates buckets containing values between 0:100
func BuildBuckets(distribution []float32) ([]float32, error) {
	verify := verifier.New()
	verify.That(distribution != nil, "distribution can't be nil")
	verify.That(len(distribution) > 0, "distribution must not be empty")
	verify.That(util.Sum(distribution) == 1, "distribution must add up 100")
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
	randomValue := generateRandomValue(0, 100)
	return SelectDeterministicBucket(randomValue, buckets)
}

// It looks for the bucket containing the value sent as parameter
// It returns the idx of the bucket containing the value, -1 if the value was not found
func SelectDeterministicBucket(value int, buckets []float32) int {
	bucket := 0
	for bucket < len(buckets) && float32(value) > buckets[bucket] {
		bucket++
	}
	if bucket >= len(buckets) {
		bucket = -1
	}
	return bucket
}

func generateRandomValue(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}
