package rps

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/TalentInc/resume-parsing-service-client/httpclient"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
)

// For ease of unit testing.
// Declaring these functions as global variables
// makes it easy to mock them.
var (
	jsonMarshal           = json.Marshal
	newRequestWithContext = http.NewRequestWithContext
	newHttpClient         = httpclient.New
)

type checkRetryPolicy retryablehttp.CheckRetry

// ResumeParsingServiceClient defines the interface for a client capable of sending
// resume documents to the Resume Parsing Service and receiving parsed data in response.
type ResumeParsingServiceClient interface {
	// ParseDocument sends a resume document for parsing and returns the parsed data.
	ParseDocument(ctx context.Context, fileContents []byte) (*Resume, error)
}

// resumeParsingServiceClient implements ResumeParsingServiceClient interface.
type resumeParsingServiceClient struct {
	rioParseToken   string
	rioParseBaseUrl string

	checkRetryPolicy    checkRetryPolicy
	maxIdleConns        int
	maxIdleConnsPerHost int
	maxConnsPerHost     int
	maxRetries          int
	retryWaitMin        time.Duration
	retryWaitMax        time.Duration
	requestDumpLogger   func(dump []byte)
	dumpRequestBody     bool

	httpClient httpclient.Client
}

// newResumeParsingServiceClient applies the options and returns a
// new instance of a client for the Resume Parsing Service.
func newResumeParsingServiceClient(options []Option) *resumeParsingServiceClient {
	client := new(resumeParsingServiceClient)
	for _, option := range options {
		option(client)
	}
	return client
}

// NewResumeParsingServiceClient initializes a new instance of a client for the Resume Parsing Service.
func NewResumeParsingServiceClient(rioParseToken, rioParseBaseUrl string, options ...Option) ResumeParsingServiceClient {
	client := newResumeParsingServiceClient(options)
	client.rioParseToken = rioParseToken
	client.rioParseBaseUrl = rioParseBaseUrl
	httpClient := newHttpClient(
		httpclient.WithMaxIdleConns(client.maxIdleConns),
		httpclient.WithMaxIdleConnsPerHost(client.maxIdleConnsPerHost),
		httpclient.WithMaxConnsPerHost(client.maxConnsPerHost),
		httpclient.WithMaxRetries(client.maxRetries),
		httpclient.WithRetryWaitMin(client.retryWaitMin),
		httpclient.WithRetryWaitMax(client.retryWaitMax),
		httpclient.WithCheckRetryPolicy(retryablehttp.CheckRetry(client.checkRetryPolicy)),
		httpclient.WithRequestDumpLogger(client.requestDumpLogger, client.dumpRequestBody),
	)
	client.httpClient = httpClient
	return client
}

func (r *resumeParsingServiceClient) ParseDocument(ctx context.Context, fileContents []byte) (*Resume, error) {
	url := fmt.Sprintf("%s/%s", r.rioParseBaseUrl, "api/parse")
	encodedFileContents := base64.StdEncoding.EncodeToString(fileContents)
	parseDocumentRequest := &parseDocumentRequest{
		Base64Data: encodedFileContents,
	}
	j, err := jsonMarshal(parseDocumentRequest)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling parse document request")
	}
	req, err := newRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(j))
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", r.rioParseToken)
	var resume Resume
	resp, err := r.httpClient.SendRequestAndUnmarshallJsonResponse(req, &resume)
	if err != nil {
		return nil, errors.Wrap(err, "performing request")
	}
	defer resp.Body.Close()
	return &resume, nil
}
