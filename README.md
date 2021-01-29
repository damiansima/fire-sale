# FIRE SALE
[![CircleCI](https://circleci.com/gh/damiansima/fire-sale.svg?style=shield&circle-token=e2f5b5b9357eb8cc430df5619e92925502ea606f)](<https://app.circleci.com/pipelines/github/damiansima/fire-sale>)
[![license](http://img.shields.io/badge/license-Apache%20v2-orange.svg)](https://raw.githubusercontent.com/master/LICENSE)

``Everything must GO(lang)``

FireSale is a performance testing tool designed to be lightweight and fast. 
Its simple DSL allows you to generate load and spike like traffic in order 
to stress your services in a way that reflects your production traffic. 

## REQUIREMENTS
It requires GO 1.15

## BUILDING
```
$ go build
$ go test -cover ./...
```

## USAGE
```
$ ./fire-sale -config ./path/to/my-config.yml
```

For usage options just type:
```
$ ./fire-sale
  -config string
    	Path to the test-configuration.yml
  -log string
    	Define the log level [panic|fatal|error|warn|info|debug|trace] (default "info")
  -report-path string
    	Define the report file path. If not provided it'll be printed to stdout
  -report-type string
    	Define the report type [std|json] (default "std")
```

# DSL 
The goals behind FireSale is to make it super simple to use. 
With that in mind we've come up with the following DSL which could be fed to the engine in order to run your tests:
The DSL describes stress tests in which you define a traffic profile you want to reproduce. 
It's composed by the following sections: 
* Basic
* Parameters
* Certificates
* Scenarios

*Note*: It supports YAML & JSON files.

```yaml
name: da stress test
host: https://www.fake-host.com
parameters:
  noofrequest: 10
  noofwarmuprequest: 2
  testduration: 0
  warmupduration: 0
  workers: 1
  maxrequest: 0
  rampup:
    step: 1
    time: 0
certificates:
  clientkeyfile:  /path/to/your-key-file.key
  clientcertfile: /path/to/your-cert-file.crt
  cacertfile:     /path/to/your-ca-file.crt
scenarios:
  - name: First endpoint
    distribution: 0.7
    timeout: -1
    method: GET
    path: /
    headers:
      user-agent: fire-sale/0.0.1
  - name: Another endpoint
      distribution: 0.3
      timeout: -1
      method: GET
      path: /another-endpoint
      headers:
        user-agent: fire-sale/0.0.1  
```

## DSL :: Basic
This section is just the root, it describes the:
 - name 
 - host

```yaml
name: da stress test
host: https://www.fake-host.com
``` 
The host will be used as the base of all the HTTP requests

## DSL :: Parameters
This section describes the amount of traffic you want to generate and how to get there:
```yaml
  parameters:
    noofrequest: 10
    noofwarmuprequest: 2
    testduration: 0
    warmupduration: 0
    workers: 1
    maxrequest: 0
    rampup:
      step: 1
      time: 0
```
- **noofrequest**: It's used when you just want to generate an specific number of hits. The run will finish after executing all the requests defined.
- **noofwarmuprequest**: Indicate the number of request done in order to warmup the service. These request will be accounted differently. 
- **testduration**: It instructs the tool to run for a period of time. If present it takes precedence over `noofrequest`. Valid units are ["ns", "us", "ms", "s", "m", "h"] as described [here](https://golang.org/pkg/time/#ParseDuration). If not provided the default unit is `minutes (m)` 
- **warmupduration**: It instructs the tool signal the request as warm up for a period of time. If present the request done during that period of time will be for warm up purposes. These request will be accounted differently. Valid units are ["ns", "us", "ms", "s", "m", "h"] as described [here](https://golang.org/pkg/time/#ParseDuration). If not provided the default unit is `minutes (m)`
- **workers**: It defines the number of concurrent users you want to simulate.
- **maxrequest**: It defines an overall max to the number of request generated per second.  If `0` there is no limitation, if `10` it'll only generate 10 request per second regardless of the number of workers.

### DSL :: Parameters :: Ramp Up
This section defines how the maximum number of workers is reached. 
```yaml
  parameters:
    rampup:
      step: 1
      time: 0
```
If not defined it will spin up all workers at the same time generating a traffic profile that will look like a spike.
Normally you want to smooth the curve of traffic to simulate actual traffic.
- **time**: It defines the time you want to wait between the beginning of the execution up until reaching the maximum number of workers running. Valid units are ["ns", "us", "ms", "s", "m", "h"] as described [here](https://golang.org/pkg/time/#ParseDuration). If not provided the default unit is `minutes (m)`
- **step**: It takes the defined time and splits it by this number generating and scalonated traffic. The larger this number the smoother the traffic curve. 

*Note*: if not provided the default ram up is `step:1 , time:0` 
 
## DSL :: Certificates
FireSale supports the usage of key/cert files for when you need to hit services behind TSL. 
If present the below section will load the certs and use it for all scenarios. As a general part of the file it will use them for all `scenarios`  
```yaml
certificates:
  clientkeyfile:  /path/to/your-key-file.key
  clientcertfile: /path/to/your-cert-file.crt
  cacertfile:     /path/to/your-ca-file.crt
```

*Note:* if you do not have a pem file for the CA file just point to you *.crt
*Note:* for ease of use, by default it support self signed certificates by skipping insecure verifications

## DSL :: Scenarios
The `scenarios` is where you describe the actual HTTP requests to be made. 
The order of execution of the scenarios is random, respecting the distribution profile.
```yaml
scenarios:
  - name: First endpoint
    distribution: 0.7
    timeout: -1
    method: GET
    path: /
    headers:
      user-agent: fire-sale/0.0.1
  - name: Another endpoint
      distribution: 0.3
      timeout: -1
      method: PUT
      path: /another-endpoint
      headers:
        user-agent: fire-sale/0.0.1  
      body: {id:123, value:'some value'}  
```

As you can see it's an array and you can define as many as required.

- **name**: It defines the name of the test used for reporting purposes.
- **timeout**: It defines the timeout for this specifc type of request. If `-1` it will not timeout. If it times out it'll be reported.
- **method**: It describles the HTTP method to be used it supports the same valid methods as http.request Go package.
- **path**: It describes the path to hit.
- **headers**: It's a key-value map with the headers to be sent.
- **body**: The body to be sent.

## DSL :: Scenarios :: Distribution          
This part deserves its own section.
```yaml
- distribution: 0.3
``` 
The distribution allows you to assign a percentage value from `0` to `1`.
The engine will randomly select a scenario to be run each time. If you only have one scenario with a distribution of 1 it will then only run that scenario. 
But if you have two scenarios you can ask the engine to distribute the executions evenly, that's 50/50, or you can ask it to do 70/30.

This feature allows you to, in one execution, replicate complex traffic profiles as usually your services expose a number of endpoints but not all of them are hit in the same proportion. With this you can stress your service and more realistic conditions.  

*Note*: The sum of the distributions of all scenarios must add up to `1` or  the execution will fail.

### DSL :: Value Generators
There are a number of functions you can use to generate values randomly. 
The goal of it is to define just one scenario and allow the engine to select in each request a random value in order to replicate a more realistic use case. 
This functions can be used in the following sections of the `DSL`: 
* host
* path
* body 

A function in the DSL is placeholded by two curly braces like so `{{` & `}}`: 
```yaml
path: /?id={{RandInRange(0,11)}}
```

#### Functions
- RandInRange
- RandInList
- RandInFile


**RandInRange** 

`RandInRange(0,11)`

It returns and string representing and integer number between the [min, max) values

**RandInList** 

`RandInList(a,b,1,2,3)`

It returns and string by selecting a random item from the list sent as parameter.

**RandInFile** 

`RandInFile(./my-exampl-file.dat)`

It returns and string by selecting a random item from the file sent as parameter. Each line in the file is an item.

# REPORTS
**Note**: It currently only prints reports in the console
```
******************************************************** 
*                      Results                         * 
******************************************************** 
======================================================== 
=                     Scenarios                        = 
======================================================== 
Scenario - Get Retailers Products 50 to 100 - ID: [3] 
Success [100.000000%] - Fail [0.000000%]     
Request average [282.136205ms]               
Request total [2103] average [1.359391885s]  
99th 1035.738501ms                           
90th 298.641789ms                            
75th 272.957098ms                            
50th 263.477751ms                            
-------------------------------------------------------- 
Scenario - Get Retailers 0 to 30 - ID: [0]   
Success [100.000000%] - Fail [0.000000%]     
Request average [27.765715ms]                
Request total [2368] average [1.088025691s]  
99th 380.675666ms                            
90th 44.338142ms                             
75th 20.708892ms                             
50th 11.321021ms                             
-------------------------------------------------------- 
Scenario - Get Retailers 50 to 100 - ID: [1] 
Success [100.000000%] - Fail [0.000000%]     
Request average [29.725625ms]                
Request total [2265] average [1.086346552s]  
99th 417.755235ms                            
90th 47.587332ms                             
75th 20.801271ms                             
50th 11.725982ms                             
-------------------------------------------------------- 
Scenario - Get Retailers Products 0 to 50 - ID: [2] 
Success [100.000000%] - Fail [0.000000%]     
Request average [30.708097ms]                
Request total [2187] average [1.094739772s]  
99th 428.276235ms                            
90th 50.205013ms                             
75th 19.797776ms                             
50th 11.357867ms                             
-------------------------------------------------------- 
======================================================== 
=                     Overall                          = 
======================================================== 
Success [100.000000%] - Fail [0.000000%]     
Request average [88.935201ms]                
Request total [8923] average [1.153201479s]  
99th 494.090542ms                            
90th 266.930053ms                            
75th 215.387629ms                            
50th 15.555311ms                             
Timeout [0] - Fail [0] - Success [8923]      
Execution took [3m1.856963175s]
[¡¡¡SOLD!!!]                     
```
The report will print statistics per each scenario, and an overall result. 
In each section it will show the following: 
```
Scenario - Get Retailers Products 50 to 100 - ID: [3] 
Success [100.000000%] - Fail [0.000000%]     
Request average [282.136205ms]               
Request total [2103] average [1.359391885s]  
99th 1035.738501ms                           
90th 298.641789ms                            
75th 272.957098ms                            
50th 263.477751ms                            
```
- Name: taken from the DSL
- Success & Fail: a percentage of all the request that where successful, and those which didn't. Any response whose status code is higher than 300 is considered a fail.
- Request average: average time per request (not quite informative TBH)
- Request total: total number of request done for this scenario
- Request time percentiles: ..... 


## DSL :: JSON Example 
```json
{
  "Name": "da stress test",
  "Host": "https://www.fake-host.com",
  "Parameters": {
    "NoOfRequest": 10,
    "TestDuration": 0,
    "Workers": 1,
    "MaxRequest": 0,
    "RampUp": {
      "Step": 1,
      "Time": 0
    }
  },
  "Certificates": {
    "ClientCertFile": "/path/to/your-cert-file.crt",
    "ClientKeyFile": "/path/to/your-key-file.key",
    "CaCertFile": "/path/to/your-ca-file.crt"
  },
  "Scenarios": [
    {
      "Name": "First endpoint",
      "Distribution": 0.7,
      "Timeout": -1,
      "Method": "GET",
      "Path": "/",
      "Headers": {
        "user-agent": "fire-sale/0.0.1"
      },
    },
    {
      "Name": "Another endpoint",
      "Distribution": 0.3,
      "Timeout": -1,
      "Method": "GET",
      "Path": "/another-endpoint",
      "Headers": {
        "user-agent": "fire-sale/0.0.1"
      },
    }
  ]
} 

```
## TODO LIST 
### DONE
| **Feature** | **Status** | *Notes* |
| --------|-------|------- |
|Engine  | [DONE] ||
|Complex request  | [DONE] ||
|Time out | [DONE] | Threshold over which we drop the test|
|Error Count | [DONE] | Count any state different to 2xx|
|Count Results by time | [DONE] | increase buffer of jobs and results and make the results be consumed by unit of time each of them will be a data point|
|Use http tracing to measure parts of the request and their times | [DONE] | https://medium.com/@ankur_anand/an-in-depth-introduction-to-99-percentile-for-programmers-22e83a00caf|
|Report 99, 95, median | [DONE] | https://medium.com/@ankur_anand/an-in-depth-introduction-to-99-percentile-for-programmers-22e83a00caf|
|Measure network turn around and try to substract that from measurements  | [DONE] | |
|Play with connection poll| [DONE] |No reall need so far to tune default poll http://tleyden.github.io/blog/2016/11/21/tuning-the-go-http-client-library-for-load-testing/ |
|Scalonated Traffic Steps vs Ramup time  | [DONE] ||
|Cap the max of request | [DONE] ||
|Force new connection (not reuse) per go routine  | [DONE] | to simulate different users|
|Add option to job to reuse http connection between workers| [DONE] ||
|Multiple runs per ACTUAL RUN| [DONE] ||
|Define timeout per request| [DONE] ||
|Multiple scenario | [DONE] | report by job of request and percentiles|
|Add build to github  | [DONE] ||
|DSL|[DONE]||
|Value generators|[DONE]| to randomize request for those parameters that can allow for auto generation between min and max|
|Certificates handling  | [DONE] | Support for key/cert and PEM files |
|Command line |[DONE]| |
|Reporting |[DONE]| Different type of output reports|
|Reporting - As a module|[DONE]| |
|Warm up request|[DONE]||
| FEAT |[DONE]|   TIMEOUT : allow for units in time & define default unit|
### PENDING

| **Feature** | **Status** | *Notes* |
| --------|-------|------- |
| BUG   |[PENDING]| IF ADDING A SPACE AFTER THE COMMA BREAKS body: "{{RandInRange(0, 1000)}}”|
| BUG   |[PENDING]| THIS BRAKES WITH panic: runtime error: index out of range [-1] check "moliva-iob-bug.yml"|
| FEAT|[PENDING]| default distribution if no distribution distribute uniformly and if pressent in some just complete until 1|
| DOC|[PENDING]| explain this log lines properly "Request per 10 second [403] | per 1 second [40]..."|
|Reporting - Warm up request|[PENDING]||
|Tests dude  | [PENDING] ||
|Deal with TODOs  | [PENDING] ||
|Reporting - Reporting Suggestion|[PENDING]| In the event of several DNS resolution suggest change url|
|Auto Generation of DSL  |[PENDING]| Swagger Scaffolding|
|Events  |[PENDING]| Generate load by creating events|
|Payload size analisis   | [PENDING] ||
|Fire Sale as module  | [PENDING] | Make a module out of this and move actual job testing outside|
|Speed up to a desired number of request vs concurrent users  | [PENDING] ||
|DSL short for Capacity testing | [PENDING] |Keep adding workers up until timeout is constant|
|DSL short for Spike  Traffic | [PENDING] ||
|Check Job Buffer size to size it based on the expected max RPS| [PENDING] |we size the job buffer to at least the number of jobs so not to choke the producer|
|Numbers with network latency should be a configuration| [PENDING] | |
|BUG in the combination of number of jobs workers and running time  | [PENDING] | |
|Support for gPRC|[PENDING]||
|Reduce binary size|[PENDING]| it currently wights 9+ mb [check](https://golang.org/doc/install/gccgo), this `go build -ldflags "-w" ` to it to 7.3  |

## PLATFORM TODO
| **Feature** | **Status** | *Notes* |
| --------|-------|------- |
|Soak Testing  | [PENDING] ||
