# FIRE SALE
[![CircleCI](https://circleci.com/gh/damiansima/fire-sale.svg?style=shield&circle-token=e2f5b5b9357eb8cc430df5619e92925502ea606f)](<https://app.circleci.com/pipelines/github/damiansima/fire-sale>)

``Everything must GO(lang)``

FireSale is a performance testing tool designed to be ligtwayth and fast. 
Its simple DSL allows you to generate load and spike like traffic in order 
to stress your services in a way that reflects your production traffic. 

By unit of time 
* Spike traffic
* Sustained traffic
* Paced traffic

## REQUIREMENTS
It requires GO 1.15

## BUILDING
```
$ go build
$ go test -cover ./...
```

## USAGE
```
$ ./fire-sale
  -config string
    	 Path to the test-configuration.yml
  -log string
    	 Define the log level [panic|fatal|error|warn|info|debug|trace] (default "info")
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
### PENDING

| **Feature** | **Status** | *Notes* |
| --------|-------|------- |
|Auto Generation of DSL  |[PENDING]| Swagger Scaffolding|
|Reporting |[PENDING]| Printers Inject printers by configuration|
|Reporting |[PENDING]| Different type of output reports|
|Reporting - As a module|[PENDING]| Request through time, percentiles through time, latency|
|Reporting - Reporting Suggestion|[PENDING]| In the event of several DNS resolution suggest change url|
|Events  |[PENDING]| Generate load by creating events|
|Tests dude  | [PENDING] ||
|Deal with TODOs  | [PENDING] ||
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


# DSL 
One of the goals behind FireSale is to make it supper simple to use. 
With that in mind we've come up with the following DSL which could be feed to the engine in order to run your tests:

```
name: da test
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
## DSL :: Parameters
### DSL :: Parameters :: Rampup
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