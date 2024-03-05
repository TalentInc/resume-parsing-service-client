package httpclient

import (
	"net/http"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

// retryableHttpClient defines an interface for an HTTP client
// with retry capabilities.
type retryableHttpClient interface {
	// SetRetryMax sets the maximum number of retries for a request.
	SetRetryMax(maxRetries int)

	// SetRetryWaitMin sets the minimum wait time between retries.
	SetRetryWaitMin(minWait time.Duration)

	// SetRetryWaitMax sets the maximum wait time between retries.
	SetRetryWaitMax(maxWait time.Duration)

	// SetCheckRetry specifies a custom retry policy function.
	SetCheckRetry(checkRetry retryablehttp.CheckRetry)

	// Do sends an HTTP request and returns an HTTP response, applying retry logic as configured.
	Do(req *retryablehttp.Request) (*http.Response, error)
}

// retryableHttpClientWrapper implements the retryableHttpClient interface,
// wrapping the HashiCorp retryablehttp.Client (github.com/hashicorp/go-retryablehttp)
// to provide enhanced functionality and easier configuration of retry policies.
type retryableHttpClientWrapper struct {
	rhc *retryablehttp.Client
}

// This construct aids in mocking by allowing users to implement only
// the functions they need for tests or other use cases.
var _ retryableHttpClient = (*retryableHttpClientWrapper)(nil)

func (r *retryableHttpClientWrapper) SetRetryMax(maxRetries int) {
	r.rhc.RetryMax = maxRetries
}

func (r *retryableHttpClientWrapper) SetRetryWaitMin(minWait time.Duration) {
	r.rhc.RetryWaitMin = minWait
}

func (r *retryableHttpClientWrapper) SetRetryWaitMax(maxWait time.Duration) {
	r.rhc.RetryWaitMax = maxWait
}

func (r *retryableHttpClientWrapper) SetCheckRetry(checkRetry retryablehttp.CheckRetry) {
	r.rhc.CheckRetry = checkRetry
}

func (r *retryableHttpClientWrapper) Do(req *retryablehttp.Request) (*http.Response, error) {
	return r.rhc.Do(req)
}
