package httpclient

import (
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// Option represents a client option.
type Option func(*client)

// WithMaxIdleConns defines the maximum number of idle (keep-alive)
// connections across all hosts.
func WithMaxIdleConns(n int) Option {
	return func(c *client) {
		c.maxIdleConns = n
	}
}

// WithMaxIdleConnsPerHost defines the maximum idle (keep-alive)
// connections to keep per-host.
func WithMaxIdleConnsPerHost(n int) Option {
	return func(c *client) {
		c.maxIdleConnsPerHost = n
	}
}

// WithMaxConnsPerHost limits the total number of
// connections per host.
func WithMaxConnsPerHost(n int) Option {
	return func(c *client) {
		c.maxConnsPerHost = n
	}
}

// WithMaxRetries limits the maximum number of retries.
func WithMaxRetries(n int) Option {
	return func(c *client) {
		c.maxRetries = n
	}
}

// WithRetryWaitMin specifies minimum time to wait before retrying.
func WithRetryWaitMin(n time.Duration) Option {
	return func(c *client) {
		c.retryWaitMin = n
	}
}

// WithRetryWaitMax specifies maximum time to wait before retrying.
func WithRetryWaitMax(n time.Duration) Option {
	return func(c *client) {
		c.retryWaitMax = n
	}
}

// WithCheckRetryPolicy specifies the policy for handling retries,
// and is called after each request.
func WithCheckRetryPolicy(checkRetryPolicy retryablehttp.CheckRetry) Option {
	return func(c *client) {
		c.checkRetryPolicy = checkRetryPolicy
	}
}

// WithRequestDumpLogger specifies a function that receives
// the request dump along its body (optionally) for
// logging purposes.
func WithRequestDumpLogger(requestDumpLogger func(dump []byte), dumpRequestBody bool) Option {
	return func(c *client) {
		c.requestDumpLogger = requestDumpLogger
		c.dumpRequestBody = dumpRequestBody
	}
}
