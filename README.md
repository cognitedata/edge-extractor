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

### Device drivers 

Supported camera drivers : 

- Axis - `axis`  
- Hikvision - `hikvision`
- Reolink - `reolink`
- File system - `fscam`
- Generic URL camera - `urlcam`
- Flir Ax8 - `flir_ax8`
- Dahua - `dahua`


### Configurations

The service is using 2 types of configurations : 
1. Static - loaded durign service startup . 
2. Dynamic - loaded from remote endpoint (CDF) during service startup and periodically updated. Supported remote sources : CDF Extraction pipelines configs.

#### Static configuration

Static configuration is loaded from local config file (JSON) and contains information about CDF project , CDF cluster , CDF dataset , CDF authentication , etc.

Parameter | ENV_VAR | Description | Example
--- | --- | --- | ---
`ExtractorID` | EDGE_EXT_EXTRACTOR_ID | Unique ID of the extractor | `edge-extractor-dev-1`
`ProjectName` | EDGE_EXT_CDF_PROJECT_NAME | Name of the CDF project | `my-project`
`CdfCluster` | EDGE_EXT_CDF_CLUSTER | Name of the CDF cluster | `westeurope-1`
`AdTenantId` | EDGE_EXT_AD_TENANT_ID | Azure AD tenant ID | `example-tenant-4b07-a4f8-0841557a570c`
`AuthTokenUrl` | EDGE_EXT_AD_AUTH_TOKEN_URL | Azure AD token endpoint URL | `https://login.microsoftonline.com/example-tenant-4b07-a4f8-0841557a570c/oauth2/v2.0/token`
`ClientID` | EDGE_EXT_AD_CLIENT_ID | Azure AD client ID | `example-3d72-4b07-a4f8-0841557a570c`
`Secret` | EDGE_EXT_AD_SECRET | Azure AD client secret | `example-secret`
`Scopes` | EDGE_EX_AD_SCOPES | Azure AD scopes | `https://westeurope-1.cognitedata.com/.default`
`CdfDatasetID` | EDGE_EXT_CDF_DATASET_ID | CDF dataset ID | `866030833773755`
`EnabledIntegrations` | EDGE_EXT_ENABLED_INTEGRATIONS | List of enabled integrations | `ip_cams_to_cdf`
`LogLevel` | EDGE_EXT_LOG_LEVEL | Log level | `debug`
`IsEncrypted` | EDGE_EXT_IS_ENCRYPTED | Is config encrypted (true/false) | `false`
`Secrets` |  | Map of secrets | `{"cdf_client_secret":"_encrypted_secret_"}`
`Integrations` |  | Collection of integration specific configurations | `{"ip_cams_to_cdf":{...}}`


### Integrations 

#### ip_cams_to_cdf 
The process connects to each IP camera , fetches images using camera specific API and uploads images to CDF Files and optionally links them to camera asset.
The processes supports parallel data retrieval from multiple devices. 

Configurations : 

Parameter | Description | Example
--- | --- | ---
`Name` | Name of the camera | `camera-1`
`Id` | ID of Asset that repsents camera . All images are linked to that Asset if configured | 403447394704254
`Model` | Camera model (from the list of supported camera drivers) | `axis`
`Address` | Camera endpoint URI | `http://10.22.15.62` , `rtsp://` , `./imgdump`
`PollingInterval` | Polling interval in seconds | 10
`Username` | Username | `admin`
`Password` | Password. It can be either plain text value of key that must exist in Secrets section of config or ENV variable. | `admin`
`State` | State of the camera (enabled/disabled) | `enabled`



### Service CLI parameters

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

Minimal local config for remotely configured cameras.

Content of `config.json` file :

````
{
   "ProjectName": "_cdf_project_name_",
   "CdfCluster": "_cdf_cluster_name_",
   "AdTenantId": "_azure_ad_tenant_id_",
   "AuthTokenUrl": "https://login.microsoftonline.com/_azure_ad_tenant_id_/oauth2/v2.0/token",
   "ClientID": "_service_principal_client_id_",
   "Secret": "cdf_client_secret",
   "Scopes": [
     "https://az-power-no-northeurope.cognitedata.com/.default"
   ],
   "CdfDatasetID": 2626756768281823,
   "ExtractorID": "edge-extractor-1",
   "RemoteConfigSource": "ext_pipeline_config",
   "ConfigReloadInterval": 0,
   "EnabledIntegrations": [
     "ip_cams_to_cdf"
   ],
   "LogLevel": "debug",
   "LogDir": "-",
   "IsEncrypted": true,
   "Secrets": {
     "cdf_client_secret": "_encrypted_secret_",
   }
 }
