package engine

import (
	"crypto/tls"
	"crypto/x509"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptrace"
	"time"
)

func DoRequest(method, url string, reqBody io.Reader, headers map[string]string, timeout time.Duration, allowConnectionReuse bool, certificates Certificates) Result {
	log.Tracef("Making request  %s - %s ", method, url)

	transport := buildTransport(allowConnectionReuse, certificates)

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

func buildTransport(allowConnectionReuse bool, certificates Certificates) http.RoundTripper {
	var transport http.RoundTripper
	if allowConnectionReuse {
		transport = http.DefaultTransport
	} else {
		workerTransport := newDefaultTransportWithTLSSupport(certificates.ClientCertFile, certificates.ClientKeyFile, certificates.CaCertFile)
		transport = workerTransport
	}
	return transport
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
