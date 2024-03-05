package httpclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
)

type (
	dummyType struct {
		Key string `json:"key"`
	}
)

func TestDefaultPolicy(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer svr.Close()
	url := svr.URL
	client := New()
	// req, err := NewRequest(context.TODO(), http.MethodGet, url)
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf(`creating request for "%v": %v`, url, err)
	}
	_, err = client.SendRequest(req)
	expectedError := fmt.Errorf(`request to %s failed. `+
		`httpStatus: [ %d ] responseBody: [  ] `+
		`error: [ <nil> ]`, url, http.StatusBadGateway)
	require.NotNil(t, err)
	require.Equal(t, expectedError.Error(), err.Error())
}

func TestEofRetryPolicy(t *testing.T) {
	checkRetryPolicy := func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if err != nil {
			return strings.Contains(err.Error(), "EOF"), err
		}
		return false, err
	}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(errors.New("random error"))
	}))
	defer svr.Close()
	url := svr.URL
	client := New(WithMaxRetries(1), WithCheckRetryPolicy(checkRetryPolicy))
	// req, err := NewRequest(context.TODO(), http.MethodGet, url)
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf(`creating request for "%v": %v`, url, err)
	}
	_, err = client.SendRequest(req)
	expectedError := fmt.Errorf(`request to %s failed. `+
		`httpStatus: [ no status ] responseBody: [  ] `+
		`error: [ GET %s giving up after 2 attempt(s): Get "%s": EOF ]`, url, url, url)
	require.NotNil(t, err)
	require.Equal(t, expectedError.Error(), err.Error())
}

func TestNew(t *testing.T) {
	testCases := []struct {
		name                        string
		options                     []Option
		checkRetryPolicy            bool
		checkRequestDumpLogger      bool
		expectedMaxIdleConns        int
		expectedMaxIdleConnsPerHost int
		expectedMaxConnsPerHost     int
		expectedMaxRetries          int
		expectedRetryWaitMin        time.Duration
		expectedRetryWaitMax        time.Duration
		expectedRequestDumpLogger   func(dump []byte)
		expectedDumpRequestBody     bool
	}{
		{
			name:    "no options provided",
			options: []Option{},
		},
		{
			name: "with all options",
			options: []Option{
				WithCheckRetryPolicy(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
					return false, nil
				}),
				WithMaxIdleConns(1),
				WithMaxIdleConnsPerHost(1),
				WithMaxConnsPerHost(1),
				WithMaxRetries(1),
				WithRetryWaitMin(1 * time.Second),
				WithRetryWaitMax(1 * time.Second),
				WithRequestDumpLogger(func(dump []byte) {}, true),
			},
			expectedMaxIdleConns:        1,
			expectedMaxIdleConnsPerHost: 1,
			expectedMaxConnsPerHost:     1,
			expectedMaxRetries:          1,
			expectedRetryWaitMin:        1 * time.Second,
			expectedRetryWaitMax:        1 * time.Second,
			expectedDumpRequestBody:     true,
			checkRetryPolicy:            true,
			checkRequestDumpLogger:      true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := New(tc.options...)
			clientWrapper, ok := c.(*client)
			require.True(t, ok)
			require.True(t, ok)
			require.Equal(t, tc.expectedMaxIdleConns, clientWrapper.maxIdleConns)
			require.Equal(t, tc.expectedMaxIdleConnsPerHost, clientWrapper.maxIdleConnsPerHost)
			require.Equal(t, tc.expectedMaxConnsPerHost, clientWrapper.maxConnsPerHost)
			require.Equal(t, tc.expectedMaxRetries, clientWrapper.maxRetries)
			require.Equal(t, tc.expectedRetryWaitMin, clientWrapper.retryWaitMin)
			require.Equal(t, tc.expectedRetryWaitMax, clientWrapper.retryWaitMax)
			require.Equal(t, tc.expectedDumpRequestBody, clientWrapper.dumpRequestBody)
			if tc.expectedRequestDumpLogger != nil {
				require.NotNil(t, clientWrapper.requestDumpLogger)
			}
			if tc.checkRequestDumpLogger {
				require.NotNil(t, clientWrapper.requestDumpLogger)
			}
			if tc.checkRetryPolicy {
				require.NotNil(t, clientWrapper.checkRetryPolicy)
			}
		})
	}
}

