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

func TestClient_ListTeams(t *testing.T) {
	tests := []struct {
		name          string
		responseCode  int
		responseBody  interface{}
		expectedTeams []Team
		expectedError bool
	}{
		{
			name:         "successful list",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"teams": []map[string]interface{}{
					{
						"id":          "team-123",
						"name":        "Marketing Team",
						"api_enabled": true,
					},
					{
						"id":          "team-456",
						"name":        "Development Team",
						"api_enabled": false,
					},
				},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedTeams: []Team{
				{
					ID:         "team-123",
					Name:       "Marketing Team",
					APIEnabled: true,
				},
				{
					ID:         "team-456",
					Name:       "Development Team",
					APIEnabled: false,
				},
			},
			expectedError: false,
		},
		{
			name:         "empty list",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"teams": []map[string]interface{}{},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedTeams: nil,
			expectedError: false,
		},
		{
			name:          "api error",
			responseCode:  http.StatusUnauthorized,
			responseBody:  "Unauthorized",
			expectedTeams: nil,
			expectedError: true,
		},
		{
			name:         "pagination test",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"teams": []map[string]interface{}{
					{
						"id":          "team-789",
						"name":        "Sales Team",
						"api_enabled": true,
					},
				},
				"meta": map[string]interface{}{
					"limit": 1,
					"after": "",
				},
			},
			expectedTeams: []Team{
				{
					ID:         "team-789",
					Name:       "Sales Team",
					APIEnabled: true,
				},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/sites/teams" {
					t.Errorf("Expected path /api/v1/sites/teams, got %s", r.URL.Path)
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

			result, err := client.ListTeams()

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

			if diff := cmp.Diff(tt.expectedTeams, result); diff != "" {
				t.Errorf("ListTeams() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_ListTeams_Pagination(t *testing.T) {
	// Test pagination handling
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++

		var response map[string]interface{}

		if calls == 1 {
			// First page
			response = map[string]interface{}{
				"teams": []map[string]interface{}{
					{
						"id":          "team-1",
						"name":        "Team 1",
						"api_enabled": true,
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
				"teams": []map[string]interface{}{
					{
						"id":          "team-2",
						"name":        "Team 2",
						"api_enabled": false,
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

	result, err := client.ListTeams()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	expectedTeams := []Team{
		{
			ID:         "team-1",
			Name:       "Team 1",
			APIEnabled: true,
		},
		{
			ID:         "team-2",
			Name:       "Team 2",
			APIEnabled: false,
		},
	}

	if diff := cmp.Diff(expectedTeams, result); diff != "" {
		t.Errorf("ListTeams() pagination mismatch (-want +got):\n%s", diff)
	}

	if calls != 2 {
		t.Errorf("Expected 2 API calls for pagination, got %d", calls)
	}
}