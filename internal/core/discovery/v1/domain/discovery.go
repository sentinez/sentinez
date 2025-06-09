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

// Package dcvrdomain implements the discovery business logic
package dcvrdomain

import (
	"context"

	discoverypb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/discovery/v1"
	dcvrrepo "github.com/sentinez/sentinez/internal/core/discovery/v1/repos"
	"github.com/sentinez/sentinez/pkg/std/eventq"
)

var _ Discovery = (*discovery)(nil)

// Discovery implement discovery registry business logic
type Discovery interface {
	RegisterService(
		ctx context.Context, id string, req *discoverypb.RegisterRequest) error
}

// New creates a new discovery service
// and returns a Discovery interface.
func New(repo dcvrrepo.Discovery) Discovery {
	return &discovery{
		repo: repo,
	}
}

type discovery struct {
	repo dcvrrepo.Discovery
}

// RegisterService implements Discovery
func (d *discovery) RegisterService(ctx context.Context,
	id string, req *discoverypb.RegisterRequest) error {

	_ = ctx
	_ = req
	_ = id

	eventq.Publish(req.GetName(), req.GetAddress())

	return nil
}
