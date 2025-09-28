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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestClient_GetSiteByDomain(t *testing.T) {
	tests := []struct {
		name          string
		domain        string
		responseCode  int
		responseBody  interface{}
		expectedSite  *Site
		expectedError bool
	}{
		{
			name:         "site found by domain",
			domain:       "example.com",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"sites": []map[string]interface{}{
					{
						"domain":   "example.com",
						"timezone": "UTC",
					},
					{
						"domain":   "other.com",
						"timezone": "EST",
					},
				},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedSite: &Site{
				Domain:   "example.com",
				Timezone: "UTC",
			},
			expectedError: false,
		},
		{
			name:         "site not found by domain",
			domain:       "nonexistent.com",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"sites": []map[string]interface{}{
					{
						"domain":   "other.com",
						"timezone": "EST",
					},
				},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedSite:  nil,
			expectedError: false,
		},
		{
			name:          "api error",
			domain:        "test.com",
			responseCode:  http.StatusUnauthorized,
			responseBody:  "Unauthorized",
			expectedSite:  nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/sites" {
					t.Errorf("Expected path /api/v1/sites, got %s", r.URL.Path)
				}

				w.WriteHeader(tt.responseCode)
				if tt.responseCode >= 400 {
					_, _ = w.Write([]byte(tt.responseBody.(string)))
				} else {
					_ = json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			client := NewClient(Config{
				BaseURL: server.URL,
				APIKey:  "test-key",
			})

			result, err := client.GetSiteByDomain(tt.domain)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if diff := cmp.Diff(tt.expectedSite, result); diff != "" {
				t.Errorf("GetSiteByDomain() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_ListSites(t *testing.T) {
	tests := []struct {
		name          string
		responseCode  int
		responseBody  interface{}
		expectedSites []Site
		expectedError bool
	}{
		{
			name:         "successful list",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"sites": []map[string]interface{}{
					{
						"domain":   "example.com",
						"timezone": "UTC",
					},
					{
						"domain":   "test.com",
						"timezone": "EST",
					},
				},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedSites: []Site{
				{
					Domain:   "example.com",
					Timezone: "UTC",
				},
				{
					Domain:   "test.com",
					Timezone: "EST",
				},
			},
			expectedError: false,
		},
		{
			name:         "empty list",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"sites": []map[string]interface{}{},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedSites: nil,
			expectedError: false,
		},
		{
			name:          "api error",
			responseCode:  http.StatusUnauthorized,
			responseBody:  "Unauthorized",
			expectedSites: nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/sites" {
					t.Errorf("Expected path /api/v1/sites, got %s", r.URL.Path)
				}

				w.WriteHeader(tt.responseCode)
				if tt.responseCode >= 400 {
					_, _ = w.Write([]byte(tt.responseBody.(string)))
				} else {
					_ = json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			client := NewClient(Config{
				BaseURL: server.URL,
				APIKey:  "test-key",
			})

			result, err := client.ListSites()

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if diff := cmp.Diff(tt.expectedSites, result); diff != "" {
				t.Errorf("ListSites() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_ListSites_Pagination(t *testing.T) {
	// Test pagination handling
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++

		var response map[string]interface{}

		if calls == 1 {
			// First page
			response = map[string]interface{}{
				"sites": []map[string]interface{}{
					{
						"domain":   "site1.com",
						"timezone": "UTC",
					},
				},
				"meta": map[string]interface{}{
					"limit": 1,
					"after": "cursor-123",
				},
			}
		} else {
			// Second page (last page)
			response = map[string]interface{}{
				"sites": []map[string]interface{}{
					{
						"domain":   "site2.com",
						"timezone": "EST",
					},
				},
				"meta": map[string]interface{}{
					"limit": 1,
					"after": "",
				},
			}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})

	result, err := client.ListSites()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	expectedSites := []Site{
		{
			Domain:   "site1.com",
			Timezone: "UTC",
		},
		{
			Domain:   "site2.com",
			Timezone: "EST",
		},
	}

	if diff := cmp.Diff(expectedSites, result); diff != "" {
		t.Errorf("ListSites() pagination mismatch (-want +got):\n%s", diff)
	}

	if calls != 2 {
		t.Errorf("Expected 2 API calls for pagination, got %d", calls)
	}
}

func TestClient_ListGoals(t *testing.T) {
	tests := []struct {
		name          string
		siteDomain    string
		responseCode  int
		responseBody  interface{}
		expectedGoals []Goal
		expectedError bool
	}{
		{
			name:         "successful list",
			siteDomain:   "example.com",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"goals": []map[string]interface{}{
					{
						"id":         "goal-123",
						"goal_type":  "event",
						"event_name": "signup",
					},
					{
						"id":        "goal-456",
						"goal_type": "page",
						"page_path": "/checkout",
					},
				},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedGoals: []Goal{
				{
					ID:        "goal-123",
					GoalType:  "event",
					EventName: "signup",
				},
				{
					ID:       "goal-456",
					GoalType: "page",
					PagePath: "/checkout",
				},
			},
			expectedError: false,
		},
		{
			name:         "empty list",
			siteDomain:   "example.com",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"goals": []map[string]interface{}{},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedGoals: nil,
			expectedError: false,
		},
		{
			name:          "api error",
			siteDomain:    "nonexistent.com",
			responseCode:  http.StatusNotFound,
			responseBody:  "Site not found",
			expectedGoals: nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				if r.URL.Query().Get("site_id") != tt.siteDomain {
					t.Errorf("Expected site_id %s, got %s", tt.siteDomain, r.URL.Query().Get("site_id"))
				}

				w.WriteHeader(tt.responseCode)
				if tt.responseCode >= 400 {
					_, _ = w.Write([]byte(tt.responseBody.(string)))
				} else {
					_ = json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			client := NewClient(Config{
				BaseURL: server.URL,
				APIKey:  "test-key",
			})

			result, err := client.ListGoals(tt.siteDomain)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if diff := cmp.Diff(tt.expectedGoals, result); diff != "" {
				t.Errorf("ListGoals() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_ListGoals_Pagination(t *testing.T) {
	// Test pagination handling
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++

		var response map[string]interface{}

		if calls == 1 {
			// First page
			response = map[string]interface{}{
				"goals": []map[string]interface{}{
					{
						"id":         "goal-1",
						"goal_type":  "event",
						"event_name": "signup",
					},
				},
				"meta": map[string]interface{}{
					"limit": 1,
					"after": "cursor-123",
				},
			}
		} else {
			// Second page (last page)
			response = map[string]interface{}{
				"goals": []map[string]interface{}{
					{
						"id":        "goal-2",
						"goal_type": "page",
						"page_path": "/checkout",
					},
				},
				"meta": map[string]interface{}{
					"limit": 1,
					"after": "",
				},
			}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})

	result, err := client.ListGoals("example.com")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	expectedGoals := []Goal{
		{
			ID:        "goal-1",
			GoalType:  "event",
			EventName: "signup",
		},
		{
			ID:       "goal-2",
			GoalType: "page",
			PagePath: "/checkout",
		},
	}

	if diff := cmp.Diff(expectedGoals, result); diff != "" {
		t.Errorf("ListGoals() pagination mismatch (-want +got):\n%s", diff)
	}

	if calls != 2 {
		t.Errorf("Expected 2 API calls for pagination, got %d", calls)
	}
}

func TestClient_GetGoal(t *testing.T) {
	tests := []struct {
		name          string
		siteDomain    string
		goalID        string
		responseCode  int
		responseBody  interface{}
		expectedGoal  *Goal
		expectedError bool
	}{
		{
			name:         "goal exists",
			siteDomain:   "example.com",
			goalID:       "goal-123",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"goals": []map[string]interface{}{
					{
						"id":         "goal-123",
						"goal_type":  "event",
						"event_name": "signup",
					},
				},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedGoal: &Goal{
				ID:        "goal-123",
				GoalType:  "event",
				EventName: "signup",
			},
			expectedError: false,
		},
		{
			name:         "goal not found",
			siteDomain:   "example.com",
			goalID:       "nonexistent",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"goals": []map[string]interface{}{},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedGoal:  nil,
			expectedError: false,
		},
		{
			name:          "api error",
			siteDomain:    "nonexistent.com",
			goalID:        "goal-123",
			responseCode:  http.StatusNotFound,
			responseBody:  "Site not found",
			expectedGoal:  nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseCode)
				if tt.responseCode >= 400 {
					_, _ = w.Write([]byte(tt.responseBody.(string)))
				} else {
					_ = json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			client := NewClient(Config{
				BaseURL: server.URL,
				APIKey:  "test-key",
			})

			result, err := client.GetGoal(tt.siteDomain, tt.goalID)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if diff := cmp.Diff(tt.expectedGoal, result); diff != "" {
				t.Errorf("GetGoal() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_CreateGoal(t *testing.T) {
	tests := []struct {
		name          string
		siteDomain    string
		request       CreateGoalRequest
		responseCode  int
		responseBody  interface{}
		expectedGoal  *Goal
		expectedError bool
	}{
		{
			name:       "successful event goal creation",
			siteDomain: "example.com",
			request: CreateGoalRequest{
				GoalType:  "event",
				EventName: "signup",
			},
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"id":         "goal-123",
				"goal_type":  "event",
				"event_name": "signup",
			},
			expectedGoal: &Goal{
				ID:        "goal-123",
				GoalType:  "event",
				EventName: "signup",
			},
			expectedError: false,
		},
		{
			name:       "successful page goal creation",
			siteDomain: "example.com",
			request: CreateGoalRequest{
				GoalType: "page",
				PagePath: "/checkout",
			},
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"id":        "goal-456",
				"goal_type": "page",
				"page_path": "/checkout",
			},
			expectedGoal: &Goal{
				ID:       "goal-456",
				GoalType: "page",
				PagePath: "/checkout",
			},
			expectedError: false,
		},
		{
			name:       "api error",
			siteDomain: "nonexistent.com",
			request: CreateGoalRequest{
				GoalType:  "event",
				EventName: "test",
			},
			responseCode:  http.StatusNotFound,
			responseBody:  "Site not found",
			expectedGoal:  nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "PUT" {
					t.Errorf("Expected PUT request, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/sites/goals" {
					t.Errorf("Expected path /api/v1/sites/goals, got %s", r.URL.Path)
				}

				w.WriteHeader(tt.responseCode)
				if tt.responseCode >= 400 {
					_, _ = w.Write([]byte(tt.responseBody.(string)))
				} else {
					_ = json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			client := NewClient(Config{
				BaseURL: server.URL,
				APIKey:  "test-key",
			})

			result, err := client.CreateGoal(tt.siteDomain, tt.request)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if diff := cmp.Diff(tt.expectedGoal, result); diff != "" {
				t.Errorf("CreateGoal() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_DeleteGoal(t *testing.T) {
	tests := []struct {
		name          string
		goalID        string
		responseCode  int
		responseBody  string
		expectedError bool
	}{
		{
			name:          "successful deletion",
			goalID:        "goal-123",
			responseCode:  http.StatusNoContent,
			responseBody:  "",
			expectedError: false,
		},
		{
			name:          "goal not found",
			goalID:        "nonexistent",
			responseCode:  http.StatusNotFound,
			responseBody:  "Goal not found",
			expectedError: true,
		},
		{
			name:          "api error",
			goalID:        "goal-123",
			responseCode:  http.StatusBadRequest,
			responseBody:  "Bad request",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "DELETE" {
					t.Errorf("Expected DELETE request, got %s", r.Method)
				}

				w.WriteHeader(tt.responseCode)
				if tt.responseBody != "" {
					_, _ = w.Write([]byte(tt.responseBody))
				}
			}))
			defer server.Close()

			client := NewClient(Config{
				BaseURL: server.URL,
				APIKey:  "test-key",
			})

			err := client.DeleteGoal(tt.goalID)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}