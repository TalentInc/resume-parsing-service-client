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
		rioParseBaseUrl  = "<URL>"
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
