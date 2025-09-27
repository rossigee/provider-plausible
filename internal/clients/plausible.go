/*
Copyright 2023 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/rossigee/provider-plausible/apis/v1beta1"
)

const (
	errNoProviderConfig     = "no providerConfig specified"
	errGetProviderConfig    = "cannot get providerConfig"
	errTrackUsage           = "cannot track ProviderConfig usage"
	errExtractCredentials   = "cannot extract credentials"
	errUnmarshalCredentials = "cannot unmarshal credentials"

	// Default Plausible Cloud API URL
	defaultBaseURL = "https://plausible.io"

	// API version
	apiVersion = "v1"
)

// Config holds the configuration for the Plausible API client
type Config struct {
	BaseURL string
	APIKey  string
}

// Credentials holds the API key for Plausible
type Credentials struct {
	APIKey string `json:"apiKey"`
}

// Client is a Plausible API client
type Client struct {
	config     Config
	httpClient *http.Client
}

// NewClient creates a new Plausible API client
func NewClient(cfg Config) *Client {
	return &Client{
		config:     cfg,
		httpClient: &http.Client{},
	}
}

// GetConfig extracts the Plausible client configuration from a ProviderConfig
func GetConfig(ctx context.Context, c client.Client, mg resource.Managed) (*Config, error) {
	pc := &v1beta1.ProviderConfig{}

	// Extract provider config reference using interface conversion
	type providerConfigReferencer interface {
		GetProviderConfigReference() *xpv1.Reference
	}

	pcr, ok := mg.(providerConfigReferencer)
	if !ok {
		return nil, errors.New("managed resource does not implement GetProviderConfigReference")
	}

	pcRef := pcr.GetProviderConfigReference()
	if pcRef == nil {
		return nil, errors.New(errNoProviderConfig)
	}

	if err := c.Get(ctx, client.ObjectKey{Name: pcRef.Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetProviderConfig)
	}

	t := NewProviderConfigUsageTracker(c)
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackUsage)
	}

	data, err := resource.CommonCredentialExtractor(ctx, pc.Spec.Credentials.Source, c, pc.Spec.Credentials.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errExtractCredentials)
	}

	creds := &Credentials{}
	if err := json.Unmarshal(data, creds); err != nil {
		return nil, errors.Wrap(err, errUnmarshalCredentials)
	}

	baseURL := defaultBaseURL
	if pc.Spec.BaseURL != nil && *pc.Spec.BaseURL != "" {
		baseURL = *pc.Spec.BaseURL
	}

	return &Config{
		BaseURL: baseURL,
		APIKey:  creds.APIKey,
	}, nil
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s/api/%s%s", c.config.BaseURL, apiVersion, path)

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal request body")
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}

	return resp, nil
}

// parseResponse reads and unmarshals the response body
func parseResponse(resp *http.Response, target interface{}) error {
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if target != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return errors.Wrap(err, "failed to decode response")
		}
	}

	return nil
}

// Site represents a Plausible site
type Site struct {
	ID       string `json:"id"`
	Domain   string `json:"domain"`
	TeamID   string `json:"team_id,omitempty"`
	Timezone string `json:"timezone,omitempty"`
}

// CreateSiteRequest represents a request to create a site
type CreateSiteRequest struct {
	Domain   string `json:"domain"`
	TeamID   string `json:"team_id,omitempty"`
	Timezone string `json:"timezone,omitempty"`
}

// UpdateSiteRequest represents a request to update a site
type UpdateSiteRequest struct {
	Domain string `json:"domain"`
}

// ListSitesResponse represents the response from listing sites
type ListSitesResponse struct {
	Sites []Site `json:"sites"`
	Meta  struct {
		After  string `json:"after,omitempty"`
		Before string `json:"before,omitempty"`
		Limit  int    `json:"limit"`
	} `json:"meta"`
}

// GetSite retrieves a site by ID
func (c *Client) GetSite(siteID string) (*Site, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/sites/%s", siteID), nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	var site Site
	if err := parseResponse(resp, &site); err != nil {
		return nil, err
	}

	return &site, nil
}

// GetSiteByDomain retrieves a site by domain
func (c *Client) GetSiteByDomain(domain string) (*Site, error) {
	// List sites and filter by domain since there's no direct get-by-domain endpoint
	sites, err := c.ListSites()
	if err != nil {
		return nil, err
	}

	for _, site := range sites {
		if site.Domain == domain {
			return &site, nil
		}
	}

	return nil, nil
}

// ListSites retrieves all sites
func (c *Client) ListSites() ([]Site, error) {
	var allSites []Site
	after := ""

	for {
		path := "/sites"
		if after != "" {
			path = fmt.Sprintf("%s?after=%s", path, url.QueryEscape(after))
		}

		resp, err := c.doRequest("GET", path, nil)
		if err != nil {
			return nil, err
		}

		var listResp ListSitesResponse
		if err := parseResponse(resp, &listResp); err != nil {
			return nil, err
		}

		allSites = append(allSites, listResp.Sites...)

		if listResp.Meta.After == "" {
			break
		}
		after = listResp.Meta.After
	}

	return allSites, nil
}

// CreateSite creates a new site
func (c *Client) CreateSite(req CreateSiteRequest) (*Site, error) {
	resp, err := c.doRequest("POST", "/sites", req)
	if err != nil {
		return nil, err
	}

	var site Site
	if err := parseResponse(resp, &site); err != nil {
		return nil, err
	}

	return &site, nil
}

// UpdateSite updates an existing site's domain
func (c *Client) UpdateSite(siteID string, newDomain string) (*Site, error) {
	req := UpdateSiteRequest{
		Domain: newDomain,
	}

	resp, err := c.doRequest("PUT", fmt.Sprintf("/sites/%s", siteID), req)
	if err != nil {
		return nil, err
	}

	var site Site
	if err := parseResponse(resp, &site); err != nil {
		return nil, err
	}

	return &site, nil
}

// DeleteSite deletes a site
func (c *Client) DeleteSite(siteID string) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/sites/%s", siteID), nil)
	if err != nil {
		return err
	}

	return parseResponse(resp, nil)
}

// Goal represents a Plausible goal
type Goal struct {
	ID        string `json:"id"`
	GoalType  string `json:"goal_type"`
	EventName string `json:"event_name,omitempty"`
	PagePath  string `json:"page_path,omitempty"`
}

// CreateGoalRequest represents a request to create a goal
type CreateGoalRequest struct {
	GoalType  string `json:"goal_type"`
	EventName string `json:"event_name,omitempty"`
	PagePath  string `json:"page_path,omitempty"`
}

// ListGoalsResponse represents the response from listing goals
type ListGoalsResponse struct {
	Goals []Goal `json:"goals"`
	Meta  struct {
		After  string `json:"after,omitempty"`
		Before string `json:"before,omitempty"`
		Limit  int    `json:"limit"`
	} `json:"meta"`
}

// ListGoals retrieves all goals for a site
func (c *Client) ListGoals(siteDomain string) ([]Goal, error) {
	var allGoals []Goal
	after := ""

	for {
		path := fmt.Sprintf("/sites/goals?site_id=%s", url.QueryEscape(siteDomain))
		if after != "" {
			path = fmt.Sprintf("%s&after=%s", path, url.QueryEscape(after))
		}

		resp, err := c.doRequest("GET", path, nil)
		if err != nil {
			return nil, err
		}

		var listResp ListGoalsResponse
		if err := parseResponse(resp, &listResp); err != nil {
			return nil, err
		}

		allGoals = append(allGoals, listResp.Goals...)

		if listResp.Meta.After == "" {
			break
		}
		after = listResp.Meta.After
	}

	return allGoals, nil
}

// GetGoal retrieves a specific goal
func (c *Client) GetGoal(siteDomain string, goalID string) (*Goal, error) {
	goals, err := c.ListGoals(siteDomain)
	if err != nil {
		return nil, err
	}

	for _, goal := range goals {
		if goal.ID == goalID {
			return &goal, nil
		}
	}

	return nil, nil
}

// CreateGoal creates a new goal
func (c *Client) CreateGoal(siteDomain string, req CreateGoalRequest) (*Goal, error) {
	body := map[string]interface{}{
		"site_id":   siteDomain,
		"goal_type": req.GoalType,
	}

	if req.EventName != "" {
		body["event_name"] = req.EventName
	}
	if req.PagePath != "" {
		body["page_path"] = req.PagePath
	}

	resp, err := c.doRequest("PUT", "/sites/goals", body)
	if err != nil {
		return nil, err
	}

	var goal Goal
	if err := parseResponse(resp, &goal); err != nil {
		return nil, err
	}

	return &goal, nil
}

// DeleteGoal deletes a goal
func (c *Client) DeleteGoal(goalID string) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/sites/goals/%s", goalID), nil)
	if err != nil {
		return err
	}

	return parseResponse(resp, nil)
}

// IsNotFound returns true if the error indicates the resource was not found
func IsNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "status 404")
}

// Custom ProviderConfigUsage tracker implementation that works with fake clients
type providerConfigUsageTracker struct {
	kube client.Client
}

// NewProviderConfigUsageTracker creates a ProviderConfigUsage tracker that works with fake clients
func NewProviderConfigUsageTracker(kube client.Client) resource.Tracker {
	return &providerConfigUsageTracker{kube: kube}
}

func (t *providerConfigUsageTracker) Track(ctx context.Context, mg resource.Managed) error {
	pcu := &v1beta1.ProviderConfigUsage{}
	pcu.SetName(string(mg.GetUID()))

	// Handle namespace - use mg namespace or fallback to crossplane-system
	namespace := mg.GetNamespace()
	if namespace == "" {
		namespace = "crossplane-system"
	}
	pcu.SetNamespace(namespace)

	// Set OwnerReferences to create connection
	pcu.SetOwnerReferences([]metav1.OwnerReference{{
		APIVersion: mg.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:       mg.GetObjectKind().GroupVersionKind().Kind,
		Name:       mg.GetName(),
		UID:        mg.GetUID(),
	}})

	// Use CreateOrUpdate for idempotent operation
	return errors.Wrap(client.IgnoreAlreadyExists(t.kube.Create(ctx, pcu)), "cannot create ProviderConfigUsage")
}