package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/cognitedata/edge-extractor/apps/core"
	"github.com/cognitedata/edge-extractor/integrations/ip_cams_to_cdf"
	"github.com/cognitedata/edge-extractor/internal"
	"github.com/cskr/pubsub/v2"
	"github.com/kardianos/service"
	log "github.com/sirupsen/logrus"
)

var Version string
var EncryptionKey = ""
var fullConfigPath string
var systemLog service.Logger
var runMode = "service"

type Integration interface {
	Start() error
	Stop()
}

var integrReg map[string]Integration
var appManager *core.AppManager

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

// configureService configures the edge extractor as OS service. Applicable when running on Windows or Linux as service.
func configureService(isInstallOperation bool) service.Service {
	svcConfig := service.Config{
		Name:        "edge-extractor",
		DisplayName: "Cognite edge extractor",
		Description: "Cognite edge extractor service",
	}
	var prg program
	var err error
	var appService service.Service
	if runtime.GOOS == "linux" && isInstallOperation {
		err = internal.PrepareLinuxServiceEnv()
		if err != nil {
			log.Error("Failed to prepare edge-extractor environment. Err:", err.Error())
			os.Exit(1)
		}
		svcConfig.Executable = internal.LINUX_BIN
		svcConfig.Arguments = []string{"--config", internal.LINUX_CONFIG_FILE}
		svcConfig.UserName = internal.LINUX_USER
	}
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

func configureLogger(logDir, level string) {
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
	var logPath string
	if logDir != "" && logDir != "-" {
		logPath = filepath.Join(logPath, "edge-extractor.log")
	} else if runtime.GOOS == "linux" && runMode == "service" {
		// linux service must write logs to /var/log/edge-extractor directory
		logPath = filepath.Join(internal.LINUX_LOG_DIR, "edge-extractor.log")
	} else if runtime.GOOS == "windows" && runMode == "service" {
		// windows service must write logs to C:\Cognite\EdgeExtractor directory
		logPath = filepath.Join(internal.GetBinaryDir(), "edge-extractor.log")
	}
	if logPath != "" {
		fmt.Println("Log file path : ", logPath)
		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			fmt.Printf("error opening file: %v", err)
			systemLog.Error("Failed to create log , err :" + err.Error())
		}
		log.SetOutput(f)
	}
}

func encryptSecretsInConfig(configPath string) error {
	var config internal.StaticConfig
	configBody, err := os.ReadFile(configPath)

	if err != nil {
		fmt.Print("Failed to load config file. Err:", err.Error())
		return err
	}

	err = json.Unmarshal(configBody, &config)
	if err != nil {
		fmt.Print("Incorrect config file format. Err:", err.Error())
		return err
	}
	secretManager := internal.NewSecretManager(EncryptionKey)
	secretManager.LoadSecrets(config.Secrets)
	config.Secrets, err = secretManager.GetEncryptedSecrets()
	if err != nil {
		fmt.Print("Failed to encrypt config file. Err:", err)
		return err
	}
	config.IsEncrypted = true
	body, _ := json.MarshalIndent(&config, " ", "  ")
	os.WriteFile("config_encrypted.json", body, 0644)
	return nil
}

func loadStaticConfigFromEnv() internal.StaticConfig {
	var config internal.StaticConfig
	var err error
	config.ProjectName = os.Getenv("EDGE_EXT_CDF_PROJECT")
	config.CdfCluster = os.Getenv("EDGE_EXT_CDF_CLUSTER")
	config.Scopes = []string{os.Getenv("EDGE_EXT_CDF_SCOPES")}
	config.AuthTokenUrl = os.Getenv("EDGE_EXT_CDF_AUTH_TOKEN_URL")
	config.ClientID = os.Getenv("EDGE_EXT_CDF_CLIENT_ID")
	config.Secret = os.Getenv("EDGE_EXT_CDF_CLIENT_SECRET")
	config.AdTenantId = os.Getenv("EDGE_EXT_CDF_AD_TENANT_ID")
	config.CdfDatasetID, err = strconv.Atoi(os.Getenv("EDGE_EXT_CDF_DATASET_ID"))
	if err != nil {
		log.Error("Failed to parse CDF dataset ID. Err:", err.Error())
	}
	config.ExtractorID = os.Getenv("EDGE_EXT_EXTRACTOR_ID")
	config.RemoteConfigSource = os.Getenv("EDGE_EXT_CONFIG_SOURCE")

	configReloadInterval, err := strconv.Atoi(os.Getenv("EDGE_EXT_CONFIG_RELOAD_INTERVAL"))
	if err != nil {
		log.Error("Failed to parse config reload interval. Err:", err.Error())
		configReloadInterval = 60
	}
	config.ConfigReloadInterval = time.Duration(configReloadInterval)
	config.EnabledIntegrations = []string{os.Getenv("EDGE_EXT_ENABLED_INTEGRATIONS")}
	config.LogLevel = os.Getenv("EDGE_EXT_LOG_LEVEL")
	config.LogDir = os.Getenv("EDGE_EXT_LOG_DIR")
	return config
}

