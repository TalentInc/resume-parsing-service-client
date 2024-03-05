package rps

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/TalentInc/resume-parsing-service-client/httpclient"
	"github.com/stretchr/testify/require"
)

func TestNewResumeParsingServiceClient(t *testing.T) {
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
			checkRetryPolicy:            true,
			checkRequestDumpLogger:      true,
			expectedMaxIdleConns:        1,
			expectedMaxIdleConnsPerHost: 1,
			expectedMaxConnsPerHost:     1,
			expectedMaxRetries:          1,
			expectedRetryWaitMin:        1 * time.Second,
			expectedRetryWaitMax:        1 * time.Second,
			expectedDumpRequestBody:     true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewResumeParsingServiceClient("TOKEN", "URL", tc.options...)
			clientWrapper, ok := client.(*resumeParsingServiceClient)
			require.True(t, ok)
			require.Equal(t, tc.expectedMaxIdleConns, clientWrapper.maxIdleConns)
			require.Equal(t, tc.expectedMaxIdleConnsPerHost, clientWrapper.maxIdleConnsPerHost)
			require.Equal(t, tc.expectedMaxConnsPerHost, clientWrapper.maxConnsPerHost)
			require.Equal(t, tc.expectedMaxRetries, clientWrapper.maxRetries)
			require.Equal(t, tc.expectedRetryWaitMin, clientWrapper.retryWaitMin)
			require.Equal(t, tc.expectedRetryWaitMax, clientWrapper.retryWaitMax)
			require.Equal(t, tc.expectedDumpRequestBody, clientWrapper.dumpRequestBody)
			if tc.checkRequestDumpLogger {
				require.NotNil(t, clientWrapper.requestDumpLogger)
			}
			if tc.checkRetryPolicy {
				require.NotNil(t, clientWrapper.checkRetryPolicy)
			}
		})
	}
}

