package engine

import (
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httptrace"
	"net/textproto"
	"time"
)

// TraceableTransport is an http.RoundTripper that keeps track of the in-flight
// request and implements hooks to report HTTP tracing events.
type TraceableTransport struct {
	Current   *http.Request
	Trace     *Trace
	LogEnable bool
}

type Trace struct {
	GotConnTime              time.Time
	GetConnTime              time.Time
	DNSStartTime             time.Time
	DNSDoneTime              time.Time
	ConnectStartTime         time.Time
	ConnectDoneTime          time.Time
	TLSHandshakeStartTime    time.Time
	TLSHandshakeDoneTime     time.Time
	PutIdleConnTime          time.Time
	WroteRequestTime         time.Time
	GotFirstResponseByteTime time.Time
	Got100ContinueTime       time.Time
	Got1xxResponseTime       time.Time
	Wait100ContinueTime      time.Time
}

func (t *TraceableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// TODO not really sure what to use this for yet maybe remove
	t.Current = req
	return http.DefaultTransport.RoundTrip(req)
}

func (t *TraceableTransport) GetConn(hostPort string) {
	t.Trace.GetConnTime = time.Now()
	t.log(fmt.Sprintf("GetConn %s", hostPort))
}

func (t *TraceableTransport) GotConn(info httptrace.GotConnInfo) {
	t.Trace.GotConnTime = time.Now()
	t.log(fmt.Sprintf("GotConn %v", info))
}

func (t *TraceableTransport) PutIdleConn(err error) {
	t.Trace.PutIdleConnTime = time.Now()
	t.log("PutIdleConn")
}

func (t *TraceableTransport) GotFirstResponseByte() {
	t.Trace.GotFirstResponseByteTime = time.Now()
	t.log(fmt.Sprintf("GotFirstResponseByte %s", time.Now()))
}

func (t *TraceableTransport) Got100Continue() {
	t.Trace.Got100ContinueTime = time.Now()
	t.log("Got100Continue")
}

func (t *TraceableTransport) DNSStart(info httptrace.DNSStartInfo) {
	t.Trace.DNSStartTime = time.Now()
	t.log(fmt.Sprintf("DNSStart %v %s", info, time.Now()))
}

func (t *TraceableTransport) DNSDone(info httptrace.DNSDoneInfo) {
	t.Trace.DNSDoneTime = time.Now()
	t.log(fmt.Sprintf("DNSDone %v %s", info, time.Now()))
}

func (t *TraceableTransport) ConnectStart(network, addr string) {
	t.Trace.ConnectStartTime = time.Now()
	t.log(fmt.Sprintf("ConnectStart %s -- %s", network, addr))
}

func (t *TraceableTransport) ConnectDone(network, addr string, err error) {
	t.Trace.ConnectDoneTime = time.Now()
	t.log(fmt.Sprintf("ConnectDone %s -- %s", network, addr))
}

func (t *TraceableTransport) TLSHandshakeStart() {
	t.Trace.TLSHandshakeStartTime = time.Now()
	t.log("TLSHandshakeStart")
}

func (t *TraceableTransport) TLSHandshakeDone(state tls.ConnectionState, err error) {
	t.Trace.TLSHandshakeDoneTime = time.Now()
	t.log("TLSHandshakeDone ")
}

func (t *TraceableTransport) WroteRequest(info httptrace.WroteRequestInfo) {
	t.Trace.WroteRequestTime = time.Now()
	t.log(fmt.Sprintf("WroteRequest %v %s", info, time.Now()))
}

func (t *TraceableTransport) Got1xxResponse(code int, header textproto.MIMEHeader) error {
	t.Trace.Got1xxResponseTime = time.Now()
	t.log(fmt.Sprintf("Got1xxResponse %d -- %v", code, header))
	return nil
}

func (t *TraceableTransport) Wait100Continue() {
	t.Trace.Wait100ContinueTime = time.Now()
	t.log("Wait100Continue")
}
func (t *TraceableTransport) log(message string) {
	if t.LogEnable {
		log.Tracef(message)
	}
}
func NewTrace(traceableTransport TraceableTransport) *httptrace.ClientTrace {
	trace := &httptrace.ClientTrace{
		GetConn:              traceableTransport.GetConn,
		GotConn:              traceableTransport.GotConn,
		PutIdleConn:          traceableTransport.PutIdleConn,
		GotFirstResponseByte: traceableTransport.GotFirstResponseByte,
		Got100Continue:       traceableTransport.Got100Continue,
		Got1xxResponse:       traceableTransport.Got1xxResponse,
		Wait100Continue:      traceableTransport.Wait100Continue,
		DNSStart:             traceableTransport.DNSStart,
		DNSDone:              traceableTransport.DNSDone,
		ConnectStart:         traceableTransport.ConnectStart,
		ConnectDone:          traceableTransport.ConnectDone,
		TLSHandshakeStart:    traceableTransport.TLSHandshakeStart,
		TLSHandshakeDone:     traceableTransport.TLSHandshakeDone,
		WroteRequest:         traceableTransport.WroteRequest,
	}
	return trace
}
