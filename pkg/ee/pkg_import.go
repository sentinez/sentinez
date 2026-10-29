// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package ee holds the enterprise features. Only this file and the
// *_oss.go stubs live in this repository: the rest is a private
// repository, checked out here by make ee.checkout and built with the
// enterprise tag (make ... EE=1). Without it the stubs
// keep the build working with the features off.
package ee
