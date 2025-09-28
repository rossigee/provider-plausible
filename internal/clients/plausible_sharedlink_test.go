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

func TestClient_CreateSharedLink(t *testing.T) {
	tests := []struct {
		name             string
		request          CreateSharedLinkRequest
		responseCode     int
		responseBody     interface{}
		expectedLink     *SharedLink
		expectedError    bool
	}{
		{
			name: "successful creation",
			request: CreateSharedLinkRequest{
				SiteDomain: "example.com",
				Name:       "client-dashboard",
				Password:   "secure123",
			},
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"name":         "client-dashboard",
				"url":          "https://plausible.io/share/example.com?auth=abc123",
				"has_password": true,
			},
			expectedLink: &SharedLink{
				Name:        "client-dashboard",
				URL:         "https://plausible.io/share/example.com?auth=abc123",
				HasPassword: true,
			},
			expectedError: false,
		},
		{
			name: "creation without password",
			request: CreateSharedLinkRequest{
				SiteDomain: "example.com",
				Name:       "public-dashboard",
			},
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"name":         "public-dashboard",
				"url":          "https://plausible.io/share/example.com?auth=def456",
				"has_password": false,
			},
			expectedLink: &SharedLink{
				Name:        "public-dashboard",
				URL:         "https://plausible.io/share/example.com?auth=def456",
				HasPassword: false,
			},
			expectedError: false,
		},
		{
			name: "api error",
			request: CreateSharedLinkRequest{
				SiteDomain: "nonexistent.com",
				Name:       "test-link",
			},
			responseCode:  http.StatusNotFound,
			responseBody:  "Site not found",
			expectedLink:  nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "PUT" {
					t.Errorf("Expected PUT request, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/sites/shared-links" {
					t.Errorf("Expected path /api/v1/sites/shared-links, got %s", r.URL.Path)
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

			result, err := client.CreateSharedLink(tt.request)

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

			if diff := cmp.Diff(tt.expectedLink, result); diff != "" {
				t.Errorf("CreateSharedLink() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_ListSharedLinks(t *testing.T) {
	tests := []struct {
		name          string
		siteDomain    string
		responseCode  int
		responseBody  interface{}
		expectedLinks []SharedLink
		expectedError bool
	}{
		{
			name:         "successful list",
			siteDomain:   "example.com",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"shared_links": []map[string]interface{}{
					{
						"name":         "client-dashboard",
						"url":          "https://plausible.io/share/example.com?auth=abc123",
						"has_password": true,
					},
					{
						"name":         "public-dashboard",
						"url":          "https://plausible.io/share/example.com?auth=def456",
						"has_password": false,
					},
				},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedLinks: []SharedLink{
				{
					Name:        "client-dashboard",
					URL:         "https://plausible.io/share/example.com?auth=abc123",
					HasPassword: true,
				},
				{
					Name:        "public-dashboard",
					URL:         "https://plausible.io/share/example.com?auth=def456",
					HasPassword: false,
				},
			},
			expectedError: false,
		},
		{
			name:          "empty list",
			siteDomain:    "example.com",
			responseCode:  http.StatusOK,
			responseBody: map[string]interface{}{
				"shared_links": []map[string]interface{}{},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedLinks: nil,
			expectedError: false,
		},
		{
			name:          "api error",
			siteDomain:    "nonexistent.com",
			responseCode:  http.StatusNotFound,
			responseBody:  "Site not found",
			expectedLinks: nil,
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

			result, err := client.ListSharedLinks(tt.siteDomain)

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

			if diff := cmp.Diff(tt.expectedLinks, result); diff != "" {
				t.Errorf("ListSharedLinks() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_GetSharedLink(t *testing.T) {
	tests := []struct {
		name          string
		siteDomain    string
		linkName      string
		responseCode  int
		responseBody  interface{}
		expectedLink  *SharedLink
		expectedError bool
	}{
		{
			name:         "link exists",
			siteDomain:   "example.com",
			linkName:     "client-dashboard",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"shared_links": []map[string]interface{}{
					{
						"name":         "client-dashboard",
						"url":          "https://plausible.io/share/example.com?auth=abc123",
						"has_password": true,
					},
				},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedLink: &SharedLink{
				Name:        "client-dashboard",
				URL:         "https://plausible.io/share/example.com?auth=abc123",
				HasPassword: true,
			},
			expectedError: false,
		},
		{
			name:         "link not found",
			siteDomain:   "example.com",
			linkName:     "nonexistent",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"shared_links": []map[string]interface{}{},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedLink:  nil,
			expectedError: false,
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

			result, err := client.GetSharedLink(tt.siteDomain, tt.linkName)

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

			if diff := cmp.Diff(tt.expectedLink, result); diff != "" {
				t.Errorf("GetSharedLink() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_DeleteSharedLink(t *testing.T) {
	tests := []struct {
		name          string
		siteDomain    string
		linkName      string
		responseCode  int
		responseBody  string
		expectedError bool
	}{
		{
			name:          "successful deletion",
			siteDomain:    "example.com",
			linkName:      "client-dashboard",
			responseCode:  http.StatusNoContent,
			responseBody:  "",
			expectedError: false,
		},
		{
			name:          "link not found",
			siteDomain:    "example.com",
			linkName:      "nonexistent",
			responseCode:  http.StatusNotFound,
			responseBody:  "Link not found",
			expectedError: true,
		},
		{
			name:          "api error",
			siteDomain:    "nonexistent.com",
			linkName:      "test-link",
			responseCode:  http.StatusBadRequest,
			responseBody:  "Site not found",
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

			err := client.DeleteSharedLink(tt.siteDomain, tt.linkName)

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