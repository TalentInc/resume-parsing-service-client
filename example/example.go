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
