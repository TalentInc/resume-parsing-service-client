# resume-parsing-service-client

A retryable http client for [resume-parsing-service](https://github.com/resume-io/resume-parsing-service).

It uses [go-retryablehttp](https://github.com/hashicorp/go-retryablehttp) underneath.

The `Resume` model defined in [rps/models.go](./rps/models.go) is the same that is defined in [resume-parsing-service](https://github.com/resume-io/resume-parsing-service).

## highlights

- **Security-first Approach**: Utilizes [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) for continuous vulnerability scanning.
- **Code Quality**: Adheres to strict coding standards with comprehensive linting, using [golangci-lint](https://github.com/golangci/golangci-lint). Linter configuration is defined in `.golangci.yml` file.
- **100% Test Coverage**: Achieves full unit test coverage.

## available options

- `WithMaxIdleConns(n int)` defines the maximum number of idle (keep-alive) connections across all hosts.
- `WithMaxIdleConnsPerHost(n int)` defines the maximum idle (keep-alive) connections to keep per-host.
- `WithMaxConnsPerHost(n int)` limits the total number of connections per host.
- `WithMaxRetries(n int)` limits the maximum number of retries.
- `WithRetryWaitMin(d time.Duration)` specifies minimum time to wait before retrying.
- `WithRetryWaitMax(d time.Duration)` specifies maximum time to wait before retrying.
- `WithCheckRetryPolicy(checkRetryPolicy checkRetryPolicy)` specifies the policy for handling retries, and is called after each request. If none is specified, the request will not be retried by default.
- `WithRequestDumpLogger(requestDumpLogger func(dump []byte), dumpRequestBody bool)` specifies a function that receives the request dump for logging purposes. If `dumpRequestBody` is set to `true`, it will also log the request body.

## usage

### without options

From [example/example.go](./example/example.go):

```
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/TalentInc/resume-parsing-service-client/rps"
)

func main() {
	const (
		sampleResumeFile = "example/sampleResume.docx"
		rioParseToken    = "<TOKEN>"
		rioParseBaseUrl  = "<URL>"
	)
	ctx := context.Background()
	file, err := os.Open(sampleResumeFile)
	if err != nil {
		fmt.Printf(`error when opening file "%s": %v`, sampleResumeFile, err)
		os.Exit(1)
	}
	fileContents, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf(`error when reading file "%s": %v`, sampleResumeFile, err)
		os.Exit(1)
	}
	rpsClient := rps.NewResumeParsingServiceClient(rioParseToken, rioParseBaseUrl)
	resume, err := rpsClient.ParseDocument(ctx, fileContents)
	if err != nil {
		fmt.Printf(`error when uploading file "%s": %v`, sampleResumeFile, err)
		os.Exit(1)
	}
	fmt.Printf("resume: %+v\n", resume)
}

```

### with options

From [example/with-options/example.go](./example/with-options/example.go):

```
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/TalentInc/resume-parsing-service-client/rps"
)

func retryIfInternalServerError(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if resp != nil {
		statusCode := resp.StatusCode
		if statusCode == http.StatusInternalServerError {
			return true, err
		}
	}
	return false, err
}

func requestDumpLogger(dump []byte) {
	fmt.Printf("dump: %v\n", string(dump))
}

func main() {
	const (
		sampleResumeFile = "example/sampleResume.docx"
		rioParseToken    = "<TOKEN>"
		rioParseBaseUrl  = "<URL"
		maxRetries       = 4
		retryWaitMin     = 1 * time.Second
		retryWaitMax     = 5 * time.Second
	)
	ctx := context.Background()
	file, err := os.Open(sampleResumeFile)
	if err != nil {
		fmt.Printf(`error when opening file "%s": %v`, sampleResumeFile, err)
		os.Exit(1)
	}
	fileContents, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf(`error when reading file "%s": %v`, sampleResumeFile, err)
		os.Exit(1)
	}
	rpsClient := rps.NewResumeParsingServiceClient(rioParseToken,
		rioParseBaseUrl,
		rps.WithMaxRetries(maxRetries),
		rps.WithRetryWaitMin(retryWaitMin),
		rps.WithRetryWaitMax(retryWaitMax),
		rps.WithCheckRetryPolicy(retryIfInternalServerError),
		rps.WithRequestDumpLogger(requestDumpLogger, true),
	)
	resume, err := rpsClient.ParseDocument(ctx, fileContents)
	if err != nil {
		fmt.Printf(`error when uploading file "%s": %v`, sampleResumeFile, err)
		os.Exit(1)
	}
	fmt.Printf("resume: %+v\n", resume)
}

```

In this example:

- `retryIfInternalServerError` is our policy to retry the request if the response code is HTTP 500
- `requestDumpLogger` prints out the request dump
- `maxRetries` is set to `4`, which means that the request will be retried 4 times after the first attempt
- `retryWaitMin` and `retryWaitMax` are in place to make sure that we'll wait a time range before attempting the request

For example, let's suppose that [resume-parsing-service](https://github.com/resume-io/resume-parsing-service) returns `HTTP 500` for the first request and `HTTP 200` for the second request. Sample output:

```
$ go run example/with-options/example.go 

dump: POST /api/parse HTTP/1.1
Host: localhost:3333
User-Agent: Go-http-client/1.1
Content-Length: 49810
Content-Type: application/json
Token: <TOKEN>
Accept-Encoding: gzip

{"base64_data":"UEsDBBQABgA...}

2024/03/04 21:34:33 [DEBUG] POST <RIO_RESUME_PARSING_URL>/api/parse
2024/03/04 21:34:33 [DEBUG] POST <RIO_RESUME_PARSING_URL>/api/parse (status: 500): retrying in 1s (4 left)
resume: &{FirstName:Morgana MiddleName: LastName:Favero Summary:I am a Neuroscientist...}
```

Both request body and `resume` outputs were omitted for brevity.

The request body is being printed because we passed `true` to `rps.WithRequestDumpLogger(requestDumpLogger, true)`.

If we don't want it to be printed (`rps.WithRequestDumpLogger(requestDumpLogger, false)`), output would be

```
$ go run example/with-options/example.go 

dump: POST /api/parse HTTP/1.1
Host: localhost:3333
User-Agent: Go-http-client/1.1
Content-Length: 49810
Content-Type: application/json
Token: <TOKEN>
Accept-Encoding: gzip


2024/03/04 21:34:33 [DEBUG] POST <RIO_RESUME_PARSING_URL>/api/parse
2024/03/04 21:34:33 [DEBUG] POST <RIO_RESUME_PARSING_URL>/api/parse (status: 500): retrying in 1s (4 left)
resume: &{FirstName:Morgana MiddleName: LastName:Favero Summary:I am a Neuroscientist...}
```

## running unit tests

```
make test
```

## running unit tests with coverage report in html

```
make coverage
```

## running linter

```
make lint
```

## check for vulnerabilities

```
make vul-check
```

## available Makefile targets

```
make help
```