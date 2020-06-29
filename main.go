package main

import (
	"bytes"
	"github.com/damiansima/fire-sale/engine"
	log "github.com/sirupsen/logrus"
	"time"
)

func init() {
	configureLog("info")
}

const defaultTimeout = 60000 * time.Millisecond

func main() {
	log.Info("Everything must Go[lang] ...")

	noOfRequest := 1
	testDuration := 0 * time.Minute

	noOfWorkers := 1
	maxRequestPerSecond := 0

	rampUp := engine.RampUp{Step: 1, Time: 0 * time.Minute}

	log.Infof("Parameters - # of Request [%d] - Test Duration [%s] - Concurrent Users [%d] - Max RPS [%d] - Ramp Up [%v]", noOfRequest, testDuration, noOfWorkers, maxRequestPerSecond, rampUp)

	scenarios := defineScenarios()
	run(noOfWorkers, noOfRequest, testDuration, maxRequestPerSecond, scenarios, rampUp)
	log.Info("[¡¡¡SOLD!!!]")
}

func defineScenarios() []engine.Scenario {
	var scenarios []engine.Scenario
	// TODO define a unique way to id scenarios
	s0Id := 0
	s0 := engine.Scenario{
		Id:           s0Id,
		Name:         "Basic Scenario 0",
		Distribution: 1,
		JobCreator: func(id int) engine.Job {
			var method string
			var basePath string
			var timeout = defaultTimeout
			headers := make(map[string]string)
			var bodyBuffer *bytes.Buffer

			method = "GET"
			bodyBuffer = bytes.NewBuffer([]byte(""))
			basePath = "https://www.infobae.com"

			return engine.Job{Id: id, ScenarioId: s0Id, Method: method, Url: basePath, ReqBody: bodyBuffer, Headers: headers, Timeout: timeout, AllowConnectionReuse: true}
		},
	}

	scenarios = append(scenarios, s0)

	return scenarios
}

func configureLog(logLevel string) {
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

func run(noOfWorkers int, noOfRequest int, testDuration time.Duration, maxSpeedPerSecond int, scenarios []engine.Scenario, rampUp engine.RampUp) {
	start := time.Now()

	jobBufferSize := 15
	resultBufferSize := 1000 * noOfWorkers
	jobs := make(chan engine.Job, jobBufferSize)
	results := make(chan engine.Result, resultBufferSize)

	go engine.AllocateJobs(noOfRequest, testDuration, maxSpeedPerSecond, scenarios, jobs)

	done := make(chan bool)
	go engine.ConsumeResults(results, done)

	if (engine.RampUp{}) == rampUp {
		rampUp = engine.DefaultRampUp
	}
	engine.RunWorkers(noOfWorkers, rampUp, jobs, results)
	<-done

	log.Infof("Execution toke [%s]", time.Now().Sub(start))
}
