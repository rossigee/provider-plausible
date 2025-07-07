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
	"testing"
)

func TestIsNotFound(t *testing.T) {
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
			err:      &testError{msg: "API request failed with status 404: Not Found"},
			expected: true,
		},
		{
			name:     "other error",
			err:      &testError{msg: "API request failed with status 500: Internal Server Error"},
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

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestNewClient(t *testing.T) {
	cfg := Config{
		BaseURL: "https://plausible.io",
		APIKey:  "test-key",
	}

	client := NewClient(cfg)

	if client == nil {
		t.Error("NewClient() returned nil")
	}

	if client.config.BaseURL != cfg.BaseURL {
		t.Errorf("client.config.BaseURL = %v, want %v", client.config.BaseURL, cfg.BaseURL)
	}

	if client.config.APIKey != cfg.APIKey {
		t.Errorf("client.config.APIKey = %v, want %v", client.config.APIKey, cfg.APIKey)
	}

	if client.httpClient == nil {
		t.Error("client.httpClient is nil")
	}
}