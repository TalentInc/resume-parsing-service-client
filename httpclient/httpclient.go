package httpclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
)

// For ease of unit testing.
// Declaring these functions as global variables
// makes it easy to mock them.
var (
	handleUnsuccessfulResponse = func(url string, resp *http.Response,
		receivedError error) error {
		if resp != nil {
			if resp.StatusCode >= http.StatusBadRequest {
				httpErr := &HttpError{
					Url:        url,
					StatusCode: resp.StatusCode,
				}
				defer resp.Body.Close()
				respErr, err := ioReadAll(resp.Body)
				if err != nil {
					httpErr.Err = errors.Wrap(err, "parsing response")
					return httpErr
				}
				httpErr.Body = string(respErr)
				httpErr.Err = receivedError
				return httpErr
			}
		}
		if receivedError != nil {
			return &HttpError{
				Url: url,
				Err: receivedError,
			}
		}
		return nil
	}
	decodeResponse = func(url string, resp *http.Response, v interface{}) error {
		if v != nil {
			if resp != nil {
				defer resp.Body.Close()
				if err := jsonDecode(resp.Body, v); err != nil {
					return &HttpError{
						Url:        url,
						StatusCode: resp.StatusCode,
						Err:        errors.Wrap(err, "decoding response"),
					}
				}
			}
		}
		return nil
	}
	jsonDecode = func(r io.Reader, data interface{}) error {
		return json.NewDecoder(r).Decode(data)
	}
	ioReadAll = func(r io.Reader) ([]byte, error) {
		return io.ReadAll(r)
	}
	dumpRequestOut = httputil.DumpRequestOut
)

// Client defines the interface for an HTTP client that can send requests.
type Client interface {
	// SendRequest sends an HTTP request and returns the response.
	SendRequest(req *http.Request) (*http.Response, error)

	// SendRequestAndUnmarshallJsonResponse sends an HTTP request and
	// unmarshals the JSON response into a provided variable.
	SendRequestAndUnmarshallJsonResponse(req *http.Request, v interface{}) (*http.Response, error)
}

// client implements Client interface.
type client struct {
	retryableHttpClient retryableHttpClient
	maxIdleConns        int
	maxIdleConnsPerHost int
	maxConnsPerHost     int
	maxRetries          int
	checkRetryPolicy    retryablehttp.CheckRetry
	retryWaitMin        time.Duration
	retryWaitMax        time.Duration
	requestDumpLogger   func(dump []byte)
	dumpRequestBody     bool
}

// This construct aids in mocking by allowing users to implement only
// the functions they need for tests or other use cases.
var _ Client = (*client)(nil)

// doNotRetryPolicy is the default retry policy
// when a custom one is not provided, meaning that
// the request will not be retried.
// It is necessary because when a custom retry policy is
// not defined, `retryablehttp.DefaultRetryPolicy` becomes
// the default one, and it will always retry when status code >= 500,
// swalling the *http.Response.
func doNotRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	return false, nil
}

// patchRetryableClient patches retryable http client.
func patchRetryableClient(c *client) {
	c.retryableHttpClient.SetRetryMax(c.maxRetries)
	c.retryableHttpClient.SetRetryWaitMin(c.retryWaitMin)
	c.retryableHttpClient.SetRetryWaitMax(c.retryWaitMax)
	// If no custom check retry policy is provided,
	// doNotRetryPolicy will be used.
	c.retryableHttpClient.SetCheckRetry(doNotRetryPolicy)
	if c.checkRetryPolicy != nil {
		c.retryableHttpClient.SetCheckRetry(c.checkRetryPolicy)
	}
}

// newClient returns a new Client with options loaded.
func newClient(options []Option) *client {
	client := new(client)
	for _, option := range options {
		option(client)
	}
	return client
}

// New returns a new Client.
func New(options ...Option) Client {
	client := newClient(options)
	client.retryableHttpClient = &retryableHttpClientWrapper{retryablehttp.NewClient()}
	patchRetryableClient(client)
	return client
}

// do performs a request and parses the response to the given interface, if provided.
func (c *client) do(req *retryablehttp.Request, v interface{}) (*http.Response, error) {
	resp, err := c.retryableHttpClient.Do(req)
	if err := handleUnsuccessfulResponse(req.URL.String(), resp, err); err != nil {
		return resp, err
	}
	if err := decodeResponse(req.URL.String(), resp, v); err != nil {
		return resp, err
	}
	return resp, nil
}

// logRequestDump logs the request dump.
func (c *client) logRequestDump(req *http.Request) {
	if c.requestDumpLogger != nil {
		dump, err := dumpRequestOut(req, c.dumpRequestBody)
		if err == nil {
			c.requestDumpLogger(dump)
		}
	}
}

// sendRequest sends a request with or without payload.
func (c *client) sendRequest(req *http.Request, v interface{}) (*http.Response, error) {
	c.logRequestDump(req)
	resp, err := c.do(&retryablehttp.Request{Request: req}, v)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// SendRequest sends an HTTP request and returns an HTTP response.
func (c *client) SendRequest(req *http.Request) (*http.Response, error) {
	return c.sendRequest(req, nil)
}

// SendRequestAndUnmarshallJsonResponse sends an HTTP request \
// and unmarshalls the responseBody to the given interface.
func (c *client) SendRequestAndUnmarshallJsonResponse(req *http.Request, v interface{}) (*http.Response, error) {
	return c.sendRequest(req, v)
}
