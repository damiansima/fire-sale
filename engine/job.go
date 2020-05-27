package engine

import (
	log "github.com/sirupsen/logrus"
	"github.com/storozhukBM/verifier"
	"io"
	"time"
)

// TODO this probably needs a new name
type Job struct {
	Id                   int
	Name                 string
	ScenarioId           int
	Method               string
	Url                  string
	ReqBody              io.Reader
	Headers              map[string]string
	Timeout              time.Duration
	AllowConnectionReuse bool
}

// AllocateJobs creates jobs and adds them to the jobs queue
// It receives noOfJobs and testDurationMs, if the second is grated than 0 it takes precedences and keeps
// pushing jobs during the defined period. If not the specified value of jobs will be created
func AllocateJobs(noOfJobs int, testDuration time.Duration, maxSpeedPerSecond int, scenarios []Scenario, jobs chan Job) error {
	log.Debugf("Allocating jobs ...")

	distributionsBuckets, err := buildDistributionBuckets(scenarios)
	if err != nil {
		log.Fatal(err)
		return err
	}

	if testDuration > 0 {
		shouldStop := make(chan bool)
		go func() {
			log.Debugf("Allocating for [%d]ms", testDuration)
			<-time.After(testDuration)
			log.Debugf("Stop allocation")
			shouldStop <- true
		}()
		allocateJobsUntilDone(shouldStop, maxSpeedPerSecond, scenarios, distributionsBuckets, jobs)
	} else {
		allocatePredefinedNumberOfJobs(noOfJobs, maxSpeedPerSecond, scenarios, distributionsBuckets, jobs)
	}

	close(jobs)
	log.Debugf("Allocating done")
	return nil
}

func buildDistributionBuckets(scenarios []Scenario) ([]float32, error) {
	verify := verifier.New()
	verify.That(len(scenarios) > 0, "Scenarios must not be empty")
	if verify.GetError() != nil {
		return nil, verify.GetError()
	}

	scenarioDistributions := make([]float32, len(scenarios))
	for i := 0; i < len(scenarios); i++ {
		scenarioDistributions[i] = scenarios[i].Distribution
	}

	distributions, err := BuildBuckets(scenarioDistributions)
	if err != nil {
		return nil, err
	}
	return distributions, nil
}

func allocateJobsUntilDone(shouldStop chan bool, maxSpeedPerSecond int, scenarios []Scenario, distributionsBuckets []float32, jobs chan Job) {
	for i := 0; ; {
		select {
		case <-shouldStop:
			return
		default:
			if maxSpeedPerSecond > 0 {
				for j := 0; j < maxSpeedPerSecond; j++ {
					allocateJob(i, scenarios, distributionsBuckets, -1, jobs)
					i++
				}
				time.Sleep(1 * time.Second)
			} else {
				allocateJob(i, scenarios, distributionsBuckets, -1, jobs)
				i++
			}
		}

	}
}

func allocatePredefinedNumberOfJobs(noOfJobs int, maxSpeedPerSecond int, scenarios []Scenario, distributionsBuckets []float32, jobs chan Job) {
	log.Debugf("Allocating [%d]job", noOfJobs)
	for i := 0; i < noOfJobs; {
		if maxSpeedPerSecond > 0 {
			for j := 0; j < maxSpeedPerSecond; j++ {
				allocateJob(i, scenarios, distributionsBuckets, i, jobs)
				i++
			}
			time.Sleep(1 * time.Second)
		} else {
			allocateJob(i, scenarios, distributionsBuckets, i, jobs)
			i++
		}
	}
	log.Debugf("Stop allocation")
}

func selectScenario(scenarios []Scenario, buckets []float32, bucketValue int) (Scenario, error) {
	verify := verifier.New()
	verify.That(len(scenarios) > 0, "Scenarios must not be empty")
	verify.That(len(buckets) > 0, "Buckets must not be empty")
	verify.That(len(scenarios) == len(buckets), "Scenarios & Buckets must be the same size")
	if verify.GetError() != nil {
		return Scenario{}, verify.GetError()
	}

	var scenario Scenario
	if bucketValue < 0 {
		scenario = scenarios[SelectBucket(buckets)]
	} else {
		scenario = scenarios[SelectBucketContaining(bucketValue, buckets)]
	}
	log.Debugf("Selecting Scenario %s", scenario.Name)
	return scenario, nil
}

func allocateJob(id int, scenarios []Scenario, distributionsBuckets []float32, bucketValue int, jobs chan Job) {
	scenario, err := selectScenario(scenarios, distributionsBuckets, bucketValue)
	if err != nil {
		log.Errorf("Fail to select scenario to for job err: %v", err)
	} else {
		log.Debugf("Allocating job [%d]", id)
		jobs <- scenario.JobCreator(id)
	}
}