func TestSendRequest(t *testing.T) {
	testCases := []struct {
		name               string
		requestDumpLogger  func(dump []byte)
		mockClosure        func(r *retryableHttpClientMock)
		ioReadAllMock      func(r io.Reader) ([]byte, error)
		dumpRequestOutMock func(req *http.Request, body bool) ([]byte, error)
		expectedError      error
	}{
		{
			name: "happy path",
			mockClosure: func(r *retryableHttpClientMock) {
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte(""))),
				}
				r.Resp = resp
			},
		},
		{
			name: "dumping request",
			requestDumpLogger: func(dump []byte) {
				expectedDump := "POST /some/path HTTP/1.1\r\nHost: localhost\r\nUser-Agent: Go-http-client/1.1\r\nContent-Length: 0\r\nAccept-Encoding: gzip\r\n\r\n"
				require.Equal(t, expectedDump, string(dump))
			},
			mockClosure: func(r *retryableHttpClientMock) {
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte(""))),
				}
				r.Resp = resp
			},
		},
		{
			name: "error when dumping request",
			dumpRequestOutMock: func(req *http.Request, body bool) ([]byte, error) {
				return nil, errors.New("random error")
			},
			requestDumpLogger: func(dump []byte) {
				require.Equal(t, "", string(dump))
			},
			mockClosure: func(r *retryableHttpClientMock) {
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte(""))),
				}
				r.Resp = resp
			},
		},
		{
			name: "unsuccessful response with responseBody",
			mockClosure: func(r *retryableHttpClientMock) {
				resp := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"some random error"}`))),
				}
				r.Resp = resp
			},
			expectedError: errors.New(`request to http://localhost/some/path failed. ` +
				`httpStatus: [ 500 ] responseBody: [ {"error":"some random error"} ] ` +
				`error: [ <nil> ]`),
		},
		{
			name: "unsuccessful response with responseBody, error when parsing it",
			mockClosure: func(r *retryableHttpClientMock) {
				resp := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"some random error"}`))),
				}
				r.Resp = resp
			},
			ioReadAllMock: func(r io.Reader) ([]byte, error) {
				return nil, errors.New("random error")
			},
			expectedError: errors.New(`request to http://localhost/some/path failed. ` +
				`httpStatus: [ 500 ] responseBody: [  ] error: [ parsing response: random error ]`),
		},
		{
			name: "unsuccessful response with empty body, with some other error",
			mockClosure: func(r *retryableHttpClientMock) {
				resp := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewReader([]byte(``))),
				}
				r.Resp = resp
				r.Err = errors.New("random error")
			},
			expectedError: errors.New(`request to http://localhost/some/path failed. ` +
				`httpStatus: [ 500 ] responseBody: [  ] error: [ random error ]`),
		},
		{
			name: "without response",
			mockClosure: func(r *retryableHttpClientMock) {
				r.Err = errors.New("random error")
			},
			expectedError: errors.New(`request to http://localhost/some/path failed. ` +
				`httpStatus: [ no status ] responseBody: [  ] error: [ random error ]`),
		},
	}
	originalIoReadAll := ioReadAll
	originalDumpRequestOut := dumpRequestOut
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				ioReadAll = originalIoReadAll
				dumpRequestOut = originalDumpRequestOut
			}()
			if tc.ioReadAllMock != nil {
				ioReadAll = tc.ioReadAllMock
			}
			if tc.dumpRequestOutMock != nil {
				dumpRequestOut = tc.dumpRequestOutMock
			}
			c := New(WithRequestDumpLogger(tc.requestDumpLogger, false))
			clientWrapper, ok := c.(*client)
			require.True(t, ok)
			retryableHttpClientMock := new(retryableHttpClientMock)
			tc.mockClosure(retryableHttpClientMock)
			clientWrapper.retryableHttpClient = retryableHttpClientMock
			req, err := http.NewRequest(http.MethodPost, "http://localhost/some/path", nil)
			if err != nil {
				t.Fatalf(`error when creating request: "%v"`, err)
			}
			resp, err := clientWrapper.SendRequest(req)
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError.Error())
				}
				require.NotNil(t, resp)
			}
		})
	}
}

