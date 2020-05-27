package engine

import (
	"github.com/influxdata/tdigest"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"
)

// Defines the way the engine will get to the defined amount of workers
type RampUp struct {
	Step int
	Time time.Duration
}

var DefaultRampUp RampUp = RampUp{Step: 1, Time: 0}

type Job struct {
	Id                   int
	Method               string
	Url                  string
	ReqBody              io.Reader
	Headers              map[string]string
	Timeout              time.Duration
	AllowConnectionReuse bool
}
type Result struct {
	// TODO start and end are part of the trace really remove it
	Start   time.Time
	End     time.Time
	Trace   Trace
	Status  int
	Timeout bool
	job     Job
}

// AllocateJobs creates jobs and adds them to the jobs queue
// It receives noOfJobs and testDurationMs, if the second is grated than 0 it takes precedences and keeps
// pushing jobs during the defined period. If not the specified number of jobs will be created
func AllocateJobs(noOfJobs int, testDuration time.Duration, maxSpeedPerSecond int, jobCreator func(id int) Job, jobs chan Job) {
	log.Debugf("Allocating jobs ...")
	if testDuration > 0 {
		keepRunning := true
		go func() {
			log.Debugf("Allocating for [%d]ms", testDuration)
			<-time.After(testDuration)
			log.Debugf("Stop allocation")
			keepRunning = false
		}()

		if maxSpeedPerSecond > 0 {
			log.Debugf("Max request per second [%d]", maxSpeedPerSecond)
			for i := 0; keepRunning; {
				for j := 0; j < maxSpeedPerSecond; j++ {
					allocateJob(i, jobCreator, jobs)
					i++
				}
				time.Sleep(1 * time.Second)
			}
		} else {
			for i := 0; keepRunning; i++ {
				allocateJob(i, jobCreator, jobs)
			}
		}
	} else {
		log.Debugf("Allocating [%d]job", noOfJobs)
		for i := 0; i < noOfJobs; i++ {
			allocateJob(i, jobCreator, jobs)
		}
		log.Debugf("Stop allocation")
	}

	close(jobs)
	log.Debugf("Allocating done")
}

func RunWorkers(noOfWorkers int, rampUp RampUp, jobs chan Job, results chan Result) {
	log.Infof("Running [%d] concurrent workers ...", noOfWorkers)
	var wg sync.WaitGroup

	// TODO BUG rampUp.Step can not be < 0
	// TODO BUG rampUp.Step can not be > noOfWorkers
	// TODO test: step can't go over, noOfWorker
	steps := noOfWorkers / rampUp.Step
	pace := time.Duration(rampUp.Time.Nanoseconds() / int64(steps))

	log.Debugf("Ramping up in [%d] steps...", steps)
	for i := 0; i < noOfWorkers; {
		for s := 0; i < noOfWorkers && s < rampUp.Step; s++ {
			log.Debugf("Starting worker [%d]  ...", i)
			wg.Add(1)
			go work(i, &wg, jobs, results)
			i++
		}
		log.Debugf("Pacing for [%s] ...", pace)
		time.Sleep(pace)
	}
	wg.Wait()
	close(results)
	log.Infof("Workers finish job pool")
}

