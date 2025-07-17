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
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestClient_GetSite(t *testing.T) {
	tests := []struct {
		name          string
		siteID        string
		responseCode  int
		responseBody  interface{}
		expectedSite  *Site
		expectedError bool
	}{
		{
			name:         "site exists",
			siteID:       "example.com",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"domain":   "example.com",
				"timezone": "UTC",
			},
			expectedSite: &Site{
				Domain:   "example.com",
				Timezone: "UTC",
			},
			expectedError: false,
		},
		{
			name:          "site not found",
			siteID:        "nonexistent.com",
			responseCode:  http.StatusNotFound,
			responseBody:  map[string]interface{}{"error": "Site not found"},
			expectedSite:  nil,
			expectedError: false, // GetSite returns nil for 404
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/api/v1/sites/" + tt.siteID
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}
				
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseCode)
				if tt.responseBody != nil {
					_ = json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			client := &Client{
				config: Config{
					BaseURL: server.URL,
					APIKey:  "test-key",
				},
				httpClient: &http.Client{},
			}

			site, err := client.GetSite(tt.siteID)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !cmp.Equal(site, tt.expectedSite) {
				t.Errorf("Site mismatch: %s", cmp.Diff(tt.expectedSite, site))
			}
		})
	}
}

func TestClient_CreateSite_Simple(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/sites" {
			t.Errorf("Expected path /api/v1/sites, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization header 'Bearer test-key', got %s", r.Header.Get("Authorization"))
		}

		var req CreateSiteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := map[string]interface{}{
			"domain":   req.Domain,
			"timezone": req.Timezone,
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		config: Config{
			BaseURL: server.URL,
			APIKey:  "test-key",
		},
		httpClient: &http.Client{},
	}

	req := CreateSiteRequest{
		Domain:   "test.example.com",
		Timezone: "UTC",
	}

	site, err := client.CreateSite(req)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if site == nil {
		t.Error("Expected site but got nil")
		return
	}
	if site.Domain != "test.example.com" {
		t.Errorf("Expected domain 'test.example.com', got %s", site.Domain)
	}
}

func TestClient_UpdateSite_Simple(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		expectedPath := "/api/v1/sites/old.example.com"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"domain":   "new.example.com",
			"timezone": "UTC",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		config: Config{
			BaseURL: server.URL,
			APIKey:  "test-key",
		},
		httpClient: &http.Client{},
	}

	site, err := client.UpdateSite("old.example.com", "new.example.com")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if site == nil {
		t.Error("Expected site but got nil")
		return
	}
	if site.Domain != "new.example.com" {
		t.Errorf("Expected domain 'new.example.com', got %s", site.Domain)
	}
}

func TestClient_DeleteSite_Simple(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		expectedPath := "/api/v1/sites/example.com"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := &Client{
		config: Config{
			BaseURL: server.URL,
			APIKey:  "test-key",
		},
		httpClient: &http.Client{},
	}

	err := client.DeleteSite("example.com")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestIsNotFound_Simple(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "404 error",
			err:      &simpleError{msg: "API request failed with status 404: Not Found"},
			expected: true,
		},
		{
			name:     "500 error",
			err:      &simpleError{msg: "API request failed with status 500: Internal Server Error"},
			expected: false,
		},
		{
			name:     "other error",
			err:      &simpleError{msg: "connection refused"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFound(tt.err)
			if result != tt.expected {
				t.Errorf("IsNotFound() = %v, want %v", result, tt.expected)
			}
		})
	}
}

type simpleError struct {
	msg string
}

func (e *simpleError) Error() string {
	return e.msg
}

func TestIsNotFound_Production(t *testing.T) {
	// Test with actual error from production
	err := &simpleError{msg: "API request failed with status 404: {\"error\":\"Site not found\"}"}
	if !IsNotFound(err) {
		t.Error("Expected IsNotFound to return true for 404 error")
	}
	
	// Test string contains logic
	if !strings.Contains(err.Error(), "404") {
		t.Error("Error message should contain '404'")
	}
}