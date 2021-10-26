package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cognitedata/edge-extractor/integrations"
	"github.com/cognitedata/edge-extractor/internal"
	"github.com/kardianos/service"
	log "github.com/sirupsen/logrus"
)

var Version string
var systemLog service.Logger
var fullConfigPath string
var integrReg map[string]interface{}

type program struct{}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

func (p *program) run() {
	// Do work here
	systemLog.Info("----Starting edge extractor service-------")
	systemLog.Infof("Loading configuration from file %s", fullConfigPath)
	startEdgeExtractor(fullConfigPath)
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	systemLog.Info("----Stoping edge extractor service-------")
	stopExtractor()
	return nil
}

func configureService() service.Service {
	svcConfig := service.Config{Name: "cog-edge-extractor", DisplayName: "Cognite edge extractor", Description: "Cognite edge extractor service"}
	var prg program
	var err error
	var appService service.Service

	appService, err = service.New(&prg, &svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	systemLog, err = appService.Logger(nil)
	systemLog.Info("Starting edge extractor service")

	if err != nil {
		fmt.Printf("Error initializing system logger %s", err.Error())
	}

	return appService

}

func configureLogger(logPath, level string) {
	if level == "" {
		level = "info"
	}
	lvl, _ := log.ParseLevel(level)
	log.SetLevel(lvl)
	log.SetFormatter(&log.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
	// open a file
	logPath = filepath.Join(logPath, "edge-extractor.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		systemLog.Error("Failed to create log , err :" + err.Error())
	}
	log.SetOutput(f)

}

func startEdgeExtractor(mainConfigPath string) {
	var config internal.StaticConfig
	configBody, err := ioutil.ReadFile(mainConfigPath)

	if err != nil {
		log.Error("Application is not configured. Either add configuraion file or use configuraion UI to configure the application")
		systemLog.Error("Failed to load config file. Err:", err.Error())
		// TODO : Start config ui webserver here
		return
	}

	err = json.Unmarshal(configBody, &config)
	if err != nil {
		log.Error("Incorrect config file")
		systemLog.Error("Incorrect config file format. Err:", err.Error())
		// TODO : Start config ui webserver here
		return
	}

	configureLogger(internal.GetBinaryDir(), config.LogLevel)

	cdfCLient := internal.NewCdfClient(config.ProjectName, config.CdfCluster, config.ClientID, config.Secret, config.Scopes, config.AdTenantId, config.AuthTokenUrl, config.CdfDatasetID)

	integrReg = make(map[string]interface{})

	for _, integrName := range config.EnabledIntegrations {
		switch integrName {
		case "ip_cams_to_cdf":
			intgr := integrations.NewCameraImagesToCdf(cdfCLient, config.ExtractionMonitoringID)
			err = intgr.Start()
			if err != nil {
				log.Errorf(" %s integration can't be started . Error : %s", integrName, err.Error())
			} else {
				integrReg["ip_cams_to_cdf"] = intgr
			}

		}
	}
}

func stopExtractor() {
	intgr := integrReg["ip_cams_to_cdf"].(*integrations.CameraImagesToCdf)
	intgr.Stop()
}

func main() {

	log.Infof("----- Starting edge-extractor - version = %s ----------", Version)

	mainConfigPath := flag.String("config", "config.json", "Full path to main configuration file")

	base64encodedConfig := flag.String("bconfig", "", "Base64 encoded config")
	op := flag.String("op", "", "Supported operations : 'gen_config,install,uninstall,run' ")

	flag.Parse()

	if *mainConfigPath == "config.json" {
		*mainConfigPath = filepath.Join(internal.GetBinaryDir(), *mainConfigPath)
		fullConfigPath = *mainConfigPath
	}

	var config internal.StaticConfig

	// User can configure app by passing configurations as one base64 encoded string
	if *base64encodedConfig != "" {
		log.Info("Loading configuration from cmd line parameter")
		// Base64 Standard Decoding
		body, err := base64.StdEncoding.DecodeString(*base64encodedConfig)
		if err != nil {
			log.Error("Error decoding base64 encoded config: %s ", err.Error())
			return
		}
		ioutil.WriteFile("config.json", body, 0644)
	}

	switch *op {
	case "gen_config":
		log.Info("Generating config file")
		config.CdfCluster = "westeurope-1"
		config.Scopes = []string{"https://westeurope-1.cognitedata.com/.default"}
		config.AuthTokenUrl = "https://login.microsoftonline.com/set_your_tenant_id_here/oauth2/v2.0/token"
		config.EnabledIntegrations = []string{"ip_cams_to_cdf"}
		body, _ := json.MarshalIndent(&config, " ", "  ")
		ioutil.WriteFile("config.json", body, 0644)
		return
	case "install":
		log.Info("Installing edge-extractor service")
		appService := configureService()
		err := appService.Install()
		if err != nil {
			log.Error("Failed to install service.Make sure you run installation as system administrator Err: ", err.Error())
		} else {
			err = appService.Start()
			if err != nil {
				log.Error("Failed to run service. Err: ", err.Error())
			}
		}
	case "uninstall":
		log.Info("Uninstalling edge-extractor service")
		appService := configureService()
		err := appService.Uninstall()

		if err != nil {
			log.Error("Failed to uninstall service", err.Error())
		}
	case "run":
		// Should be used to start service from CLI
		startEdgeExtractor(*mainConfigPath)
		select {}
	default:
		// Used by OS service supervisor
		appService := configureService()
		err := appService.Run()
		if err != nil {
			log.Error(err)
		}
	}

}
