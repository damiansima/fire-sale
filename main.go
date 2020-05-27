package main

import (
	"flag"
	"github.com/damiansima/fire-sale/dsl"
	"github.com/damiansima/fire-sale/engine"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

func main() {
	configPathPtr := flag.String("config", "", " Path to the test-configuration.yml")
	logLevelPtr := flag.String("log", "info", " Define the log level [panic|fatal|error|warn|info|debug|trace]")
	flag.Parse()

	if *configPathPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	engine.ConfigureLog(*logLevelPtr)

	log.Info("Everything must Go...")
	run(*configPathPtr)
	log.Info("[¡¡¡SOLD!!!]")
}

func run(configPath string) {
	log.Infof("Running %s ...", configPath)
	configuration := dsl.ParseConfiguration(configPath)
	testDuration := time.Duration(configuration.Parameters.TestDuration) * time.Minute

	engine.Run(configuration.Parameters.Workers, configuration.Parameters.NoOfRequest, testDuration, configuration.Parameters.MaxRequest, dsl.MapScenarios(configuration), dsl.MapRampUp(configuration), dsl.MapCertificates(configuration))
}
