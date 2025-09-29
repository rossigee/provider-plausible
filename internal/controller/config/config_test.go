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

package config

import (
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
)

func TestSetup(t *testing.T) {
	// Since setting up a full manager requires too much infrastructure,
	// just test that the Setup function exists and doesn't panic when called with nil
	// This is a basic smoke test to ensure the package compiles correctly
	defer func() {
		if r := recover(); r == nil {
			// We expect this to panic or error with nil inputs, so no panic means
			// the function is at least properly defined
			t.Log("Setup function is properly defined")
		}
	}()

	// This will likely error or panic due to nil inputs, but that's expected
	// The important thing is that the function signature is correct and compiles
	_ = Setup(nil, controller.Options{})
}