func TestSendRequestAndUnmarshallJsonResponse(t *testing.T) {
	testCases := []struct {
		name           string
		mockClosure    func(r *retryableHttpClientMock)
		ioReadAllMock  func(r io.Reader) ([]byte, error)
		jsonDecodeMock func(r io.Reader, data any) error
		expectedData   dummyType
		expectedError  error
	}{
		{
			name: "happy path",
			mockClosure: func(r *retryableHttpClientMock) {
				body := `{"key":"value"}`
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte(body))),
				}
				r.Resp = resp
			},
			expectedData: dummyType{
				Key: "value",
			},
		},
		{
			name: "error when decoding response",
			mockClosure: func(r *retryableHttpClientMock) {
				body := `{"key":"value"}`
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte(body))),
				}
				r.Resp = resp
			},
			jsonDecodeMock: func(r io.Reader, data any) error {
				return errors.New("random error")
			},
			expectedError: errors.New("request to http://localhost/some/path failed. " +
				"httpStatus: [ 200 ] responseBody: [  ] error: [ decoding response: random error ]"),
		},
		{
			name: "unsuccessful response with responseBody",
			mockClosure: func(r *retryableHttpClientMock) {
				resp := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"some random error"}`))),
				}
				r.Resp = resp
			},
			expectedError: errors.New(`request to http://localhost/some/path failed. ` +
				`httpStatus: [ 500 ] responseBody: [ {"error":"some random error"} ] error: [ <nil> ]`),
		},
		{
			name: "unsuccessful response with responseBody, error when parsing it",
			mockClosure: func(r *retryableHttpClientMock) {
				resp := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"some random error"}`))),
				}
				r.Resp = resp
			},
			ioReadAllMock: func(r io.Reader) ([]byte, error) {
				return nil, errors.New("random error")
			},
			expectedError: errors.New(`request to http://localhost/some/path failed. ` +
				`httpStatus: [ 500 ] responseBody: [  ] error: [ parsing response: random error ]`),
		},
		{
			name: "unsuccessful response with empty body, with some other error",
			mockClosure: func(r *retryableHttpClientMock) {
				resp := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewReader([]byte(``))),
				}
				r.Resp = resp
				r.Err = errors.New("random error")
			},
			expectedError: errors.New(`request to http://localhost/some/path failed. ` +
				`httpStatus: [ 500 ] responseBody: [  ] error: [ random error ]`),
		},
		{
			name: "without response",
			mockClosure: func(r *retryableHttpClientMock) {
				r.Err = errors.New("random error")
			},
			expectedError: errors.New(`request to http://localhost/some/path failed. ` +
				`httpStatus: [ no status ] responseBody: [  ] error: [ random error ]`),
		},
	}
	originalIoReadAll := ioReadAll
	originalJsonDecode := jsonDecode
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				ioReadAll = originalIoReadAll
				jsonDecode = originalJsonDecode
			}()
			if tc.ioReadAllMock != nil {
				ioReadAll = tc.ioReadAllMock
			}
			if tc.jsonDecodeMock != nil {
				jsonDecode = tc.jsonDecodeMock
			}
			c := New()
			clientWrapper, ok := c.(*client)
			require.True(t, ok)
			retryableHttpClientMock := new(retryableHttpClientMock)
			tc.mockClosure(retryableHttpClientMock)
			clientWrapper.retryableHttpClient = retryableHttpClientMock
			req, err := http.NewRequest(http.MethodPost, "http://localhost/some/path", nil)
			if err != nil {
				t.Fatalf(`error when creating request: "%v"`, err)
			}
			var data dummyType
			resp, err := clientWrapper.SendRequestAndUnmarshallJsonResponse(req, &data)
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError.Error())
				}
				require.NotNil(t, resp)
				require.Equal(t, tc.expectedData, data)
			}
		})
	}
}

type retryableHttpClientMock struct {
	retryableHttpClient
	Resp *http.Response
	Err  error
}

func (r *retryableHttpClientMock) Do(req *retryablehttp.Request) (*http.Response, error) {
	return r.Resp, r.Err
}