func startEdgeExtractor(mainConfigPath string) {
	var config internal.StaticConfig
	var err error
	if os.Getenv("EDGE_EXT_CDF_PROJECT") != "" {
		log.Info("Loading configuration from ENV variables ")
		config = loadStaticConfigFromEnv()
	} else {
		log.Info("Loading configuration from file ", mainConfigPath)
		configBody, err := os.ReadFile(mainConfigPath)
		if err != nil {
			systemLog.Error("Failed to load config file. Err:", err.Error())
			// TODO : Start config ui webserver here
			return
		}
		err = json.Unmarshal(configBody, &config)
		if err != nil {
			systemLog.Error("Incorrect config file format. Err:", err.Error())
			// TODO : Start config ui webserver here
			return
		}
	}

	configureLogger(config.LogDir, config.LogLevel)

	log.Info("Starting edge-extractor service. Version : ", Version)
	secretManager := internal.NewSecretManager(EncryptionKey)
	secretManager.LoadEncryptedSecrets(config.Secrets)
	clientSecret := secretManager.GetSecret(config.Secret)
	if clientSecret == "" {
		log.Error("Client secret is not set. Please set it in config file or in environment variable")
		return
	}
	cdfCLient := internal.NewCdfClient(config.ProjectName, config.CdfCluster, config.ClientID, clientSecret, config.Scopes, config.AdTenantId, config.AuthTokenUrl, config.CdfDatasetID)
	configObserver := internal.NewCdfConfigObserver(config.ExtractorID, cdfCLient, config.RemoteConfigSource, secretManager)
	if config.RemoteConfigSource == internal.ConfigSourceExtPipelines {
		configObserver.Start(config.ConfigReloadInterval * time.Second)
	}

	systemEventBus := pubsub.New[string, internal.SystemEvent](20)

	appManager = core.NewAppManager(configObserver)

	integrReg = make(map[string]Integration)

	for _, integrName := range config.EnabledIntegrations {
		switch integrName {
		case "ip_cams_to_cdf":
			intgr := ip_cams_to_cdf.NewCameraImagesToCdf(cdfCLient, config.ExtractorID, configObserver, systemEventBus)
			intgr.SetSecretManager(secretManager)
			if config.RemoteConfigSource == internal.ConfigSourceLocal {
				intgr.LoadConfigFromJson(config.Integrations["ip_cams_to_cdf"])
			}
			err = intgr.Start()
			if err != nil {
				log.Errorf(" %s integration can't be started . Error : %s", integrName, err.Error())
			} else {
				integrReg["ip_cams_to_cdf"] = intgr
				appManager.SetIntegration("ip_cams_to_cdf", intgr)
			}

		case "local_files_to_cdf":
			log.Info(" local_files_to_cdf integration not implemented yet")
			// intgr := local_files_to_cdf.NewLocalFilesToCdf(cdfCLient, config.ExtractorID, config.RemoteConfigSource)
			// // intgr.SetLocalConfig(config.LocalIntegrationConfig["local_files_to_cdf"].(integrations.LocalFilesToCdfConfig))
			// err = intgr.Start()
			// if err != nil {
			// 	log.Errorf(" %s integration can't be started . Error : %s", integrName, err.Error())
			// } else {
			// 	integrReg["local_files_to_cdf"] = intgr
			// }

		}
	}

	err = appManager.LoadAppsFromRawConfig(config.Apps)
	if err != nil {
		log.Error("Failed to load apps. Err:", err.Error())
		return
	}
}

func stopExtractor() {
	for _, intgr := range integrReg {
		intgr.Stop()
	}
}

