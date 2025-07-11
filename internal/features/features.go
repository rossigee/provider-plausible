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

package features

import "github.com/crossplane/crossplane-runtime/pkg/feature"

// Feature flags.
const (
	// EnableAlphaManagementPolicies enables alpha support for
	// Management Policies. See the below design for more details.
	// https://github.com/crossplane/crossplane/blob/91edeae3fcac96c6c8a1759a723981eea4bb77e4/design/design-doc-observe-only-resources.md
	EnableAlphaManagementPolicies feature.Flag = "EnableAlphaManagementPolicies"
)