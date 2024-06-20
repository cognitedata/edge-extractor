# Axis Cameras: Specifics and Settings

This page provides detailed information about Axis cameras, including their specific features and settings.


### Supported events 

Camera advertisings supported event via ONVIF protocol. 

Discovery URL : http://<camera_ip>/vapix/services

POST request to the URL with the following body will return the supported events.

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">
 <Header/>
 <Body >
  <GetEventInstances xmlns="http://www.axis.com/vapix/ws/event1"/>
 </Body>
</Envelope>
```


### Event Subscription

Examples : 

Subscribe to all events emmited by FenceGuard application on Camera1Profile1


Topic : tnsaxis:CameraApplicationPlatform/ObjectAnalytics/Device1Scenario2

Filtes : boolean(//SimpleItem[@Name="active" and @Value="1"])


Subscribe to all events emmited by FenceGuard application on Camera1Profile1:


### Events 

```json
 {"apiVersion":"1.0","method":"events:notify","params":{"notification":{"topic":"tnsaxis:CameraApplicationPlatform/FenceGuard/Camera1Profile1","timestamp":1710752405172,"message":{"source":{},"key":{},"data":{"active":"1"}}}}}
 {"apiVersion":"1.0","method":"events:notify","params":{"notification":{"topic":"tnsaxis:CameraApplicationPlatform/FenceGuard/Camera1Profile1","timestamp":1710752408371,"message":{"source":{},"key":{},"data":{"active":"0"}}}}}
 {"apiVersion":"1.0","method":"events:notify","params":{"notification":{"topic":"tnsaxis:CameraApplicationPlatform/FenceGuard/Camera1Profile1","timestamp":1710752405172,"message":{"source":{},"key":{},"data":{"active":"1"}}}}}"
 {"apiVersion":"1.0","method":"events:notify","params":{"notification":{"topic":"tnsaxis:CameraApplicationPlatform/FenceGuard/Camera1Profile1","timestamp":1710752408371,"message":{"source":{},"key":{},"data":{"active":"0"}}}}}

```