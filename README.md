# FIRE SALE
[![CircleCI](https://circleci.com/gh/damiansima/fire-sale.svg?style=shield&circle-token=e2f5b5b9357eb8cc430df5619e92925502ea606f)](<https://app.circleci.com/pipelines/github/damiansima/fire-sale>)

``Everything must GO(lang)``

FireSale is a performance testing tool designed to be ligtwayth and fast. 
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
```

# DSL 
The goals behind FireSale is to make it supper simple to use. 
With that in mind we've come up with the following DSL which could be feed to the engine in order to run your tests:
The DSL describe a stress tests in which you define a traffic profile you want to reproduce. 
It's composed  by the following sections: 
* Basic
* Parameters
* Certificates
* Scenarios

```
name: da stress test
host: https://www.fake-host.com
parameters:
  noofrequest: 10
  testduration: 0
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
This section it's just the root it describes the:
 - name 
 - host

```
name: da stress test
host: https://www.fake-host.com
``` 
The host will be used as the base to all the HTTP requests

## DSL :: Parameters
This section describes the amount of traffic you want to generate and how to get there
```
  parameters:
    noofrequest: 10
    testduration: 0
    workers: 1
    maxrequest: 0
    rampup:
      step: 1
      time: 0
```
- **noofrequest**: It used when you just want to generate an specific number of hits. The run will finish after executing all the request defined.
- **testduration**: It the instructs the tool to run for a period of time measured in  `minutes`. If present it takes precedence over `noofrequest`.
- **workers**: It defines the number of concurrent users you want to simulate.
- **maxrequest**: It the defines an overall max to the number of request generated per second.  If `0` there is no limitation, if 10 it'll only generate 10 request per second regardless the number of workers.

### DSL :: Parameters :: Ramp Up
This section defines how to maximum number of workers gets fire up. 
```
  parameters:
    rampup:
      step: 1
      time: 0
```
If not defined it will spin up all workers at the same time generating a traffic profile that will look like an spike.
Normally you want to smooth the curve of traffic to simulate actual traffic.
- **time**: It the defines the amount of minutes you want to leave between the beginin of the run up until reach the maximum number of workers running
- **step**: It takes the defined time and split it by this number generating and scalonated traffic. The larger this number the smoother the traffic curve. 

*Note*: if not provided the default ram up is step:1 , time:0 
 
## DSL :: Certificates
FireSale supports the usage of key/cert files for when you need to hit services behind TSL. 
If present the bellow section will load the certs and use it for all scenarios. As a general part of the file it will use them for all `scenarios`  
```
certificates:
  clientkeyfile:  /path/to/your-key-file.key
  clientcertfile: /path/to/your-cert-file.crt
  cacertfile:     /path/to/your-ca-file.crt
```

*Note:* if you do not have a pem file for the CA file just point to you *.crt
*Note:* for easy of use, by default it support self signed certificates by skiping insecure verifications

## DSL :: Scenarios
The `scenarios` is where you describe the actual HTTP requests to be made. 
The order of execution of the scenarios is random, respecting the distribution profile.
```
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

As you can see it's an array and you can define as many as require.

- **name**: It defines the name of the test used for reporting purposes.
- **timeout**: It defines the time out for this especifc type of request. If -1 it will not timeout. If it timeout it'll be reported.
- **method**: It describles the HTTP method to be used it supports the same valid methods as http.request Go package.
- **path**: It describes the path to hit.
- **headers**: It's a key value map with the headers to be sent.
- **body**: The body to be sent.

## DSL :: Scenarios :: Distribution          
This part deserves its own section.
```
- distribution: 0.3
``` 
The distribution allows you to assign a percentage value from `0` to `1`.
The engine will randomly select a scenario to be run each time. If you only have one scenario with a distribution of 1 it will then only run that scenario. 
But if you have two scenarios you can ask the engine to distribute the runs evenly, that's 50/50, or you can ask it to do 70/30.

This feature allows you to, in one execution, replicate complex traffic profiles as ussully your services exposed a number of endpoints but not all of them are hit in the same proportion.  With this you can stress your service and more realistic conditions.  

*Note*: The sum of the distributions of all scenarios must add up to `1` or  the execution will fail.

### DSL :: Value Generators
There are a number of functions you can use to generated values randomly. 
The goal of it is define just one scenario and allow the engine to select in each request a random value in order to replicate a more realistic use case. 
This functions can be used in the following sections of the `DSL`: 
* host
* path
* body 

A function in the DSL is place holded by two curly braces like so `{{` & `}}`: 
```
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
### PENDING

| **Feature** | **Status** | *Notes* |
| --------|-------|------- |
|Tests dude  | [PENDING] ||
|Deal with TODOs  | [PENDING] ||
|Reporting |[PENDING]| Printers Inject printers by configuration|
|Reporting |[PENDING]| Different type of output reports|
|Reporting - As a module|[PENDING]| Request through time, percentiles through time, latency|
|Reporting - Reporting Suggestion|[PENDING]| In the event of several DNS resolution suggest change url|
|Auto Generation of DSL  |[PENDING]| Swagger Scaffolding|
|Events  |[PENDING]| Generate load by creating events|
|Payload size analisis   | [PENDING] ||
|Fire Sale as module  | [PENDING] | Make a module out of this and move actual job testing outside|
|Warm up request|[PENDING]||
|Speed up to a desired number of request vs concurrent users  | [PENDING] ||
|DSL short for Capacity testing | [PENDING] |Keep adding workers up until timeout is constant|
|DSL short for Spike  Traffic | [PENDING] ||
|Check Job Buffer size to size it based on the expected max RPS| [PENDING] |we size the job buffer to at least the number of jobs so not to choke the producer|
|Numbers with network latency should be a configuration| [PENDING] | |
|BUG in the combination of number of jobs workers and running time  | [PENDING] | |
|Support for gPRC|[PENDING]||

## PLATFORM TODO
| **Feature** | **Status** | *Notes* |
| --------|-------|------- |
|Soak Testing  | [PENDING] ||