# Edge extractor 

Edge extractor is remote data extraction agent. The service should be installed in on premise (normally DMZ network) and connect to an equipment over local network or via serial protocols. 

The service is disctributed as self contained platform specific binary. 

### Architecture 

Each extraction pipeline is build around concept of integration process with `input connectors` for data extraction from source system and `output connects` for data ingestion into sync. 

High level overview :

![Edge extractor high level diagram 1](/docs/edge-extractor-high-level.png)

Integration process

![Edge extractor high level diagram 2](/docs/edge-extractor-integr-process.png)


### Connectors 

#### Input connector 

- Ip camera input connector 

#### Output connector 

- CDF output 

### Inetration processes 

- Camera to CDF . The process loads configuraions from CDF asset -> requests images from each IP camera -> uploads images to CDF Files and link them to camera asset
- Local storage file to CDF . The process load images from SD cars or another local storage -> uploads images to CDF Files and link them to camera asset.

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