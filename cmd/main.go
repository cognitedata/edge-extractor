package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"io/ioutil"

	"github.com/cognitedata/edge-extractor/integrations"
	"github.com/cognitedata/edge-extractor/internal"
	log "github.com/sirupsen/logrus"
)

var Version string

func main() {

	log.Infof("----- Starting edge-extractor - version = %s ----------", Version)

	mainConfigPath := flag.String("config", "config.json", "Full path to main configuration file")

	base64encodedConfig := flag.String("bconfig", "", "Base64 encoded config")
	op := flag.String("op", "", "Supported operations : 'gen_config' ")

	flag.Parse()

	var config internal.StaticConfig
	var configBody []byte
	var err error

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

	if *op == "gen_config" {
		log.Info("Generating config file")
		config.CdfCluster = "westeurope-1"
		config.Scopes = []string{"https://westeurope-1.cognitedata.com/.default"}
		config.AuthTokenUrl = "https://login.microsoftonline.com/set_your_tenant_id_here/oauth2/v2.0/token"
		config.EnabledIntegrations = []string{"ip_cams_to_cdf"}
		body, _ := json.MarshalIndent(&config, " ", "  ")
		ioutil.WriteFile("config.json", body, 0644)
		return
	}

	log.SetLevel(log.DebugLevel)

	if configBody == nil {
		configBody, err = ioutil.ReadFile(*mainConfigPath)

		if err != nil {
			log.Error("Application is not configured. Either add configuraion file or use configuraion UI to configure the application")
			// TODO : Start config ui webserver here
			return
		}
	}

	err = json.Unmarshal(configBody, &config)
	if err != nil {
		log.Error("Incorrect config file")
		// TODO : Start config ui webserver here
		return
	}

	cdfCLient := internal.NewCdfClient(config.ProjectName, config.CdfCluster, config.ClientID, config.Secret, config.Scopes, config.AdTenantId, config.AuthTokenUrl, config.CdfDatasetID)

	integrReg := make(map[string]interface{})

	for _, integrName := range config.EnabledIntegrations {
		switch integrName {
		case "ip_cams_to_cdf":
			intgr := integrations.NewCameraImagesToCdf(cdfCLient)
			err = intgr.Start()
			if err != nil {
				log.Errorf(" %s integration can't be started . Error : %s", integrName, err.Error())
			} else {
				integrReg["ip_cams_to_cdf"] = intgr
			}

		}
	}

	select {}

}
