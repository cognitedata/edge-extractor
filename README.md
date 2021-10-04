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

### Configurations

The service is using 2 types of configurations : 
1. Static - loaded durign service startup . 
2. Dynamic - loaded from remote endpoint , for instance from Asset metadata. 

### Development 

TBD