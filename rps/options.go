package rps

import (
	"time"
)

// Option represents a resumeParsingServiceClient option.
type Option func(*resumeParsingServiceClient)

// WithMaxIdleConns defines the maximum number of idle (keep-alive)
// connections across all hosts.
func WithMaxIdleConns(n int) Option {
	return func(c *resumeParsingServiceClient) {
		c.maxIdleConns = n
	}
}

// WithMaxIdleConnsPerHost defines the maximum idle (keep-alive)
// connections to keep per-host.
func WithMaxIdleConnsPerHost(n int) Option {
	return func(c *resumeParsingServiceClient) {
		c.maxIdleConnsPerHost = n
	}
}

// WithMaxConnsPerHost limits the total number of
// connections per host.
func WithMaxConnsPerHost(n int) Option {
	return func(c *resumeParsingServiceClient) {
		c.maxConnsPerHost = n
	}
}

// WithMaxRetries limits the maximum number of retries.
func WithMaxRetries(n int) Option {
	return func(c *resumeParsingServiceClient) {
		c.maxRetries = n
	}
}

// WithRetryWaitMin specifies minimum time to wait before retrying.
func WithRetryWaitMin(d time.Duration) Option {
	return func(c *resumeParsingServiceClient) {
		c.retryWaitMin = d
	}
}

// WithRetryWaitMax specifies maximum time to wait before retrying.
func WithRetryWaitMax(d time.Duration) Option {
	return func(c *resumeParsingServiceClient) {
		c.retryWaitMax = d
	}
}

// WithCheckRetryPolicy specifies the policy for handling retries,
// and is called after each request.
func WithCheckRetryPolicy(checkRetryPolicy checkRetryPolicy) Option {
	return func(c *resumeParsingServiceClient) {
		c.checkRetryPolicy = checkRetryPolicy
	}
}

// WithRequestDumpLogger specifies a function that receives
// the request dump along its body (optionally) for
// logging purposes.
func WithRequestDumpLogger(requestDumpLogger func(dump []byte), dumpRequestBody bool) Option {
	return func(c *resumeParsingServiceClient) {
		c.requestDumpLogger = requestDumpLogger
		c.dumpRequestBody = dumpRequestBody
	}
}
