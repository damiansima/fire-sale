package engine

import (
	"fmt"
	"testing"
)

// TODO add proper testing to this
func TestGetDistribution(t *testing.T) {
	//distribution := []float32{0.2, 0.2, 0.2, 0.2, 0.2}
	distribution := []float32{0.25, 0.25, 0.25, 0.25}
	fmt.Printf("%v \n", distribution)
	getDistribution := GetDistribution(distribution)
	fmt.Printf("%v \n", getDistribution)
	var sum float32
	for _, i := range getDistribution {
		sum += i
	}
	if getDistribution[len(getDistribution)-1] != float32(100) {
		t.Errorf("Sum must be excatly 100 %v", getDistribution[len(getDistribution)])
	}
}

func TestSelectBucket(t *testing.T) {
	distribution := []float32{0.25, 0.25, 0.25, 0.25}
	fmt.Printf("%v \n", distribution)
	getDistribution := GetDistribution(distribution)
	fmt.Printf("%v \n", getDistribution)

	bucket := SelectBucket(getDistribution)
	fmt.Println("")
	fmt.Printf("bucket %d", bucket)
	fmt.Println("")
}
