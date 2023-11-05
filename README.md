# Edge extractor 

The edge extractor is a remote data extraction agent. The service should be installed on-premise (typically DMZ network) and talk to devices over the local network or via serial protocols.

The service is distributed as a self-contained platform-specific binary.  

### Architecture 

Each extraction pipeline is built around the concept of the integration process, with `input connectors` for data extraction from an input device and `output connects` for data ingestion into the output system.

High level overview :

![Edge extractor high level diagram 1](/docs/edge-extractor-high-level.png)

Integration process

![Edge extractor high level diagram 2](/docs/edge-extractor-integr-process.png)


### Connectors 

#### Input connector 

- Ip camera input connector 

#### Output connector 

- CDF output 

### Integration processes 

- Camera to CDF . The process loads configurations from CDF asset -> requests images from each IP camera -> uploads images to CDF Files and link them to camera asset
- Local storage file to CDF. The process loads images from SD a card or different local storage, uploads images to CDF Files, and links them to camera assets.

All processes support parallel data retrieval from multiple devices. 

### Device drivers 

Supported camera drives : 

- Axis 
- Hikvision 
- Reolink 
- File system. 
- Generic URL camera 
- Fliw Ax8
- Dahua

### Implemented integrations 

- ip_cams_to_cdf 
- local_files_to_cdf

### Configurations

The service is using 2 types of configurations : 
1. Static - loaded durign service startup . 
2. Dynamic - loaded from remote endpoint , for instance from Asset metadata. 

### Service CLI parameter 

`--op` - operation , supported operations : 
   - `run` - rund the service in command line 
   - `install` - installs the service as windows , osx or linux service
   - `uninstall` - uninstalls the service 
   - `gen_config` - generates default config
   - `encrypt_config` - encrypts all Secret and password field in config file
   - `encrypt_secret` - encrypts secret provided as `secret` CLI parameter and outputs encrypted value to stdout

`--config <path_to_config_file>` - must be used to change default location of config file 
`--bconfig <base64_encoded_string>` - base64 encoded config that can be passed to the application during startup 

Examples : 

`./edge-extractor --run --bconfig ewogICAgIlByb2plY3ROYW1lIjogIm1haW5zdHJlYW0tZGV2IiwKIC` 

`./edge-extractor --install --bconfig ewogICAgIlByb2plY3ROYW1lIjogIm1haW5zdHJlYW0tZGV2IiwKIC`

`./edge-extractor --op run --config config_test.json`

`./edge-extractor --op install --config config_test.json`

`./edge-extractor --op uninstall`

`./edge-extractor --op gen_config`

`./edge-extractor --op encrypt_config`

`./edge-extractor --op encrypt_secret --secret my_secret`

### Registering application as Windows service 

1. Create folder `C:\Cognite\EdgeExtractor`
2. Upload `edge-extractor` to the folder 
3. Open Command Prompt and run command `cd C:\Cognite\EdgeExtractor` 
4. Register `edge-extractor` as Windows service by running the command - `edge-extractor-win-amd64.exe --op install` or `edge-extractor-win-amd64.exe --op install --bconfig ewogICAgIlByb2plY3ROYW1lIjogIm1haW5zdHJlYW0tZGV2IiwKIC`


### Config file 

Minimal local config , all cameras configured remotely.

````
{
    "ExtractorID":"edge-extractor-dev-1",
    "ProjectName": "_set_your_project_name_here",
    "CdfCluster": "westeurope-1",
    "AdTenantId": "176a22cf-3d72-4b07-a4f8-0841557a570c",
    "AuthTokenUrl": "https://login.microsoftonline.com/176a22cf-3d72-4b07-a4f8-0841557a570c/oauth2/v2.0/token",
    "ClientID": "_set_your_client_id_here_",
    "Secret": "_set_your_client_secret_here_",
    "Scopes": [
      "https://westeurope-1.cognitedata.com/.default"
    ],
    "CdfDatasetID": 866030833773755,
    "EnabledIntegrations": [
      "ip_cams_to_cdf"
    ],
    "LogLevel":"debug",
    "LogDir":"-",  
  }
````
Local config , all configurations loaded from local config files.

````
{
    "ExtractorID":"edge-extractor-dev-1",
    "ProjectName": "_set_your_project_name_here",
    "CdfCluster": "westeurope-1",
    "AdTenantId": "176a22cf-3d72-4b07-a4f8-0841557a570c",
    "AuthTokenUrl": "https://login.microsoftonline.com/176a22cf-3d72-4b07-a4f8-0841557a570c/oauth2/v2.0/token",
    "ClientID": "_set_your_client_id_here_",
    "Secret": "_set_your_client_secret_here_",
    "Scopes": [
      "https://westeurope-1.cognitedata.com/.default"
    ],
    "CdfDatasetID": 866030833773755,
    "EnabledIntegrations": [
      "local_files_to_cdf"
    ],
    "LogLevel":"debug",
    "LogDir":"-",
    "IsEncrypted": false,
    "LocalIntegrationConfig": [{
      "id":403447394704254,
      "name":"local_uploader",
      "metadata": {
        "cog_model":"fscam",
        "uri":"./imgdump",
        "polling_interval":"0",
        "max_parallel_runs":"3"
      }
      
    }]
  }
````

By default Secret and all password fields (in camera config section) are stored in plain text , to make it more secure the extractor also supports encrypted mode.
To encrypt the config file run `edge-extractor --op encrypt_config` command , this will encrypt Secret and all password fields in config file and encrypted version
will be saved to the same configuration file. Create a copy of unencrypted config file before running the command. 


### Extractor monitoring 

The extractor can be monitored remotely via CDF extraction pipelines. Exraction pipelines must be created in CDF upfront (via CDF Fusion or using SDK) and 
extractor pipeline ExternalID must match `ExtractorID` in config above.

More information about CDF extraction pipelines can be found [here](https://docs.cognite.com/cdf/integration/guides/interfaces/monitor_integrations/) 


TBD