package mock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// HTTPClient returns a new HTTP client using the provided RountTrip
// function.
func HTTPClient(f func(r *http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{
		Transport: &transport{f},
	}
}

// HTTPResponse returns a new HTTP response using the provided status code
// and raw response body.
func HTTPResponse(code int, body string) *http.Response {
	return &http.Response{
		StatusCode: code,
		Body:       ioutil.NopCloser(strings.NewReader(body)),
	}
}

// HTTPResponseJSON returns a new HTTP response using the provided status code
// and JSON encoding of "v".
func HTTPResponseJSON(code int, v interface{}) *http.Response {
	body, err := json.Marshal(v)
	if err != nil {
		msg := fmt.Sprintf("unable to marshal http response json: %s", err.Error())
		panic(msg)
	}
	return &http.Response{
		StatusCode: code,
		Body:       ioutil.NopCloser(bytes.NewReader(body)),
	}
}

// transport implements the HTTP RoundTripper interface.
type transport struct {
	f func(*http.Request) (*http.Response, error)
}

// RoundTrip calls the transport's RoundTrip function.
func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	return t.f(r)
}
