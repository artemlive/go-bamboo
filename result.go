package bamboo

import (
	"fmt"
	"net/http"
)

// ResultService handles communication with build results
type ResultService service

// Result represents all the information associated with a build result
type Result struct {
	ChangeSet              `json:"changes"`
	ID                     int             `json:"id"`
	PlanName               string          `json:"planName"`
	ProjectName            string          `json:"projectName"`
	BuildResultKey         string          `json:"buildResultKey"`
	LifeCycleState         string          `json:"lifeCycleState"`
	BuildStartedTime       string          `json:"buildStartedTime"`
	BuildCompletedTime     string          `json:"buildCompletedTime"`
	BuildDurationInSeconds int             `json:"buildDurationInSeconds"`
	VcsRevisionKey         string          `json:"vcsRevisionKey"`
	BuildTestSummary       string          `json:"buildTestSummary"`
	SuccessfulTestCount    int             `json:"successfulTestCount"`
	FailedTestCount        int             `json:"failedTestCount"`
	QuarantinedTestCount   int             `json:"quarantinedTestCount"`
	SkippedTestCount       int             `json:"skippedTestCount"`
	Finished               bool            `json:"finished"`
	Successful             bool            `json:"successful"`
	BuildReason            string          `json:"buildReason"`
	ReasonSummary          string          `json:"reasonSummary"`
	Key                    string          `json:"key"`
	State                  string          `json:"state"`
	BuildState             string          `json:"buildState"`
	Number                 int             `json:"number"`
	Variables              VariableContext `json:"variableContext"`
	BuildNumber            int             `json:"buildNumber"`
}

// represents variables list from api response
type VariableContext struct {
	Size       int        `json:"size"`
	MaxResults int        `json:"max-results"`
	StartIndex int        `json:"start-index"`
	Variables  []Variable `json:"variable"`
}

type Variable struct {
	Key          string `json:"key"`
	Value        string `json:"value"`
	VariableType string `json:"variableType"`
	IsPassword   bool   `json:"isPassword"`
}

// ChangeSet represents a collection of type Change
type ChangeSet struct {
	Set []Change `json:"change"`
}

// Change represents the author and commit hash of a source code change
type Change struct {
	Author      string `json:"author"`
	ChangeSetID string `json:"changesetId"`
}

// LatestResult returns the latest result information for the given plan key
func (r *ResultService) LatestResult(key string) (*Result, *http.Response, error) {
	result, resp, err := r.NumberedResult(key + "-latest")
	return result, resp, err
}

// NumberedResult returns the result information for the given plan key which includes the build number of the desired result
func (r *ResultService) NumberedResult(key string) (*Result, *http.Response, error) {
	request, err := r.client.NewRequest(http.MethodGet, numberedResultURL(key), nil)
	if err != nil {
		return nil, nil, err
	}

	result := Result{}
	response, err := r.client.Do(request, &result)
	if err != nil {
		return nil, response, err
	}

	if response.StatusCode != 200 {
		return nil, response, &simpleError{fmt.Sprintf("API returned unexpected status code %d", response.StatusCode)}
	}

	return &result, response, err
}
