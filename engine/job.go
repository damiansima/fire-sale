package engine

import (
	"io"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/storozhukBM/verifier"
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
	IsWarmup             bool
	SuccessValidator     func(status int) bool
}

func (j *Job) ValidateSuccess(status int) bool {
	if j.SuccessValidator != nil {
		return j.SuccessValidator(status)
	} else {
		return status > 0 && status < 300
	}
}

// AllocateJobs creates jobs and adds them to the jobs queue
// It receives noOfJobs and testDurationMs, if the second is grated than 0 it takes precedences and keeps
// pushing jobs during the defined period. If not the specified value of jobs will be created
func AllocateJobs(noOfJobs int, noOfWarmupJobs int, testDuration time.Duration, warmupDuration time.Duration, maxSpeedPerSecond int, scenarios []Scenario, jobs chan Job) error {
	log.Debugf("Allocating jobs ...")

	distributionsBuckets, err := buildDistributionBuckets(scenarios)
	if err != nil {
		log.Fatal(err)
		return err
	}

	if testDuration > 0 {
		shouldWarmupStop := make(chan bool)
		go func() {
			if warmupDuration != 0 {
				log.Debugf("Starting warmup for [%d]ms", warmupDuration)
				<-time.After(warmupDuration)
				log.Debugf("Warmup Done")
			}
			shouldWarmupStop <- true
		}()

		shouldStop := make(chan bool)
		go func() {
			log.Debugf("Allocating for [%d]ms", testDuration)
			<-time.After(testDuration)
			log.Debugf("Stop allocation")
			shouldStop <- true
		}()

		allocateJobsUntilDone(shouldStop, shouldWarmupStop, maxSpeedPerSecond, scenarios, distributionsBuckets, jobs)
	} else {
		allocatePredefinedNumberOfJobs(noOfJobs, noOfWarmupJobs, maxSpeedPerSecond, scenarios, distributionsBuckets, jobs)
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

func allocateJobsUntilDone(shouldStop chan bool, shouldWarmupStop chan bool, maxSpeedPerSecond int, scenarios []Scenario, distributionsBuckets []float32, jobs chan Job) {
	isWarmup := true
	for i := 0; ; {
		select {
		case <-shouldStop:
			return
		case <-shouldWarmupStop:
			isWarmup = false
		default:
			if maxSpeedPerSecond > 0 {
				for j := 0; j < maxSpeedPerSecond; j++ {
					allocateJob(i, isWarmup, scenarios, distributionsBuckets, -1, jobs)
					i++
				}
				time.Sleep(1 * time.Second)
			} else {
				allocateJob(i, isWarmup, scenarios, distributionsBuckets, -1, jobs)
				i++
			}
		}

	}
}

func allocatePredefinedNumberOfJobs(noOfJobs int, noOfWarmupJobs int, maxSpeedPerSecond int, scenarios []Scenario, distributionsBuckets []float32, jobs chan Job) {
	log.Debugf("Allocating [%d]job", noOfJobs)
	isWarmup := false
	for i := 0; i < noOfJobs; {
		if noOfWarmupJobs > 0 && noOfWarmupJobs > i {
			isWarmup = true
		} else {
			isWarmup = false
		}
		if maxSpeedPerSecond > 0 {
			for j := 0; j < maxSpeedPerSecond; j++ {
				allocateJob(i, isWarmup, scenarios, distributionsBuckets, -1, jobs)
				i++
			}
			time.Sleep(1 * time.Second)
		} else {
			allocateJob(i, isWarmup, scenarios, distributionsBuckets, -1, jobs)
			i++
		}
	}
	log.Debugf("Stop allocation")
}

func allocateJob(id int, isWarmup bool, scenarios []Scenario, distributionsBuckets []float32, bucketValue int, jobs chan Job) {
	scenario, err := selectScenario(scenarios, distributionsBuckets, bucketValue)
	if err != nil {
		log.Errorf("Fail to select scenario to for job err: %v", err)
	} else {
		job := scenario.JobCreator(id)
		job.IsWarmup = isWarmup
		log.Debugf("Allocating Job id:[%d], name:[%s], isWarmup:[%t]", job.Id, job.Name, job.IsWarmup )
		jobs <- job
	}
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
		// TODO this scenario is only when you want to run scenarios in a certain order
		// TODO it needs to be fixed is not used in time based tests and in request based tests will fail over 100 reques
		scenario = scenarios[SelectBucketContaining(bucketValue, buckets)]
	}
	log.Debugf("Selecting Scenario id: [%d], name: [%s]", scenario.Id, scenario.Name)
	return scenario, nil
}
