# fire-sale

``Everything must go(lang)``
Fire sale is a performance testing tool designed to be ligtwayth and fast. 

It's simple DSL allows you to generate load and spike like traffic. 

By unit of time 
* Spike traffic
* Sustained traffic
* Paced traffic

## TODO
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
### PENDING
| **Feature** | **Status** | *Notes* |
| --------|-------|------- |
|Multiple scenario | [PENDING] | report by job of request and percentiles|
|Tests dude  | [PENDING] ||
|Deal with TODOs  | [PENDING] ||
|Speed up to a desired number of request vs concurrent users  | [PENDING] ||
|Check Job Buffer size to size it based on the expected max RPS| [PENDING] |we size the job buffer to at least the number of jobs so not to choke the producer|
|Fire Sale as module  | [PENDING] | Make a module out of this and move actual job testing outside|
|Warm up request|[PENDING]||
|Value generators|[PENDING]| to randomize request for those parameters that can allow for auto generation between min and max|
|Input files for request parametrization|[PENDING]||
|DSL|[PENDING]||
|Certificates handling  | [PENDING] ||
|Numbers with network latency should be a configuration| [PENDING] | |
|BUG in the combination of number of jobs workers and running time  | [PENDING] | |
|Capacity testing | [PENDING] |Keep adding workers up until timeout is constant|
|Spike  Traffic | [PENDING] ||
|Reporting As a module|[PENDING]| Request through time, percentiles through time, latency|
|Reporting Suggestion|[PENDING]| In the event of several DNS resolution suggest change url|
|Swagger Scaffolding|[PENDING]||
|Support for gPRC|[PENDING]||

## PLATFORM TODO
| **Feature** | **Status** | *Notes* |
| --------|-------|------- |
|Soak Testing  | [PENDING] ||

## REQUIREMENTS
## BUILDING
## RUN
