package main

import (
	"flag"
	"github.com/damiansima/fire-sale/dsl"
	"github.com/damiansima/fire-sale/engine"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	configPathPtr := flag.String("config", "", "Path to the test-configuration.yml")
	logLevelPtr := flag.String("log", "info", "Define the log level [panic|fatal|error|warn|info|debug|trace]")
	reportTypePtr := flag.String("report-type", "std", "Define the report type [std|json]")
	reportFilePathPtr := flag.String("report-path", "", "Define the report file path. If not provided it'll be printed to stdout")
	flag.Parse()

	if *configPathPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	engine.ConfigureLog(*logLevelPtr)

	log.Info("Everything must Go...")
	run(*configPathPtr, *reportTypePtr, *reportFilePathPtr)
	log.Info("[¡¡¡SOLD!!!]")
}

func run(configPath, reportType, reportFilePath string) {
	log.Infof("Running %s ...", configPath)
	configuration := dsl.ParseConfiguration(configPath)

	config := engine.Configuration{
		Name: configuration.Name,
		Parameters: engine.Parameters{
			NoOfRequest:       configuration.Parameters.NoOfRequest,
			NoOfWarmupRequest: configuration.Parameters.NoOfWarmupRequest,
			TestDuration:      dsl.ParseDuration(configuration.Parameters.TestDuration),
			WarmupDuration:    dsl.ParseDuration(configuration.Parameters.WarmupDuration),
			Workers:           configuration.Parameters.Workers,
			MaxRequest:        configuration.Parameters.MaxRequest,
			RampUp:            dsl.MapRampUp(configuration),
		},
		Certificates: dsl.MapCertificates(configuration),
		Scenarios:    dsl.MapScenarios(configuration),
	}

	engine.Run(config, reportType, reportFilePath)
}
