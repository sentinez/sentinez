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

	tnpb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/tenant/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	"github.com/sentinez/sentinez/internal/shared/tables"
	"github.com/sentinez/sentinez/pkg/storage/database"
	"github.com/sentinez/sentinez/pkg/storage/database/postgres"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func New(appConf *confpb.Config) (*Resource, error) {
	storage, err := postgres.New[*tnpb.Resource](appConf, tables.Resources)
	if err != nil {
		return nil, err
	}

	return &Resource{
		storage: storage,
	}, nil
}

var _ IResource = (*Resource)(nil)

type IResource interface {
	Create(ctx context.Context, rs *tnpb.Resource) (*tnpb.Resource, error)
	Update(ctx context.Context, rs *tnpb.Resource) (*tnpb.Resource, error)
	Get(ctx context.Context, id string) (*tnpb.Resource, error)
	Delete(ctx context.Context, id string) error
}

type Resource struct {
	storage database.Database[*tnpb.Resource]
}

func (rsc *Resource) Create(
	ctx context.Context, rs *tnpb.Resource) (*tnpb.Resource, error) {
	//TODO implement me
	panic("implement me")
}

func (rsc *Resource) Update(
	ctx context.Context, rs *tnpb.Resource) (*tnpb.Resource, error) {
	rs.Metadata.UpdatedAt = timestamppb.Now()

	err := rsc.storage.Set(ctx, rs.GetId(), rs)
	if err != nil {
		return nil, err
	}

	return rs, nil
}

func (rsc *Resource) Get(
	ctx context.Context, id string) (*tnpb.Resource, error) {
	return rsc.storage.Get(ctx, id)
}

func (rsc *Resource) Delete(ctx context.Context, id string) error {
	_, err := rsc.Get(ctx, id)
	if err != nil {
		return err
	}

	return rsc.storage.Delete(ctx, id)
}
