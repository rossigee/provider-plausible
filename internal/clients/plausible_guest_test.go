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

func TestClient_CreateGuest(t *testing.T) {
	tests := []struct {
		name          string
		request       CreateGuestRequest
		responseCode  int
		responseBody  interface{}
		expectedGuest *Guest
		expectedError bool
	}{
		{
			name: "successful invitation",
			request: CreateGuestRequest{
				SiteDomain: "example.com",
				Email:      "analyst@company.com",
				Role:       "viewer",
			},
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"email":      "analyst@company.com",
				"role":       "viewer",
				"status":     "pending",
				"invited_at": "2023-10-01T12:00:00Z",
			},
			expectedGuest: &Guest{
				Email:     "analyst@company.com",
				Role:      "viewer",
				Status:    "pending",
				InvitedAt: "2023-10-01T12:00:00Z",
			},
			expectedError: false,
		},
		{
			name: "admin invitation",
			request: CreateGuestRequest{
				SiteDomain: "example.com",
				Email:      "admin@company.com",
				Role:       "admin",
			},
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"email":      "admin@company.com",
				"role":       "admin",
				"status":     "pending",
				"invited_at": "2023-10-01T12:00:00Z",
			},
			expectedGuest: &Guest{
				Email:     "admin@company.com",
				Role:      "admin",
				Status:    "pending",
				InvitedAt: "2023-10-01T12:00:00Z",
			},
			expectedError: false,
		},
		{
			name: "api error",
			request: CreateGuestRequest{
				SiteDomain: "nonexistent.com",
				Email:      "test@example.com",
				Role:       "viewer",
			},
			responseCode:  http.StatusNotFound,
			responseBody:  "Site not found",
			expectedGuest: nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "PUT" {
					t.Errorf("Expected PUT request, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/sites/guests" {
					t.Errorf("Expected path /api/v1/sites/guests, got %s", r.URL.Path)
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

			result, err := client.CreateGuest(tt.request)

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

			if diff := cmp.Diff(tt.expectedGuest, result); diff != "" {
				t.Errorf("CreateGuest() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_ListGuests(t *testing.T) {
	tests := []struct {
		name           string
		siteDomain     string
		responseCode   int
		responseBody   interface{}
		expectedGuests []Guest
		expectedError  bool
	}{
		{
			name:         "successful list",
			siteDomain:   "example.com",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"guests": []map[string]interface{}{
					{
						"email":       "analyst@company.com",
						"role":        "viewer",
						"status":      "accepted",
						"invited_at":  "2023-10-01T12:00:00Z",
						"accepted_at": "2023-10-01T13:00:00Z",
					},
					{
						"email":      "admin@company.com",
						"role":       "admin",
						"status":     "pending",
						"invited_at": "2023-10-01T14:00:00Z",
					},
				},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedGuests: []Guest{
				{
					Email:      "analyst@company.com",
					Role:       "viewer",
					Status:     "accepted",
					InvitedAt:  "2023-10-01T12:00:00Z",
					AcceptedAt: "2023-10-01T13:00:00Z",
				},
				{
					Email:     "admin@company.com",
					Role:      "admin",
					Status:    "pending",
					InvitedAt: "2023-10-01T14:00:00Z",
				},
			},
			expectedError: false,
		},
		{
			name:         "empty list",
			siteDomain:   "example.com",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"guests": []map[string]interface{}{},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedGuests: nil,
			expectedError:  false,
		},
		{
			name:           "api error",
			siteDomain:     "nonexistent.com",
			responseCode:   http.StatusNotFound,
			responseBody:   "Site not found",
			expectedGuests: nil,
			expectedError:  true,
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

			result, err := client.ListGuests(tt.siteDomain)

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

			if diff := cmp.Diff(tt.expectedGuests, result); diff != "" {
				t.Errorf("ListGuests() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_GetGuest(t *testing.T) {
	tests := []struct {
		name          string
		siteDomain    string
		email         string
		responseCode  int
		responseBody  interface{}
		expectedGuest *Guest
		expectedError bool
	}{
		{
			name:         "guest exists",
			siteDomain:   "example.com",
			email:        "analyst@company.com",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"guests": []map[string]interface{}{
					{
						"email":       "analyst@company.com",
						"role":        "viewer",
						"status":      "accepted",
						"invited_at":  "2023-10-01T12:00:00Z",
						"accepted_at": "2023-10-01T13:00:00Z",
					},
				},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedGuest: &Guest{
				Email:      "analyst@company.com",
				Role:       "viewer",
				Status:     "accepted",
				InvitedAt:  "2023-10-01T12:00:00Z",
				AcceptedAt: "2023-10-01T13:00:00Z",
			},
			expectedError: false,
		},
		{
			name:         "guest not found",
			siteDomain:   "example.com",
			email:        "nonexistent@example.com",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"guests": []map[string]interface{}{},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedGuest: nil,
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

			result, err := client.GetGuest(tt.siteDomain, tt.email)

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

			if diff := cmp.Diff(tt.expectedGuest, result); diff != "" {
				t.Errorf("GetGuest() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_DeleteGuest(t *testing.T) {
	tests := []struct {
		name          string
		siteDomain    string
		email         string
		responseCode  int
		responseBody  string
		expectedError bool
	}{
		{
			name:          "successful deletion",
			siteDomain:    "example.com",
			email:         "analyst@company.com",
			responseCode:  http.StatusNoContent,
			responseBody:  "",
			expectedError: false,
		},
		{
			name:          "guest not found",
			siteDomain:    "example.com",
			email:         "nonexistent@example.com",
			responseCode:  http.StatusNotFound,
			responseBody:  "Guest not found",
			expectedError: true,
		},
		{
			name:          "api error",
			siteDomain:    "nonexistent.com",
			email:         "test@example.com",
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

			err := client.DeleteGuest(tt.siteDomain, tt.email)

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