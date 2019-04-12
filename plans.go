package bamboo

import (
	"fmt"
	"net/http"
	"strconv"
)

// PlanService handles communication with the plan related methods
type PlanService service

// PlanCreateBranchOptions specifies the optional parameters
// for the CreatePlanBranch method
type PlanCreateBranchOptions struct {
	VCSBranch string
}

// PlanResponse encapsultes a response from the plan service
type PlanResponse struct {
	*ResourceMetadata
	Plans *Plans `json:"plans"`
}

// Plans is a collection of Plan objects
type Plans struct {
	*CollectionMetadata
	PlanList []*Plan `json:"plan"`
}

// Plan is the definition of a single plan
type Plan struct {
	ShortName string   `json:"shortName,omitempty"`
	ShortKey  string   `json:"shortKey,omitempty"`
	Type      string   `json:"type,omitempty"`
	Enabled   bool     `json:"enabled,omitempty"`
	Link      *Link    `json:"link,omitempty"`
	Key       string   `json:"key,omitempty"`
	Name      string   `json:"name,omitempty"`
	PlanKey   *PlanKey `json:"planKey,omitempty"`
}

type PlanInfo struct {
	Expand      string `json:"expand"`
	ProjectKey  string `json:"projectKey"`
	ProjectName string `json:"projectName"`
	Project     struct {
		Key         string `json:"key"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Link        struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"link"`
	} `json:"project"`
	ShortName string `json:"shortName"`
	BuildName string `json:"buildName"`
	ShortKey  string `json:"shortKey"`
	Type      string `json:"type"`
	Enabled   bool   `json:"enabled"`
	Link      struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"link"`
	IsFavourite               bool    `json:"isFavourite"`
	IsActive                  bool    `json:"isActive"`
	IsBuilding                bool    `json:"isBuilding"`
	AverageBuildTimeInSeconds float64 `json:"averageBuildTimeInSeconds"`
	Actions                   struct {
		Size       int `json:"size"`
		StartIndex int `json:"start-index"`
		MaxResult  int `json:"max-result"`
	} `json:"actions"`
	Stages struct {
		Size       int `json:"size"`
		StartIndex int `json:"start-index"`
		MaxResult  int `json:"max-result"`
	} `json:"stages"`
	Branches struct {
		Size       int `json:"size"`
		StartIndex int `json:"start-index"`
		MaxResult  int `json:"max-result"`
	} `json:"branches"`
	Variables VariableContext `json:"variableContext"`
	PlanKey struct {
		Key string `json:"key"`
	} `json:"planKey"`
	Key     string `json:"key"`
	Name    string `json:"name"`
}

// PlanKey holds the plan-key for a plan
type PlanKey struct {
	Key string `json:"key,omitempty"`
}

//http://bamboo.epom.com/rest/api/latest/plan/DEV-TEST?expand=variableContext

func (p *PlanService) PlanVariables(planKey string) (VariableContext, *http.Response, error) {
	u := fmt.Sprintf("plan/%s%s", planKey, variablesListURL())
	request, err := p.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return VariableContext{}, nil, err
	}
	planVars := PlanInfo{}
	response, err := p.client.Do(request, &planVars)
	if err != nil {
		return planVars.Variables, response, err
	}
	return planVars.Variables, response, nil
}

// CreatePlanBranch will create a plan branch with the given branch name for the specified build
func (p *PlanService) CreatePlanBranch(planKey, branchName string, options *PlanCreateBranchOptions) (bool, *http.Response, error) {
	var u string
	if !emptyStrings(planKey, branchName) {
		u = fmt.Sprintf("plan/%s/branch/%s.json", planKey, branchName)
	} else {
		return false, nil, &simpleError{"Project key and/or branch name cannot be empty"}
	}

	request, err := p.client.NewRequest(http.MethodPut, u, nil)
	if err != nil {
		return false, nil, err
	}

	if options != nil && options.VCSBranch != "" {
		values := request.URL.Query()
		values.Add("vcsBranch", options.VCSBranch)
		request.URL.RawQuery = values.Encode()
	}

	response, err := p.client.Do(request, nil)
	if err != nil {
		return false, response, err
	}

	if !(response.StatusCode == 200) {
		return false, response, &simpleError{fmt.Sprintf("Create returned %d", response.StatusCode)}
	}

	return true, response, nil
}

// NumberOfPlans returns the number of plans on the Bamboo server
func (p *PlanService) NumberOfPlans() (int, *http.Response, error) {
	request, err := p.client.NewRequest(http.MethodGet, "plan.json", nil)
	if err != nil {
		return 0, nil, err
	}

	// Restrict results to one for speed
	values := request.URL.Query()
	values.Add("max-results", "1")
	request.URL.RawQuery = values.Encode()

	planResp := PlanResponse{}
	response, err := p.client.Do(request, &planResp)
	if err != nil {
		return 0, response, err
	}

	if response.StatusCode != 200 {
		return 0, response, &simpleError{fmt.Sprintf("Getting the number of plans returned %s", response.Status)}
	}

	return planResp.Plans.Size, response, nil
}

// ListPlans gets information on all plans
func (p *PlanService) ListPlans() ([]*Plan, *http.Response, error) {
	// Get number of plans to set max-results
	numPlans, resp, err := p.NumberOfPlans()
	if err != nil {
		return nil, resp, err
	}

	request, err := p.client.NewRequest(http.MethodGet, "plan.json", nil)
	if err != nil {
		return nil, nil, err
	}

	q := request.URL.Query()
	q.Add("max-results", strconv.Itoa(numPlans))
	request.URL.RawQuery = q.Encode()

	planResp := PlanResponse{}
	response, err := p.client.Do(request, &planResp)
	if err != nil {
		return nil, response, err
	}

	if response.StatusCode != 200 {
		return nil, response, &simpleError{fmt.Sprintf("Getting plan information returned %s", response.Status)}
	}

	return planResp.Plans.PlanList, response, nil
}

// ListPlanKeys get all the plan keys for all build plans on Bamboo
func (p *PlanService) ListPlanKeys() ([]string, *http.Response, error) {
	plans, response, err := p.ListPlans()
	if err != nil {
		return nil, response, err
	}
	keys := make([]string, len(plans))

	for i, p := range plans {
		keys[i] = p.Key
	}
	return keys, response, nil
}

// ListPlanNames returns a list of ShortNames of all plans
func (p *PlanService) ListPlanNames() ([]string, *http.Response, error) {
	plans, response, err := p.ListPlans()
	if err != nil {
		return nil, response, err
	}
	names := make([]string, len(plans))

	for i, p := range plans {
		names[i] = p.ShortName
	}
	return names, response, nil
}

// PlanNameMap returns a map[string]string where the PlanKey is the key and the ShortName is the value
func (p *PlanService) PlanNameMap() (map[string]string, *http.Response, error) {
	plans, response, err := p.ListPlans()
	if err != nil {
		return nil, response, err
	}

	planMap := make(map[string]string, len(plans))

	for _, p := range plans {
		planMap[p.Key] = p.ShortName
	}
	return planMap, response, nil
}

// DisablePlan will disable a plan or plan branch
func (p *PlanService) DisablePlan(planKey string) (*http.Response, error) {
	u := fmt.Sprintf("plan/%s/enable", planKey)
	request, err := p.client.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	response, err := p.client.Do(request, nil)
	if err != nil {
		return response, err
	}
	return response, nil
}

// Run plan without variables
func (p *PlanService) RunPlan(projectKey, planKey string) (*http.Response, error) {
	return p.runPlan(projectKey, planKey, nil)
}

// Run plan with variables
func (p *PlanService) RunPlanCustomized(projectKey, planKey string, variables map[string]string) (*http.Response, error) {
	return p.runPlan(projectKey, planKey, variables)
}

// internal method for build plan running, avoid duplicate
func (p *PlanService) runPlan(projectKey, planKey string, variables map[string]string) (*http.Response, error) {
	var u = ""
	if variables != nil {
		var varsString = ""
		for varName, varValue := range variables {
			varsString += fmt.Sprintf("&bamboo.variable.%s=%s", varName, varValue)
		}
		u = fmt.Sprintf("queue/%s-%s?stage&executeAllStages&%s", projectKey, planKey, varsString)
	} else {
		u = fmt.Sprintf("queue/%s-%s?stage&executeAllStages", projectKey, planKey)
	}
	request, err := p.client.NewRequest(http.MethodPost, u, nil)
	if err != nil {
		return nil, err
	}

	response, err := p.client.Do(request, nil)
	if err != nil {
		return response, err
	}
	return response, nil
}
