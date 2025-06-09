// Copyright 2025 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package discovery implements the discovery service
package discovery

import (
	dcvrdomain "github.com/sentinez/sentinez/internal/core/discovery/v1/domain"
	dcvrhandler "github.com/sentinez/sentinez/internal/core/discovery/v1/handler"
	dcvrrepo "github.com/sentinez/sentinez/internal/core/discovery/v1/repos"
	"github.com/sentinez/sentinez/pkg/core/sentinez/v1"
)

var (
	// discovery dependency repo - domain - controller
	_ = sentinez.Inject(dcvrrepo.New)
	_ = sentinez.Inject(dcvrdomain.New)
	_ = sentinez.Inject(dcvrhandler.New)
)
