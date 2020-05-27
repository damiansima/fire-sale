package engine_test

import (
	"crypto/tls"
	"github.com/damiansima/fire-sale/engine"
	"github.com/stretchr/testify/assert"
	"net/http/httptrace"
	"net/textproto"
	"testing"
)

func TestTraceableTransport_GetConn(t *testing.T) {
	traceableTransport := &engine.TraceableTransport{Trace: &engine.Trace{}}
	assert.True(t, traceableTransport.Trace.GetConnTime.IsZero())

	traceableTransport.GetConn("6666")
	assert.False(t, traceableTransport.Trace.GetConnTime.IsZero())
}

func TestTraceableTransport_ConnectStart(t *testing.T) {
	traceableTransport := &engine.TraceableTransport{Trace: &engine.Trace{}}
	assert.True(t, traceableTransport.Trace.ConnectStartTime.IsZero())

	traceableTransport.ConnectStart("someNetwork", "someAddres")
	assert.False(t, traceableTransport.Trace.ConnectStartTime.IsZero())
}

func TestTraceableTransport_DNSDone(t *testing.T) {
	traceableTransport := &engine.TraceableTransport{Trace: &engine.Trace{}}
	assert.True(t, traceableTransport.Trace.DNSDoneTime.IsZero())

	var info httptrace.DNSDoneInfo
	traceableTransport.DNSDone(info)
	assert.False(t, traceableTransport.Trace.DNSDoneTime.IsZero())
}

func TestTraceableTransport_DNSStart(t *testing.T) {
	traceableTransport := &engine.TraceableTransport{Trace: &engine.Trace{}}
	assert.True(t, traceableTransport.Trace.DNSStartTime.IsZero())

	var info httptrace.DNSStartInfo
	traceableTransport.DNSStart(info)
	assert.False(t, traceableTransport.Trace.DNSStartTime.IsZero())
}

func TestTraceableTransport_Got100Continue(t *testing.T) {
	traceableTransport := &engine.TraceableTransport{Trace: &engine.Trace{}}
	assert.True(t, traceableTransport.Trace.Got100ContinueTime.IsZero())

	traceableTransport.Got100Continue()
	assert.False(t, traceableTransport.Trace.Got100ContinueTime.IsZero())
}

func TestTraceableTransport_Got1xxResponse(t *testing.T) {
	traceableTransport := &engine.TraceableTransport{Trace: &engine.Trace{}}
	assert.True(t, traceableTransport.Trace.Got1xxResponseTime.IsZero())

	var code int
	var header textproto.MIMEHeader
	traceableTransport.Got1xxResponse(code, header)
	assert.False(t, traceableTransport.Trace.Got1xxResponseTime.IsZero())
}

func TestTraceableTransport_GotConn(t *testing.T) {
	traceableTransport := &engine.TraceableTransport{Trace: &engine.Trace{}}
	assert.True(t, traceableTransport.Trace.GetConnTime.IsZero())

	var info httptrace.GotConnInfo
	traceableTransport.GotConn(info)
	assert.False(t, traceableTransport.Trace.GotConnTime.IsZero())
}

func TestTraceableTransport_GotFirstResponseByte(t *testing.T) {
	traceableTransport := &engine.TraceableTransport{Trace: &engine.Trace{}}
	assert.True(t, traceableTransport.Trace.GotFirstResponseByteTime.IsZero())

	traceableTransport.GotFirstResponseByte()
	assert.False(t, traceableTransport.Trace.GotFirstResponseByteTime.IsZero())
}

func TestTraceableTransport_PutIdleConn(t *testing.T) {
	traceableTransport := &engine.TraceableTransport{Trace: &engine.Trace{}}
	assert.True(t, traceableTransport.Trace.PutIdleConnTime.IsZero())

	var err error
	traceableTransport.PutIdleConn(err)
	assert.False(t, traceableTransport.Trace.PutIdleConnTime.IsZero())
}

func TestTraceableTransport_TLSHandshakeDone(t *testing.T) {
	traceableTransport := &engine.TraceableTransport{Trace: &engine.Trace{}}
	assert.True(t, traceableTransport.Trace.TLSHandshakeDoneTime.IsZero())

	var err error
	var state tls.ConnectionState
	traceableTransport.TLSHandshakeDone(state, err)
	assert.False(t, traceableTransport.Trace.TLSHandshakeDoneTime.IsZero())
}

func TestTraceableTransport_TLSHandshakeStart(t *testing.T) {
	traceableTransport := &engine.TraceableTransport{Trace: &engine.Trace{}}
	assert.True(t, traceableTransport.Trace.TLSHandshakeStartTime.IsZero())

	traceableTransport.TLSHandshakeStart()
	assert.False(t, traceableTransport.Trace.TLSHandshakeStartTime.IsZero())
}

func TestTraceableTransport_Wait100Continue(t *testing.T) {
	traceableTransport := &engine.TraceableTransport{Trace: &engine.Trace{}}
	assert.True(t, traceableTransport.Trace.Wait100ContinueTime.IsZero())

	traceableTransport.Wait100Continue()
	assert.False(t, traceableTransport.Trace.Wait100ContinueTime.IsZero())
}

func TestTraceableTransport_WroteRequest(t *testing.T) {
	traceableTransport := &engine.TraceableTransport{Trace: &engine.Trace{}}
	assert.True(t, traceableTransport.Trace.WroteRequestTime.IsZero())

	var info httptrace.WroteRequestInfo
	traceableTransport.WroteRequest(info)
	assert.False(t, traceableTransport.Trace.WroteRequestTime.IsZero())
}