// TODO this needs to be moved
func ConsumeResults(results chan Result, done chan bool) {
	var durationSum time.Duration
	var durationRequestSum time.Duration
	var count int64

	var failCount int64
	var successCount int64
	var timeoutCount int64

	// TODO THIS SHOULD BE ACCUMULATED FOR REPORT PURPOSES
	var last int64
	go func() {
		for _ = range time.Tick(10 * time.Second) {
			requestPerPeriod := count - last
			log.Infof("Request per 10 second [%d] | per 1 second [%d]...", requestPerPeriod, requestPerPeriod/10)
			last = count
		}
	}()

	//TODO we need to change this value and do memory profile
	td := tdigest.NewWithCompression(100000)
	var elapsedNetworkLast time.Duration

	// TODO allow for a channel to plot data points
	for result := range results {
		count++
		elapsedOverall := result.End.Sub(result.Start)
		elapsedNetwork := result.Trace.ConnectDoneTime.Sub(result.Trace.ConnectStartTime)
		elapsedRequest := result.Trace.GotFirstResponseByteTime.Sub(result.Trace.WroteRequestTime)
		//TODO WE ASSUME NETWORK AS LATENCY MAY BE KILL IT?
		if elapsedNetwork != 0 {
			elapsedNetworkLast = elapsedNetwork
			log.Tracef("change network time")
		}
		// TODO this measurement should be able to turn on and off
		actualServerTime := elapsedRequest - elapsedNetworkLast
		if actualServerTime < 0 {
			actualServerTime = -1 * actualServerTime
		}
		// TODO we should not account failed request but we should account timeout
		durationSum += elapsedOverall
		durationRequestSum += actualServerTime
		td.Add(actualServerTime.Seconds(), 1)

		log.Tracef("The job id [%d] lasted [%s||%s||%s] status [%d] - timeout [%s]", result.job.Id, elapsedOverall, elapsedRequest, actualServerTime, result.Status, result.Timeout)
		if result.Timeout {
			timeoutCount++
		} else if result.Status > 0 && result.Status < 300 {
			successCount++
		} else {
			failCount++
		}
	}
	// TODO this needs to me moved to a report module
	// TODO BUG: Fail percentage is not accurate
	log.Infof("Success [%f%%] - Fail [%f%%]", float32((successCount*100)/count), float32(((timeoutCount+failCount)*100)/count))
	// TODO average time should taken from configuration with/without latency
	log.Infof("Request total [%d] average [%s] ", count, time.Duration(durationSum.Nanoseconds()/count))
	log.Infof("Request total [%d] average [%s] ", count, time.Duration(durationRequestSum.Nanoseconds()/count))
	// TODO we may need to change this lib at least inject it
	log.Infof("99th %fms", td.Quantile(0.99)/time.Millisecond.Seconds())
	log.Infof("90th %fms", td.Quantile(0.9)/time.Millisecond.Seconds())
	log.Infof("75th %fms", td.Quantile(0.75)/time.Millisecond.Seconds())
	log.Infof("50th %fms", td.Quantile(0.5)/time.Millisecond.Seconds())

	log.Infof("Timeout [%d] - Fail [%d] - Success [%d]  ", timeoutCount, failCount, successCount)

	done <- true
}

func allocateJob(id int, jobCreator func(id int) Job, jobs chan Job) {
	log.Debugf("Allocating job [%d]", id)
	jobs <- jobCreator(id)
}

func work(workerId int, wg *sync.WaitGroup, jobs chan Job, results chan Result) {
	var transport http.RoundTripper
	workerTransport := NewDefaultTransport()
	for job := range jobs {
		if job.AllowConnectionReuse {
			transport = http.DefaultTransport
		} else {
			transport = workerTransport
		}
		log.Debugf("Worker [%d] running job [%d] ...", workerId, job.Id)
		output := doRequest(job.Method, job.Url, job.ReqBody, job.Headers, job.Timeout, transport)
		output.job = job
		results <- output
	}
	wg.Done()
}

func doRequest(method, url string, reqBody io.Reader, headers map[string]string, timeout time.Duration, transport http.RoundTripper) Result {
	log.Tracef("Making request  %s - %s ", method, url)

	request, err := http.NewRequest(method, url, reqBody)
	if headers != nil {
		for k, v := range headers {
			log.Tracef("Setting header %s : %s", k, v)
			request.Header.Set(k, v)
		}
	}

	if err != nil {
		log.Tracef("Fail to create request %s", err)
		// TODO this should be different
		return Result{}
	}

	client := http.Client{Transport: transport}
	client.Timeout = timeout
	log.Tracef("Defined timout %s", client.Timeout)

	traceableTransport := &TraceableTransport{trace: &Trace{}}
	trace := NewTrace(*traceableTransport)
	request = request.WithContext(httptrace.WithClientTrace(request.Context(), trace))

	start := time.Now()
	resp, err := client.Do(request)
	end := time.Now()

	if err != nil {
		log.Tracef("Fail to execute request %s", err)
		isTimeOut := false
		if err, ok := err.(net.Error); ok && err.Timeout() {
			isTimeOut = true
		}
		return Result{Start: start, End: end, Timeout: isTimeOut, Trace: *traceableTransport.trace}
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Tracef("Fail to read response %s", err)
	} else {
		log.Tracef("Resp Headers [%v]", resp.Header)
		log.Tracef(string(body))
	}

	return Result{Start: start, End: end, Status: resp.StatusCode, Trace: *traceableTransport.trace}
}

// This method ensures a new instance of the Transport struct
// The goal is use it to force no reuse of connections between go routines
// and simulate different users
func NewDefaultTransport() http.RoundTripper {
	var newTransport http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return newTransport
}