func main() {

	log.Infof("----- Starting edge-extractor - version = %s ----------", Version)
	mainConfigPath := flag.String("config", "config.json", "Full path to main configuration file")

	base64encodedConfig := flag.String("bconfig", "", "Base64 encoded config")
	op := flag.String("op", "", "Supported operations : 'gen_config,install,uninstall,run' ")
	textToEncrypt := flag.String("secret", "", "Secret to encrypt")
	encryptionKey := flag.String("key", "", "Encryption key")
	flag.Parse()

	if *encryptionKey != "" {
		EncryptionKey = *encryptionKey
	} else if os.Getenv("EDGE_EXT_ENCRYPTION_KEY") != "" {
		EncryptionKey = os.Getenv("EDGE_EXT_ENCRYPTION_KEY")
	}

	if *mainConfigPath == "config.json" {
		*mainConfigPath = filepath.Join(internal.GetBinaryDir(), *mainConfigPath)
		fullConfigPath = *mainConfigPath
	} else {
		fullConfigPath = *mainConfigPath
	}

	var config internal.StaticConfig

	// User can configure app by passing configurations as one base64 encoded string
	if *base64encodedConfig != "" {
		log.Info("Loading configuration from cmd line parameter")
		// Base64 Standard Decoding
		body, err := base64.StdEncoding.DecodeString(*base64encodedConfig)
		if err != nil {
			log.Errorf("Error decoding base64 encoded config: %s ", err.Error())
			return
		}
		os.WriteFile("config.json", body, 0644)
	}

	internal.Key = EncryptionKey
	if EncryptionKey != "" {
		log.Info("Encryption key is set . Will try to decrypt config file")
	} else {
		log.Info("Encryption key is not set .")
	}

	switch *op {
	case "gen_config":
		log.Info("Generating config file")
		config.CdfCluster = "westeurope-1"
		config.Scopes = []string{"https://westeurope-1.cognitedata.com/.default"}
		config.AuthTokenUrl = "https://login.microsoftonline.com/set_your_tenant_id_here/oauth2/v2.0/token"
		config.EnabledIntegrations = []string{"ip_cams_to_cdf"}
		body, _ := json.MarshalIndent(&config, " ", "  ")
		os.WriteFile("config.json", body, 0644)
		return
	case "version":
		fmt.Println(Version)

	case "encrypt_config":
		if EncryptionKey == "" {
			fmt.Println("Please provide encryption key")
			return
		}

		err := encryptSecretsInConfig(*mainConfigPath)
		if err != nil {
			fmt.Println("Failed to encrypt config file. Err:", err.Error())
		}
		fmt.Println("Config file has been encrypted")
		return

	case "encrypt_secret":
		if EncryptionKey == "" {
			fmt.Println("Please provide encryption key")
			return
		}
		if *textToEncrypt == "" {
			fmt.Println("Please provide text to encrypt")
			return
		}
		encrypted, err := internal.EncryptString(internal.Key, *textToEncrypt)
		if err != nil {
			fmt.Println("Failed to encrypt string. Err:", err.Error())
			return
		}
		fmt.Println("Encrypted string : ", encrypted)

	case "install":
		log.Info("Installing edge-extractor service")
		runMode = "service"
		appService := configureService(true)
		err := appService.Install()
		if err != nil {
			log.Error("Failed to install service.Make sure you run installation as system administrator Err: ", err.Error())
		} else {
			err = appService.Start()
			if err != nil {
				log.Error("Failed to run service. Err: ", err.Error())
			}
			log.Info("Service has been installed and started")
		}
	case "uninstall":
		log.Info("Uninstalling edge-extractor service")
		appService := configureService(false)
		err := appService.Uninstall()

		if err != nil {
			log.Error("Failed to uninstall service", err.Error())
		}
		if runtime.GOOS == "linux" {
			internal.RemoveLinuxServiceEnv()
		}
	case "update":
		log.Info("Updating edge-extractor service binary")
		internal.UpdateLinuxServiceBinary()
	case "event-bus":
		log.Info("Starting event bus")
		eventBus := pubsub.New[string, string](20)
		topic := []string{"1/tnsaxis:CameraApplicationPlatform/FenceGuard/Camera1Profile1"}
		stream := eventBus.Sub(topic...)
		go func() {
			for msg := range stream {
				log.Info("Received message : ", msg)
				break
			}
		}()
		log.Info("Publishing message")
		eventBus.TryPub("test", "1/tnsaxis:CameraApplicationPlatform/FenceGuard/Camera1Profile1")
		select {}

	case "run":
		// Should be used to start service from CLI\
		runMode = "cli"
		startEdgeExtractor(*mainConfigPath)
		select {}
	default:
		// Used by OS service supervisor
		runMode = "service"
		appService := configureService(false)
		err := appService.Run()
		if err != nil {
			log.Error(err)
		}
	}

}