func TestParseDocument(t *testing.T) {
	testCases := []struct {
		name                      string
		mockJsonMarshal           func(v any) ([]byte, error)
		mockNewRequestWithContext func(ctx context.Context, method string, url string, body io.Reader) (*http.Request, error)
		newHttpClientMock         func(options ...httpclient.Option) httpclient.Client
		expectedOutput            *Resume
		expectedError             error
	}{
		{
			name: "happy path",
			mockJsonMarshal: func(v any) ([]byte, error) {
				return []byte{}, nil
			},
			mockNewRequestWithContext: func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
				r := new(http.Request)
				r.Header = make(http.Header)
				return r, nil
			},
			newHttpClientMock: func(options ...httpclient.Option) httpclient.Client {
				h := new(httpClientMock)
				body := `{"first_name":"Morgana","middle_name":"","last_name":"Favero","summary":"I am a Neuroscientist...","pdf":"pdf location","location":{"formatted":"3850 Woodhaven Road, Philadelphia, PA, USA","street":"Woodhaven Road","city":"Philadelphia","state":"Pennsylvania","country":"United States","countryCode":"US"},"emails":["favero.morgana@gmail.com"],"profession":"Postdoctoral Researcher","positions":[{"title":"Postdoctoral Researcher","title_normalized":"Postdoctoral Researcher","organization":"The Children's Hospital of Philadelphia","start_date":"2015-11-01T00:00:00Z","end_date":"2024-03-03T00:00:00Z","description":"","location":{"formatted":"Philadelphia, PA, USA","street":"","city":"Philadelphia","state":"Pennsylvania","country":"United States","countryCode":"US"},"management_level":"Low"},{"title":"Assistant Professor","title_normalized":"Assistant Professor","organization":"University of Verona","start_date":"2013-03-01T00:00:00Z","end_date":"2015-10-01T00:00:00Z","description":"description","location":{"formatted":"Verona, VR, Italy","street":"","city":"Verona","state":"Verona","country":"Italy","countryCode":"IT"},"management_level":"Low"},{"title":"Postdoctoral Researcher","title_normalized":"Postdoctoral Researcher","organization":"Drexel University College of Medicine","start_date":"2009-01-01T00:00:00Z","end_date":"2013-02-01T00:00:00Z","description":"description","location":{"formatted":"Philadelphia, PA, USA","street":"","city":"Philadelphia","state":"Pennsylvania","country":"United States","countryCode":"US"},"management_level":"Low"}],"educations":[{"organization":"University of Verona in","degree":"Doctor of Philosophy","start_date":"2002-01-01T00:00:00Z","end_date":"2008-01-01T00:00:00Z","location":{"formatted":"Verona, VR, Italy","street":"","city":"Verona","state":"Verona","country":"Italy","countryCode":"IT"},"education_level":"doctoral"},{"organization":"University of Padova in","degree":"MD, Medicine and Surgery,","start_date":"1995-01-01T00:00:00Z","end_date":"2002-01-01T00:00:00Z","location":{"formatted":"Padova, PD, Italy","street":"","city":"Padova","state":"Padova","country":"Italy","countryCode":"IT"},"education_level":""}],"social_urls":[],"phone_numbers":[{"country_code":"+1","country_name":"US","national_number":"(267) 721-0053"}],"languages":["French","English","Italian","Spanish"],"detected_language":"en","skills":[{"name":"Collaboration","num_months":31},{"name":"Editing","num_months":0},{"name":"Research","num_months":80},{"name":"EndNote","num_months":0},{"name":"Physiology","num_months":31},{"name":"Reference Management","num_months":0},{"name":"Scopus","num_months":0},{"name":"Detail Oriented","num_months":0},{"name":"Authorization (Computing)","num_months":31},{"name":"Strategic Planning","num_months":0},{"name":"Microsoft Excel","num_months":0},{"name":"Planning","num_months":0},{"name":"Journals","num_months":0},{"name":"Physical Therapy","num_months":31},{"name":"Pharmacology","num_months":0},{"name":"Reference Management Software","num_months":0},{"name":"Synaptic","num_months":0},{"name":"Teamwork","num_months":0},{"name":"Research Papers","num_months":0},{"name":"Microsoft PowerPoint","num_months":0},{"name":"Optogenetics","num_months":0},{"name":"Teaching","num_months":0},{"name":"Electrophysiology","num_months":0},{"name":"Writing","num_months":0},{"name":"Communications","num_months":0},{"name":"Time Management","num_months":0},{"name":"Critical Thinking","num_months":0},{"name":"Adobe Acrobat","num_months":0},{"name":"Creative Thinking","num_months":0},{"name":"Pubmed","num_months":0},{"name":"Pharmaceuticals","num_months":0},{"name":"Neuroscience","num_months":0},{"name":"Presentations","num_months":0},{"name":"Management","num_months":0},{"name":"Microsoft Word","num_months":0}],"raw_text":"MORGANA FAVERO, MD, PhD 3850 Woodhaven Road, Philadelphia, PA 19154..."}`
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte(body))),
				}
				h.Resp = resp
				return h
			},
			expectedOutput: buildExpectedOutput(),
		},
		{
			name: "error when marshalling",
			mockJsonMarshal: func(v any) ([]byte, error) {
				return []byte{}, errors.New("marshalling error")
			},
			mockNewRequestWithContext: func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
				r := new(http.Request)
				r.Header = make(http.Header)
				return r, nil
			},
			newHttpClientMock: func(options ...httpclient.Option) httpclient.Client {
				return &httpClientMock{}
			},
			expectedError: errors.New("marshalling parse document request: marshalling error"),
		},
		{
			name: "error when creating request",
			mockJsonMarshal: func(v any) ([]byte, error) {
				return []byte{}, nil
			},
			mockNewRequestWithContext: func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
				return nil, errors.New("create request error")
			},
			newHttpClientMock: func(options ...httpclient.Option) httpclient.Client {
				return &httpClientMock{}
			},
			expectedError: errors.New("creating request: create request error"),
		},
		{
			name: "error when performing request",
			mockJsonMarshal: func(v any) ([]byte, error) {
				return []byte{}, nil
			},
			mockNewRequestWithContext: func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
				r := new(http.Request)
				r.Header = make(http.Header)
				return r, nil
			},
			newHttpClientMock: func(options ...httpclient.Option) httpclient.Client {
				return &httpClientMock{
					Err: errors.New("random error"),
				}
			},
			expectedError: errors.New("performing request: random error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonMarshal = tc.mockJsonMarshal
			newRequestWithContext = tc.mockNewRequestWithContext
			newHttpClient = tc.newHttpClientMock
			rpsClient := NewResumeParsingServiceClient("", "")
			output, err := rpsClient.ParseDocument(context.TODO(), []byte{})
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError.Error())
				}
				require.Equal(t, tc.expectedOutput, output)
			}
		})
	}
}

