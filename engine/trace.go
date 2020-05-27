package engine

import (
	"crypto/tls"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httptrace"
	"net/textproto"
	"time"
)

// TraceableTransport is an http.RoundTripper that keeps track of the in-flight
// request and implements hooks to report HTTP tracing events.
type TraceableTransport struct {
	current *http.Request
	trace   *Trace
}
//TODO we may need to remove all the tracef with the call
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
}

func (t *TraceableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// TODO not really sure what to use this for yet maybe remove
	t.current = req
	return http.DefaultTransport.RoundTrip(req)
}

func (t *TraceableTransport) GetConn(hostPort string) {
	t.trace.GetConnTime = time.Now()
	log.Tracef("GetConn %s", hostPort)
}

func (t *TraceableTransport) GotConn(info httptrace.GotConnInfo) {
	t.trace.GotConnTime = time.Now()
	log.Tracef("GotConn %v", info)
}

func (t *TraceableTransport) PutIdleConn(err error) {
	t.trace.PutIdleConnTime = time.Now()
	log.Tracef("PutIdleConn")
}

func (t *TraceableTransport) GotFirstResponseByte() {
	t.trace.GotFirstResponseByteTime = time.Now()
	log.Tracef("GotFirstResponseByte %s", time.Now())
}

func (t *TraceableTransport) Got100Continue() {
	log.Tracef("Got100Continue")
}

func (t *TraceableTransport) DNSStart(info httptrace.DNSStartInfo) {
	t.trace.DNSStartTime = time.Now()
	log.Tracef("DNSStart %v %s", info, time.Now())
}

func (t *TraceableTransport) DNSDone(info httptrace.DNSDoneInfo) {
	t.trace.DNSDoneTime = time.Now()
	log.Tracef("DNSDone %v %s", info, time.Now())
}

func (t *TraceableTransport) ConnectStart(network, addr string) {
	t.trace.ConnectStartTime = time.Now()
	log.Tracef("ConnectStart %s -- %s", network, addr)
}

func (t *TraceableTransport) ConnectDone(network, addr string, err error) {
	t.trace.ConnectDoneTime = time.Now()
	log.Tracef("ConnectDone %s -- %s", network, addr)
}

func (t *TraceableTransport) TLSHandshakeStart() {
	t.trace.TLSHandshakeStartTime = time.Now()
	log.Tracef("TLSHandshakeStart")
}

func (t *TraceableTransport) TLSHandshakeDone(state tls.ConnectionState, err error) {
	t.trace.TLSHandshakeDoneTime = time.Now()
	log.Tracef("TLSHandshakeDone ")
}

func (t *TraceableTransport) WroteRequest(info httptrace.WroteRequestInfo) {
	t.trace.WroteRequestTime = time.Now()
	log.Tracef("WroteRequest %v %s", info, time.Now())
}

func (t *TraceableTransport) Got1xxResponse(code int, header textproto.MIMEHeader) error {
	log.Tracef("Got1xxResponse %d -- %v", code, header)
	return nil
}

func (t *TraceableTransport) Wait100Continue() {
	log.Tracef("Wait100Continue")
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
