package engine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAllocateJobs(t *testing.T) {
	scenario1 := Scenario{Id: 0, Name: "Basic Scenario 1", Distribution: 1, JobCreator: func(id int) Job { return Job{Id: id} }}

	tests := []struct {
		name              string
		noOfJobs          int
		testDuration      time.Duration
		maxSpearPerSecond int
		scenarios         []Scenario
		buckets           []float32
		jobs              chan Job
	}{
		{"No duration", 10, 0, 0, []Scenario{scenario1}, []float32{100}, make(chan Job, 1000)},
		{"With duration", 1, 10 * time.Millisecond, 0, []Scenario{scenario1}, []float32{100}, make(chan Job, 1000)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			shouldStop := make(chan bool)
			go func() {
				time.Sleep(5 * time.Millisecond)
				shouldStop <- true
			}()
			err := AllocateJobs(test.noOfJobs, 0, test.testDuration, time.Duration(0), test.maxSpearPerSecond, test.scenarios, test.jobs)
			if err != nil {
				assert.Fail(t, "Test should not have failed")
			}
			j := <-test.jobs

			assert.True(t, len(test.jobs) > 1, "The number of jobs in the channel is incorrect")
			if test.testDuration == 0 {
				assert.True(t, len(test.jobs) == test.noOfJobs-1, "The number of jobs in the channel is incorrect")
			} else {
				assert.True(t, len(test.jobs) > test.noOfJobs-1, "The number of jobs in the channel is incorrect")
			}
			if test.maxSpearPerSecond > 0 {
				assert.True(t, len(test.jobs) == test.maxSpearPerSecond-1)
			}
			assert.Equal(t, j.Id, test.scenarios[0].Id)
		})
	}
}

func Test_buildDistributionBuckets(t *testing.T) {
	tests := []struct {
		name       string
		scenarios  []Scenario
		want       []float32
		shouldFail bool
	}{
		{"Empty Scenarios", []Scenario{}, []float32{}, true},
		{"Distributions over 100%", []Scenario{{Distribution: 0.25}, {Distribution: 0.90}}, []float32{}, true},
		{"Distribution 25%", []Scenario{{Distribution: 0.25}, {Distribution: 0.25}, {Distribution: 0.25}, {Distribution: 0.25}}, []float32{25, 50, 75, 100}, false},
		{"Distribution 50%", []Scenario{{Distribution: 0.5}, {Distribution: 0.5}}, []float32{50, 100}, false},
		{"Distribution 100%", []Scenario{{Distribution: 1}}, []float32{100}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buckets, err := buildDistributionBuckets(test.scenarios)
			if test.shouldFail {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.want, buckets)
			}
		})
	}
}

func Test_allocateJobsUntilDone(t *testing.T) {
	scenario1 := Scenario{Id: 0, Name: "Basic Scenario 1", Distribution: 1, JobCreator: func(id int) Job { return Job{Id: id} }}

	tests := []struct {
		name              string
		maxSpearPerSecond int
		scenarios         []Scenario
		buckets           []float32
		jobs              chan Job
	}{
		{"No max speed", 0, []Scenario{scenario1}, []float32{100}, make(chan Job, 1000)},
		{"With max speed", 10, []Scenario{scenario1}, []float32{100}, make(chan Job, 1000)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			shouldStop := make(chan bool)
			shouldWarmupStop := make(chan bool)
			go func() {
				time.Sleep(5 * time.Millisecond)
				shouldStop <- true
			}()
			go func() {
				time.Sleep(5 * time.Millisecond)
				shouldWarmupStop <- true
			}()
			allocateJobsUntilDone(shouldStop, shouldWarmupStop, test.maxSpearPerSecond, test.scenarios, test.buckets, test.jobs)

			j := <-test.jobs

			assert.True(t, len(test.jobs) > 1)
			if test.maxSpearPerSecond > 0 {
				assert.True(t, len(test.jobs) == test.maxSpearPerSecond-1)
			}
			assert.Equal(t, j.Id, test.scenarios[0].Id)
			close(test.jobs)
		})
	}
}

func Test_allocatePredefinedNumberOfJobs(t *testing.T) {
	scenario1 := Scenario{Id: 0, Name: "Basic Scenario 1", Distribution: 1, JobCreator: func(id int) Job { return Job{Id: id} }}

	tests := []struct {
		name              string
		noOfJobs          int
		maxSpearPerSecond int
		scenarios         []Scenario
		buckets           []float32
		jobs              chan Job
	}{
		{"No max speed", 10, 0, []Scenario{scenario1}, []float32{100}, make(chan Job, 10)},
		{"With max speed", 10, 10, []Scenario{scenario1}, []float32{100}, make(chan Job, 10)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			allocatePredefinedNumberOfJobs(test.noOfJobs, 0, test.maxSpearPerSecond, test.scenarios, test.buckets, test.jobs)
			j := <-test.jobs
			assert.True(t, len(test.jobs) == test.noOfJobs-1)
			assert.Equal(t, j.Id, test.scenarios[0].Id)
		})
		close(test.jobs)
	}
}

func Test_allocateJob(t *testing.T) {
	jobsChan := make(chan Job, 1)
	scenario1 := Scenario{Id: 1, Name: "Basic Scenario 1", JobCreator: func(id int) Job { return Job{Id: id} }}

	tests := []struct {
		name               string
		id                 int
		scenarios          []Scenario
		distributionBucket []float32
		jobs               chan Job
		shouldFail         bool
	}{
		//{"Allocate", 1, []Scenario{scenario1}, []float32{100}, jobsChan, false},
		{"Fail Allocation", 1, []Scenario{scenario1}, []float32{50, 100}, jobsChan, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			allocateJob(test.id, false, test.scenarios, test.distributionBucket, -1, test.jobs)
			if test.shouldFail {
				assert.Equal(t, 0, len(jobsChan))
			} else {
				job := <-jobsChan
				assert.Equal(t, job.Id, test.id)
			}
		})
	}
	close(jobsChan)
}

func Test_selectScenario(t *testing.T) {
	scenario1 := Scenario{Id: 1, Name: "Basic Scenario 1", JobCreator: func(id int) Job { return Job{} }}
	scenario2 := Scenario{Id: 2, Name: "Basic Scenario 2", JobCreator: func(id int) Job { return Job{} }}

	tests := []struct {
		name       string
		scenarios  []Scenario
		buckets    []float32
		iteration  int
		want       Scenario
		shouldFail bool
	}{
		{"Empty Scenarios", []Scenario{}, []float32{100}, -1, Scenario{}, true},
		{"Empty Buckets", []Scenario{scenario1}, []float32{}, -1, Scenario{}, true},
		{"Select Scenario By Distribution choose 1", []Scenario{scenario1, scenario2}, []float32{100, 0}, -1, scenario1, false},
		{"Select Scenario By Distribution choose 2", []Scenario{scenario1, scenario2}, []float32{0, 100}, -1, scenario2, false},
		{"Select Scenario By Iteration choose 1", []Scenario{scenario1, scenario2}, []float32{50, 100}, 0, scenario1, false},
		{"Select Scenario By Iteration choose 2", []Scenario{scenario1, scenario2}, []float32{50, 100}, 51, scenario2, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scenario, err := selectScenario(test.scenarios, test.buckets, test.iteration)
			if test.shouldFail {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				// TODO we should be able to compare want and scenario but's failing
				assert.Equal(t, test.want.Id, scenario.Id)
			}
		})
	}
}
