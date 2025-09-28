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

func TestClient_CreateCustomProperty(t *testing.T) {
	tests := []struct {
		name             string
		request          CreateCustomPropertyRequest
		responseCode     int
		responseBody     interface{}
		expectedProperty *CustomProperty
		expectedError    bool
	}{
		{
			name: "successful creation",
			request: CreateCustomPropertyRequest{
				SiteDomain:  "example.com",
				Key:         "user_segment",
				Description: "Customer segment tracking",
			},
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"key":         "user_segment",
				"description": "Customer segment tracking",
				"is_enabled":  true,
			},
			expectedProperty: &CustomProperty{
				Key:         "user_segment",
				Description: "Customer segment tracking",
				IsEnabled:   true,
			},
			expectedError: false,
		},
		{
			name: "creation without description",
			request: CreateCustomPropertyRequest{
				SiteDomain: "example.com",
				Key:        "page_category",
			},
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"key":         "page_category",
				"description": "",
				"is_enabled":  true,
			},
			expectedProperty: &CustomProperty{
				Key:         "page_category",
				Description: "",
				IsEnabled:   true,
			},
			expectedError: false,
		},
		{
			name: "api error",
			request: CreateCustomPropertyRequest{
				SiteDomain: "nonexistent.com",
				Key:        "test_prop",
			},
			responseCode:     http.StatusNotFound,
			responseBody:     "Site not found",
			expectedProperty: nil,
			expectedError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "PUT" {
					t.Errorf("Expected PUT request, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/sites/custom-props" {
					t.Errorf("Expected path /api/v1/sites/custom-props, got %s", r.URL.Path)
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

			result, err := client.CreateCustomProperty(tt.request)

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

			if diff := cmp.Diff(tt.expectedProperty, result); diff != "" {
				t.Errorf("CreateCustomProperty() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_ListCustomProperties(t *testing.T) {
	tests := []struct {
		name               string
		siteDomain         string
		responseCode       int
		responseBody       interface{}
		expectedProperties []CustomProperty
		expectedError      bool
	}{
		{
			name:         "successful list",
			siteDomain:   "example.com",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"custom_properties": []map[string]interface{}{
					{
						"key":         "user_segment",
						"description": "Customer segment tracking",
						"is_enabled":  true,
					},
					{
						"key":         "page_category",
						"description": "Page categorization",
						"is_enabled":  false,
					},
				},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedProperties: []CustomProperty{
				{
					Key:         "user_segment",
					Description: "Customer segment tracking",
					IsEnabled:   true,
				},
				{
					Key:         "page_category",
					Description: "Page categorization",
					IsEnabled:   false,
				},
			},
			expectedError: false,
		},
		{
			name:         "empty list",
			siteDomain:   "example.com",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"custom_properties": []map[string]interface{}{},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedProperties: nil,
			expectedError:      false,
		},
		{
			name:               "api error",
			siteDomain:         "nonexistent.com",
			responseCode:       http.StatusNotFound,
			responseBody:       "Site not found",
			expectedProperties: nil,
			expectedError:      true,
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

			result, err := client.ListCustomProperties(tt.siteDomain)

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

			if diff := cmp.Diff(tt.expectedProperties, result); diff != "" {
				t.Errorf("ListCustomProperties() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_GetCustomProperty(t *testing.T) {
	tests := []struct {
		name             string
		siteDomain       string
		propertyKey      string
		responseCode     int
		responseBody     interface{}
		expectedProperty *CustomProperty
		expectedError    bool
	}{
		{
			name:         "property exists",
			siteDomain:   "example.com",
			propertyKey:  "user_segment",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"custom_properties": []map[string]interface{}{
					{
						"key":         "user_segment",
						"description": "Customer segment tracking",
						"is_enabled":  true,
					},
				},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedProperty: &CustomProperty{
				Key:         "user_segment",
				Description: "Customer segment tracking",
				IsEnabled:   true,
			},
			expectedError: false,
		},
		{
			name:         "property not found",
			siteDomain:   "example.com",
			propertyKey:  "nonexistent",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{
				"custom_properties": []map[string]interface{}{},
				"meta": map[string]interface{}{
					"limit": 100,
				},
			},
			expectedProperty: nil,
			expectedError:    false,
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

			result, err := client.GetCustomProperty(tt.siteDomain, tt.propertyKey)

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

			if diff := cmp.Diff(tt.expectedProperty, result); diff != "" {
				t.Errorf("GetCustomProperty() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_DeleteCustomProperty(t *testing.T) {
	tests := []struct {
		name          string
		siteDomain    string
		propertyKey   string
		responseCode  int
		responseBody  string
		expectedError bool
	}{
		{
			name:          "successful deletion",
			siteDomain:    "example.com",
			propertyKey:   "user_segment",
			responseCode:  http.StatusNoContent,
			responseBody:  "",
			expectedError: false,
		},
		{
			name:          "property not found",
			siteDomain:    "example.com",
			propertyKey:   "nonexistent",
			responseCode:  http.StatusNotFound,
			responseBody:  "Property not found",
			expectedError: true,
		},
		{
			name:          "api error",
			siteDomain:    "nonexistent.com",
			propertyKey:   "test_prop",
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

			err := client.DeleteCustomProperty(tt.siteDomain, tt.propertyKey)

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