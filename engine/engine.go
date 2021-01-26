package engine

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Defines the way the engine will get to the defined amount of workers
type RampUp struct {
	Step int
	Time time.Duration
}

type Certificates struct {
	ClientCertFile string
	ClientKeyFile  string
	CaCertFile     string
}

var DefaultRampUp RampUp = RampUp{Step: 1, Time: 0}

// TODO this probably needs a new name
type Scenario struct {
	Id           int
	Name         string
	Distribution float32
	JobCreator   func(id int) Job
}

type Result struct {
	// TODO start and end are part of the Trace really remove it
	Start   time.Time
	End     time.Time
	Trace   Trace
	Status  int
	Timeout bool
	job     Job
}

func ConfigureLog(logLevel string) {
	log.SetFormatter(&log.TextFormatter{})

	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)
	}
}

func Run(noOfWorkers int, noOfRequest int, noOfWarmupJobs int, testDuration time.Duration, warmupDuration time.Duration, maxSpeedPerSecond int, scenarios []Scenario, rampUp RampUp, certificates Certificates, reportType, reportFilePath string) {
	log.Infof("Parameters - # of Request [%d] - Test Duration [%s] - Concurrent Users [%d] - Max RPS [%d] - Ramp Up [%v]", noOfRequest, testDuration, noOfWorkers, maxSpeedPerSecond, rampUp)
	start := time.Now()

	jobBufferSize := 15
	resultBufferSize := 1000 * noOfWorkers
	jobs := make(chan Job, jobBufferSize)
	results := make(chan Result, resultBufferSize)

	go AllocateJobs(noOfRequest, noOfWarmupJobs, testDuration, warmupDuration, maxSpeedPerSecond, scenarios, jobs)

	done := make(chan bool)
	report := Report{}
	go ConsumeResults(results, done, &report)

	if (RampUp{}) == rampUp {
		rampUp = DefaultRampUp
	}
	runWorkers(noOfWorkers, rampUp, certificates, jobs, results)
	<-done

	printReport(report, reportType, reportFilePath)
	log.Infof("Execution took [%s]", time.Now().Sub(start))
}

func runWorkers(noOfWorkers int, rampUp RampUp, certificates Certificates, jobs chan Job, results chan Result) {
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
			go work(i, &wg, jobs, results, certificates)
			i++
		}
		log.Debugf("Pacing for [%s] ...", pace)
		time.Sleep(pace)
	}
	wg.Wait()
	close(results)
	log.Infof("Workers finish job pool")
}

func work(workerId int, wg *sync.WaitGroup, jobs chan Job, results chan Result, certificates Certificates) {
	var transport http.RoundTripper
	for job := range jobs {
		if job.AllowConnectionReuse {
			transport = http.DefaultTransport
		} else {
			workerTransport := newDefaultTransportWithTLSSupport(certificates.ClientCertFile, certificates.ClientKeyFile, certificates.CaCertFile)
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

	traceableTransport := &TraceableTransport{Trace: &Trace{}}
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
		return Result{Start: start, End: end, Timeout: isTimeOut, Trace: *traceableTransport.Trace}
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Tracef("Fail to read response %s", err)
	} else {
		log.Tracef("Resp Headers [%v]", resp.Header)
		log.Tracef(string(body))
	}

	return Result{Start: start, End: end, Status: resp.StatusCode, Trace: *traceableTransport.Trace}
}

// This method ensures a new instance of the Transport struct
// The goal is use it to force no reuse of connections between go routines
// and simulate different users
func newDefaultTransportWithTLSSupport(clientCertFile string, clientKeyFile string, caCertFile string) http.RoundTripper {
	tlsConfig := buildTlsConfig(clientCertFile, clientKeyFile, caCertFile)

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
		TLSClientConfig:       &tlsConfig,
	}

	return newTransport
}

func buildTlsConfig(clientCertFile string, clientKeyFile string, caCertFile string) tls.Config {
	var tlsConfig tls.Config
	// TODO WE MAY not be releasing the files properly after reading it
	if "" != clientCertFile && "" != clientKeyFile && caCertFile != "" {
		cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
		if err != nil {
			log.Fatalf("Error creating x509 keypair from client cert file %s and client key file %s", clientCertFile, clientKeyFile)
		}

		caCert, err := ioutil.ReadFile(caCertFile)
		if err != nil {
			log.Fatalf("Error opening cert file %s, Error: %s", caCertFile, err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig = tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: true, // TODO this must be configuration
		}
	}
	return tlsConfig
}
