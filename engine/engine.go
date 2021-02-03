package engine

import (
	"errors"
	"fmt"
	"github.com/storozhukBM/verifier"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Configuration struct {
	Name         string
	Parameters   Parameters
	Certificates Certificates
	Scenarios    []Scenario
}

type Parameters struct {
	NoOfRequest       int
	NoOfWarmupRequest int
	TestDuration      time.Duration
	WarmupDuration    time.Duration
	Workers           int
	MaxRequest        int
	RampUp            RampUp
}

// Defines the way the engine will get to the defined amount of workers
type RampUp struct {
	Step int
	Time time.Duration
}

type Certificates struct {
	ClientCertFile string
	ClientKeyFile  string
	CaCertFile     string
}

type Scenario struct {
	Id           int
	Name         string
	Distribution float32
	JobCreator   func(id int) Job
}


var DefaultRampUp RampUp = RampUp{Step: 1, Time: 0}

func (c *Configuration) init() error {
	log.Debugf("Initializing configuration ...")
	var err error
	c.initRampUp()
	err = c.balanceScenarioDistribution()
	return err
}

func (c *Configuration) initRampUp() {
	log.Debugf("Initializing ramp up ...")
	if (RampUp{}) == c.Parameters.RampUp {
		log.Debugf("Ramp up not defined defaulting")
		c.Parameters.RampUp = DefaultRampUp
	}
}

func (c *Configuration) balanceScenarioDistribution() error {
	log.Debugf("Balancing out scenario distribution ...")
	balancedScenarios, err := balanceScenarioDistribution(c.Scenarios)
	if err != nil {
		return err
	}
	c.Scenarios = balancedScenarios
	return nil
}

func balanceScenarioDistribution(scenarios []Scenario) ([]Scenario, error) {
	var balancedScenarios []Scenario

	verify := verifier.New()
	verify.That(len(scenarios) > 0, "Scenarios must not be empty")
	if verify.GetError() != nil {
		return scenarios, verify.GetError()
	}

	remainingDistribution := float32(1)
	scenariosWithNoDistribution := []Scenario{}
	for _, scenario := range scenarios {
		if scenario.Distribution == 0 {
			scenariosWithNoDistribution = append(scenariosWithNoDistribution, scenario)
		} else {
			balancedScenarios = append(balancedScenarios, scenario)
			remainingDistribution -= scenario.Distribution
		}
	}

	if remainingDistribution < 0 || (remainingDistribution == 0 && len(scenariosWithNoDistribution) > 0) {
		message := fmt.Sprintf("Scenarios Distribution should add up to 1 ")
		err := errors.New(message)
		log.Error(err)
		return scenarios, err
	}

	if remainingDistribution > 0 {
		assignedDistribution := remainingDistribution / float32(len(scenariosWithNoDistribution))
		log.Debugf("Remaining distrubition %.2f. It will be assigned evently with %.2f", remainingDistribution, assignedDistribution)
		for _, scenario := range scenariosWithNoDistribution {
			log.Debugf("Assigning distrubition %.2f to scenario id %d, name: %s ", assignedDistribution, scenario.Id, scenario.Name)
			scenario.Distribution = assignedDistribution
			balancedScenarios = append(balancedScenarios, scenario)
		}
	} else {
		log.Debugf("Remaining distrubition %.2f. Doing nothing", remainingDistribution)
	}
	return balancedScenarios, nil
}

func ConfigureLog(logLevel string) {
	log.SetFormatter(&log.TextFormatter{})

	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)
	}
}

func Run(config Configuration, reportType, reportFilePath string) {
	var err error

	err = config.init()
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Parameters - # of Request [%d] - Test Duration [%s] - Warm up Request [%d] - Warmup Duration [%s] - Concurrent Users [%d] - Max RPS [%d] - Ramp Up [%v]", config.Parameters.NoOfRequest, config.Parameters.TestDuration, config.Parameters.NoOfWarmupRequest, config.Parameters.WarmupDuration, config.Parameters.Workers, config.Parameters.MaxRequest, config.Parameters.RampUp)
	start := time.Now()

	jobBufferSize := 15
	resultBufferSize := 1000 * config.Parameters.Workers
	jobs := make(chan Job, jobBufferSize)
	results := make(chan Result, resultBufferSize)

	go AllocateJobs(config.Parameters.NoOfRequest, config.Parameters.NoOfWarmupRequest, config.Parameters.TestDuration, config.Parameters.WarmupDuration, config.Parameters.MaxRequest, config.Scenarios, jobs)

	done := make(chan bool)
	report := Report{}
	go ConsumeResults(results, done, &report)

	runWorkers(config.Parameters.Workers, config.Parameters.RampUp, config.Certificates, jobs, results)
	<-done

	printReport(report, reportType, reportFilePath)
	log.Infof("Execution took [%.2fs]", time.Now().Sub(start).Seconds())
}

func runWorkers(noOfWorkers int, rampUp RampUp, certificates Certificates, jobs chan Job, results chan Result) {
	log.Infof("Running [%d] concurrent workers ...", noOfWorkers)
	var wg sync.WaitGroup

	// TODO BUG rampUp.Step can not be < 0
	// TODO BUG rampUp.Step can not be > noOfWorkers
	// TODO test: step can't go over, noOfWorker
	steps := noOfWorkers / rampUp.Step
	pace := time.Duration(rampUp.Time.Nanoseconds() / int64(steps))

	log.Debugf("Ramping up in [%d] steps...", steps)
	for i := 0; i < noOfWorkers; {
		for s := 0; i < noOfWorkers && s < rampUp.Step; s++ {
			log.Debugf("Starting worker [%d]  ...", i)
			wg.Add(1)
			go work(i, &wg, jobs, results, certificates)
			i++
		}
		log.Debugf("Pacing for [%s] ...", pace)
		time.Sleep(pace)
	}
	wg.Wait()
	close(results)
	log.Infof("Workers finish job pool")
}

func work(workerId int, wg *sync.WaitGroup, jobs chan Job, results chan Result, certificates Certificates) {
	for job := range jobs {
		log.Debugf("Worker [%d] running job [%d] ...", workerId, job.Id)
		result := DoRequest(job.Method, job.Url, job.ReqBody, job.Headers, job.Timeout, job.AllowConnectionReuse, certificates)
		result.job = job
		results <- result
	}
	wg.Done()
}