func output() *Resume {
	const layout = "2006-01-02 15:04:05 -0700 MST"

	startDatePosition1, _ := time.Parse(layout, "2015-11-01 00:00:00 +0000 UTC")
	endDatePosition1, _ := time.Parse(layout, "2024-03-03 00:00:00 +0000 UTC")
	startDatePosition2, _ := time.Parse(layout, "2013-03-01 00:00:00 +0000 UTC")
	endDatePosition2, _ := time.Parse(layout, "2015-10-01 00:00:00 +0000 UTC")
	startDatePosition3, _ := time.Parse(layout, "2009-01-01 00:00:00 +0000 UTC")
	endDatePosition3, _ := time.Parse(layout, "2013-02-01 00:00:00 +0000 UTC")

	startDateEducation1, _ := time.Parse(layout, "2002-01-01 00:00:00 +0000 UTC")
	endDateEducation1, _ := time.Parse(layout, "2008-01-01 00:00:00 +0000 UTC")
	startDateEducation2, _ := time.Parse(layout, "1995-01-01 00:00:00 +0000 UTC")
	endDateEducation2, _ := time.Parse(layout, "2002-01-01 00:00:00 +0000 UTC")

	resume := &Resume{
		FirstName:  "Morgana",
		MiddleName: "",
		LastName:   "Favero",
		Summary:    "I am a Neuroscientist...",
		Pdf:        "pdf location",
		Location: Location{
			Formatted:   "3850 Woodhaven Road, Philadelphia, PA, USA",
			Street:      "Woodhaven Road",
			City:        "Philadelphia",
			State:       "Pennsylvania",
			Country:     "United States",
			CountryCode: "US",
		},
		Emails:     []string{"favero.morgana@gmail.com"},
		Profession: "Postdoctoral Researcher",
		Positions: []Position{
			{
				Title:           "Postdoctoral Researcher",
				TitleNormalized: "Postdoctoral Researcher",
				Organization:    "The Children's Hospital of Philadelphia",
				StartDate:       &startDatePosition1,
				EndDate:         &endDatePosition1,
				Description:     "",
				Location: Location{
					Formatted:   "Philadelphia, PA, USA",
					City:        "Philadelphia",
					State:       "Pennsylvania",
					Country:     "United States",
					CountryCode: "US",
				},
				ManagementLevel: "Low",
			},
			{
				Title:           "Assistant Professor",
				TitleNormalized: "Assistant Professor",
				Organization:    "University of Verona",
				StartDate:       &startDatePosition2,
				EndDate:         &endDatePosition2,
				Description:     "description",
				Location: Location{
					Formatted:   "Verona, VR, Italy",
					City:        "Verona",
					State:       "Verona",
					Country:     "Italy",
					CountryCode: "IT",
				},
				ManagementLevel: "Low",
			},
			{
				Title:           "Postdoctoral Researcher",
				TitleNormalized: "Postdoctoral Researcher",
				Organization:    "Drexel University College of Medicine",
				StartDate:       &startDatePosition3,
				EndDate:         &endDatePosition3,
				Description:     "description",
				Location: Location{
					Formatted:   "Philadelphia, PA, USA",
					City:        "Philadelphia",
					State:       "Pennsylvania",
					Country:     "United States",
					CountryCode: "US",
				},
				ManagementLevel: "Low",
			},
		},
		Educations: []Education{
			{
				Organization: "University of Verona",
				Degree:       "Doctor of Philosophy",
				StartDate:    &startDateEducation1,
				EndDate:      &endDateEducation1,
				Location: Location{
					Formatted:   "Verona, VR, Italy",
					City:        "Verona",
					State:       "Verona",
					Country:     "Italy",
					CountryCode: "IT",
				},
				EducationLevel: "doctoral",
			},
			{
				Organization: "University of Padova",
				Degree:       "MD, Medicine and Surgery",
				StartDate:    &startDateEducation2,
				EndDate:      &endDateEducation2,
				Location: Location{
					Formatted:   "Padova, PD, Italy",
					City:        "Padova",
					State:       "Padova",
					Country:     "Italy",
					CountryCode: "IT",
				},
				EducationLevel: "",
			},
		},
		SocialUrls:       []SocialUrl{},
		PhoneNumbers:     []PhoneNumber{{CountryCode: "+1", CountryName: "US", NationalNumber: "(267) 721-0053"}},
		Languages:        []string{"French", "Italian", "Spanish", "English"},
		DetectedLanguage: "en",
		Skills: []Skill{
			{Name: "Synaptic", NumMonths: 0},
			{Name: "Microsoft Word", NumMonths: 0},
			{Name: "Detail Oriented", NumMonths: 0},
			{Name: "Presentations", NumMonths: 0},
			{Name: "Teaching", NumMonths: 0},
			{Name: "Writing", NumMonths: 0},
			{Name: "Planning", NumMonths: 0},
			{Name: "Electrophysiology", NumMonths: 0},
			{Name: "Neuroscience", NumMonths: 0},
			{Name: "Critical Thinking", NumMonths: 0},
			{Name: "Adobe Acrobat", NumMonths: 0},
			{Name: "Management", NumMonths: 0},
			{Name: "Editing", NumMonths: 0},
			{Name: "Physical Therapy", NumMonths: 31},
			{Name: "Reference Management", NumMonths: 0},
			{Name: "Pubmed", NumMonths: 0},
			{Name: "Optogenetics", NumMonths: 0},
			{Name: "Microsoft Excel", NumMonths: 0},
			{Name: "Teamwork", NumMonths: 0},
			{Name: "Time Management", NumMonths: 0},
			{Name: "Pharmaceuticals", NumMonths: 0},
			{Name: "Creative Thinking", NumMonths: 0},
			{Name: "Communications", NumMonths: 0},
			{Name: "Microsoft PowerPoint", NumMonths: 0},
			{Name: "Pharmacology", NumMonths: 0},
			{Name: "Collaboration", NumMonths: 31},
			{Name: "Scopus", NumMonths: 0},
			{Name: "Authorization (Computing)", NumMonths: 31},
			{Name: "Research", NumMonths: 80},
			{Name: "EndNote", NumMonths: 0},
			{Name: "Strategic Planning", NumMonths: 0},
			{Name: "Physiology", NumMonths: 31},
			{Name: "Reference Management Software", NumMonths: 0},
			{Name: "Research Papers", NumMonths: 0},
			{Name: "Journals", NumMonths: 0},
		},
		RawText: "MORGANA FAVERO, MD, PhD 3850 Woodhaven Road, Philadelphia, PA 19154...",
	}
	return resume
}

