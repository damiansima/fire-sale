package engine

import (
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func Test_doRequest(t *testing.T) {
	type args struct {
		method    string
		url       string
		reqBody   io.Reader
		headers   map[string]string
		timeout   time.Duration
		transport http.RoundTripper
	}
	tests := []struct {
		name string
		args args
		want Result
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := doRequest(tt.args.method, tt.args.url, tt.args.reqBody, tt.args.headers, tt.args.timeout, tt.args.transport); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("doRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
