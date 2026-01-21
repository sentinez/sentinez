// Copyright 2025 Sentinéz Labs.
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

package resourcerepo

import (
	"context"

	tenantpb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/tenant/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	"github.com/sentinez/sentinez/internal/shared/tables"
	"github.com/sentinez/sentinez/pkg/storage/dbx"
	"github.com/sentinez/sentinez/pkg/storage/dbx/postgres"
)

func New(ctx context.Context, appConf *confpb.Config) (*Resource, error) {
	storage, err := postgres.New[*tenantpb.Resource](ctx, appConf,
		dbx.WithTable(tables.Resources),
	)
	if err != nil {
		return nil, err
	}

	return &Resource{
		storage: storage,
	}, nil
}

var _ IResource = (*Resource)(nil)

// nolint
type IResource interface {
	Create(ctx context.Context, rs *tenantpb.Resource) (*tenantpb.Resource, error)
	Update(ctx context.Context, rs *tenantpb.Resource) (*tenantpb.Resource, error)
	Get(ctx context.Context, id string) (*tenantpb.Resource, error)
	Delete(ctx context.Context, id string) error
}

type Resource struct {
	storage dbx.Database[*tenantpb.Resource]
}

func (rsc *Resource) Create(
	ctx context.Context, rs *tenantpb.Resource) (*tenantpb.Resource, error) {
	//TODO implement me
	panic("implement me")
}

func (rsc *Resource) Update(
	ctx context.Context, rs *tenantpb.Resource) (*tenantpb.Resource, error) {
	//TODO implement me
	panic("implement me")
}

func (rsc *Resource) Get(
	ctx context.Context, id string) (*tenantpb.Resource, error) {
	//TODO implement me
	panic("implement me")
}

func (rsc *Resource) Delete(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}