func buildExpectedOutput() *Resume {
	const layout = "2006-01-02 15:04:05 -0700 MST"

	startDatePosition1, _ := time.Parse(layout, "2015-11-01 00:00:00 +0000 UTC")
	endDatePosition1, _ := time.Parse(layout, "2024-03-03 00:00:00 +0000 UTC")
	startDatePosition2, _ := time.Parse(layout, "2013-03-01 00:00:00 +0000 UTC")
	endDatePosition2, _ := time.Parse(layout, "2015-10-01 00:00:00 +0000 UTC")
	startDatePosition3, _ := time.Parse(layout, "2009-01-01 00:00:00 +0000 UTC")
	endDatePosition3, _ := time.Parse(layout, "2013-02-01 00:00:00 +0000 UTC")

	startDateEducation1, _ := time.Parse(layout, "2002-01-01 00:00:00 +0000 UTC")
	endDateEducation1, _ := time.Parse(layout, "2008-01-01 00:00:00 +0000 UTC")
	startDateEducation2, _ := time.Parse(layout, "1995-01-01 00:00:00 +0000 UTC")
	endDateEducation2, _ := time.Parse(layout, "2002-01-01 00:00:00 +0000 UTC")

	resume := &Resume{
		FirstName:  "Morgana",
		MiddleName: "",
		LastName:   "Favero",
		Summary:    "I am a Neuroscientist...",
		Pdf:        "pdf location",
		Location: Location{
			Formatted:   "3850 Woodhaven Road, Philadelphia, PA, USA",
			Street:      "Woodhaven Road",
			City:        "Philadelphia",
			State:       "Pennsylvania",
			Country:     "United States",
			CountryCode: "US",
		},
		Emails:     []string{"favero.morgana@gmail.com"},
		Profession: "Postdoctoral Researcher",
		Positions: []Position{
			{
				Title:           "Postdoctoral Researcher",
				TitleNormalized: "Postdoctoral Researcher",
				Organization:    "The Children's Hospital of Philadelphia",
				StartDate:       &startDatePosition1,
				EndDate:         &endDatePosition1,
				Description:     "",
				Location: Location{
					Formatted:   "Philadelphia, PA, USA",
					City:        "Philadelphia",
					State:       "Pennsylvania",
					Country:     "United States",
					CountryCode: "US",
				},
				ManagementLevel: "Low",
			},
			{
				Title:           "Assistant Professor",
				TitleNormalized: "Assistant Professor",
				Organization:    "University of Verona",
				StartDate:       &startDatePosition2,
				EndDate:         &endDatePosition2,
				Description:     "description",
				Location: Location{
					Formatted:   "Verona, VR, Italy",
					City:        "Verona",
					State:       "Verona",
					Country:     "Italy",
					CountryCode: "IT",
				},
				ManagementLevel: "Low",
			},
			{
				Title:           "Postdoctoral Researcher",
				TitleNormalized: "Postdoctoral Researcher",
				Organization:    "Drexel University College of Medicine",
				StartDate:       &startDatePosition3,
				EndDate:         &endDatePosition3,
				Description:     "description",
				Location: Location{
					Formatted:   "Philadelphia, PA, USA",
					City:        "Philadelphia",
					State:       "Pennsylvania",
					Country:     "United States",
					CountryCode: "US",
				},
				ManagementLevel: "Low",
			},
		},
		Educations: []Education{
			{
				Organization: "University of Verona",
				Degree:       "Doctor of Philosophy",
				StartDate:    &startDateEducation1,
				EndDate:      &endDateEducation1,
				Location: Location{
					Formatted:   "Verona, VR, Italy",
					City:        "Verona",
					State:       "Verona",
					Country:     "Italy",
					CountryCode: "IT",
				},
				EducationLevel: "doctoral",
			},
			{
				Organization: "University of Padova",
				Degree:       "MD, Medicine and Surgery",
				StartDate:    &startDateEducation2,
				EndDate:      &endDateEducation2,
				Location: Location{
					Formatted:   "Padova, PD, Italy",
					City:        "Padova",
					State:       "Padova",
					Country:     "Italy",
					CountryCode: "IT",
				},
				EducationLevel: "",
			},
		},
		SocialUrls:       []SocialUrl{},
		PhoneNumbers:     []PhoneNumber{{CountryCode: "+1", CountryName: "US", NationalNumber: "(267) 721-0053"}},
		Languages:        []string{"French", "Italian", "Spanish", "English"},
		DetectedLanguage: "en",
		Skills: []Skill{
			{Name: "Synaptic", NumMonths: 0},
			{Name: "Microsoft Word", NumMonths: 0},
			{Name: "Detail Oriented", NumMonths: 0},
			{Name: "Presentations", NumMonths: 0},
			{Name: "Teaching", NumMonths: 0},
			{Name: "Writing", NumMonths: 0},
			{Name: "Planning", NumMonths: 0},
			{Name: "Electrophysiology", NumMonths: 0},
			{Name: "Neuroscience", NumMonths: 0},
			{Name: "Critical Thinking", NumMonths: 0},
			{Name: "Adobe Acrobat", NumMonths: 0},
			{Name: "Management", NumMonths: 0},
			{Name: "Editing", NumMonths: 0},
			{Name: "Physical Therapy", NumMonths: 31},
			{Name: "Reference Management", NumMonths: 0},
			{Name: "Pubmed", NumMonths: 0},
			{Name: "Optogenetics", NumMonths: 0},
			{Name: "Microsoft Excel", NumMonths: 0},
			{Name: "Teamwork", NumMonths: 0},
			{Name: "Time Management", NumMonths: 0},
			{Name: "Pharmaceuticals", NumMonths: 0},
			{Name: "Creative Thinking", NumMonths: 0},
			{Name: "Communications", NumMonths: 0},
			{Name: "Microsoft PowerPoint", NumMonths: 0},
			{Name: "Pharmacology", NumMonths: 0},
			{Name: "Collaboration", NumMonths: 31},
			{Name: "Scopus", NumMonths: 0},
			{Name: "Authorization (Computing)", NumMonths: 31},
			{Name: "Research", NumMonths: 80},
			{Name: "EndNote", NumMonths: 0},
			{Name: "Strategic Planning", NumMonths: 0},
			{Name: "Physiology", NumMonths: 31},
			{Name: "Reference Management Software", NumMonths: 0},
			{Name: "Research Papers", NumMonths: 0},
			{Name: "Journals", NumMonths: 0},
		},
		RawText: "MORGANA FAVERO, MD, PhD 3850 Woodhaven Road, Philadelphia, PA 19154...",
	}
	return resume
}

type httpClientMock struct {
	httpclient.Client
	Resp *http.Response
	Err  error
}

func (m *httpClientMock) SendRequestAndUnmarshallJsonResponse(req *http.Request, v any) (*http.Response, error) {
	r, _ := v.(*Resume)
	*r = *output()
	return m.Resp, m.Err
}
