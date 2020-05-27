package dsl

import (
	"bytes"
	"github.com/damiansima/fire-sale/engine"
	"github.com/damiansima/fire-sale/processor"
	"io/ioutil"
	"os"
	"time"
)

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Configuration struct {
	Name       string
	Host       string
	Parameters struct {
		NoOfRequest  int
		TestDuration int
		Workers      int
		MaxRequest   int
		RampUp       struct {
			Step int
			Time int
		}
	}
	Certificates struct {
		ClientCertFile string
		ClientKeyFile  string
		CaCertFile     string
	}
	Scenarios []Scenario
}

type Scenario struct {
	Name         string
	Distribution float32
	Timeout      int
	Method       string
	Host         string
	Path         string
	Headers      map[string]string
	Body         string
}

func ParseConfiguration(configPath string) Configuration {
	configFile, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}
	defer configFile.Close()
	cfBytes, _ := ioutil.ReadAll(configFile)

	var configuration Configuration
	err = yaml.Unmarshal(cfBytes, &configuration)
	if err != nil {
		panic(err)
	}
	return configuration
}

func MapRampUp(configuration Configuration) engine.RampUp {
	// TODO define default ram up if empty
	return engine.RampUp{
		Step: configuration.Parameters.RampUp.Step,
		Time: time.Duration(configuration.Parameters.RampUp.Time) * time.Minute,
	}
}

func MapCertificates(configuration Configuration) engine.Certificates {
	return engine.Certificates{
		ClientCertFile: configuration.Certificates.ClientCertFile,
		ClientKeyFile:  configuration.Certificates.ClientKeyFile,
		CaCertFile:     configuration.Certificates.CaCertFile,
	}

}

func MapScenarios(configuration Configuration) []engine.Scenario {
	var engineScenarios []engine.Scenario
	for i, s := range configuration.Scenarios {
		engineScenarios = append(engineScenarios, mapScenario(i, s, configuration.Host))
	}
	return engineScenarios
}

func mapScenario(scId int, dslSc Scenario, host string) engine.Scenario {
	basePath := ""
	if dslSc.Host != "" {
		basePath = dslSc.Host
	} else {
		basePath = host
	}
	// TODO create Scenario builder object with context use it here and in the main so we have a default way to create scenarios and job  creators with processor
	return engine.Scenario{
		Id:           scId,
		Name:         dslSc.Name,
		Distribution: dslSc.Distribution,
		JobCreator: func(id int) engine.Job {
			url, err := processor.Process(basePath + dslSc.Path)
			body, err := processor.Process(dslSc.Body)

			// TODO fix this it should fail even before it start running
			if err != nil {
				log.Fatalf("Fail to process scenario  [%d,%s] -  Error:  %v", scId, dslSc.Name, err)
			}

			return engine.Job{
				Id:                   id,
				Name:                 dslSc.Name,
				ScenarioId:           scId,
				Method:               dslSc.Method,
				Url:                  url,
				ReqBody:              bytes.NewBuffer([]byte(body)),
				Headers:              dslSc.Headers,
				Timeout:              time.Duration(dslSc.Timeout),
				AllowConnectionReuse: false, // TODO is this correct?
			}
		},
	}
}