````
Content of remote configuration (CDF extraction pipeline config) :

```
{
"Integrations":{
     "ip_cams_to_cdf": {
       "Cameras": [
         {
           "ID": 1,
           "Name": "Camera 1",
           "Model": "fscam",
           "Address": "./test/data/imgdump",
           "Username": "root",
           "Password": "camera_1_password",
           "Mode": "camera",
           "PollingInterval": 15,
           "State": "enabled"
         },
         {
           "ID": 2,
           "Name": "URL virtual camera",
           "Model": "urlcam",
           "Address": "https://media.istockphoto.com/id/1496920323/photo/thinking-robot.jpg?s=612x612&w=0&k=20&c=heLdpKSUvt0w7dtm5IXmu4JJ5GvPvdpcywrqcL78P9s=",
           "Username": "",
           "Password": "",
           "Mode": "camera",
           "PollingInterval": 15,
           "State": "disabled"
         }
       ]
     }
},
"Secrets":{}
}

```

Local config with locally configured cameras.

````
{
   "ProjectName": "_cdf_project_name_",
   "CdfCluster": "_cdf_cluster_name_",
   "AdTenantId": "_azure_ad_tenant_id_",
   "AuthTokenUrl": "https://login.microsoftonline.com/_azure_ad_tenant_id_/oauth2/v2.0/token",
   "ClientID": "_service_principal_client_id_",
   "Secret": "cdf_client_secret",
   "Scopes": [
     "https://az-power-no-northeurope.cognitedata.com/.default"
   ],
   "CdfDatasetID": 2626756768281823,
   "ExtractorID": "edge-extractor-1",
   "RemoteConfigSource": "local",
   "ConfigReloadInterval": 0,
   "EnabledIntegrations": [
     "ip_cams_to_cdf"
   ],
   "LogLevel": "debug",
   "LogDir": "-",
   "IsEncrypted": true,
   "Integrations":{
    "ip_cams_to_cdf": {
      "Cameras": [
        {
          "ID": 1,
          "Name": "Camera 1",
          "Model": "fscam",
          "Address": "./test/data/imgdump",
          "Username": "root",
          "Password": "camera_1_password",
          "Mode": "camera",
          "PollingInterval": 15,
          "State": "enabled"
        },
        {
          "ID": 2,
          "Name": "URL virtual camera",
          "Model": "urlcam",
          "Address": "https://media.istockphoto.com/id/1496920323/photo/thinking-robot.jpg?s=612x612&w=0&k=20&c=heLdpKSUvt0w7dtm5IXmu4JJ5GvPvdpcywrqcL78P9s=",
          "Username": "",
          "Password": "",
          "Mode": "camera",
          "PollingInterval": 15,
          "State": "disabled"
        }
      ]
    }
 },
   "Secrets": {
     "cdf_client_secret": "_encrypted_secret_",
   }
 }
````

### Secret management

The service support 3 ways of storing secrets :
- In plain text in config file
- In encrypted form in config file (In this case `IsEncrypted` parameter must be set to `true` and `Secrets` section must be present in config file)
- In environment variables

Encryption and decryption is done using AES-256 algorithm with 32 bytes long key. The key is set during build time but can be changed by setting `EDGE_EXT_SECRET_KEY` environment variable.

The service also can fetch secrets from environment variables. The name of the environment variable must match the name of the secret in config file. Example : a secret can be set as environment variable `CDF_CLIENT_SECRET` and the service will fetch the value from environment variable , in config file it should be referenced  `"Secret": "cdf_client_secret"` or `"Password": "CDF_CLIENT_SECRET"`


The service provides convenient way to encrypt all secrets in config file using CLI command `edge-extractor --op encrypt_config`. The command will generate new config file with encrypted secrets and save it to `config_encrypted.json` file.

Another command can be used to encrypt one secret `edge-extractor --op encrypt_secret --secret <secret_value>` and output encrypted value to stdout.

### Extractor monitoring and remote configuration

The extractor can be monitored remotely via CDF extraction pipelines. Exraction pipelines must be created in CDF upfront (via CDF Fusion or using SDK) and 
extractor pipeline ExternalID must match `ExtractorID` in config above.

Remote configuration using CDF Fusion UI :

![Remote config](/docs/remote-config.png)

Remote monitoring using CDF Fusion UI :

![Remote monitoring](/docs/remote-monitoring.png)

More information about CDF extraction pipelines can be found [here](https://docs.cognite.com/cdf/integration/guides/interfaces/monitor_integrations/) 


TBD