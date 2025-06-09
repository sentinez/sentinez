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

// Package dcvrrepo implements the discovery repository
package dcvrrepo

import (
	"context"

	"github.com/sentinez/sentinez/pkg/common/utils"

	discoverypb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/discovery/v1"
	"github.com/sentinez/sentinez/pkg/common/decor"
	"github.com/sentinez/sentinez/pkg/infra/cache/mem"
	"github.com/sentinez/sentinez/pkg/std/errors"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

// New creates a new discovery repository
func New() Discovery {
	return &discovery{
		cache: mem.NewDefault[*discoverypb.Registrar](),
	}
}

// Discovery is a discovery repository
type Discovery interface {
	List(ctx context.Context,
		req *discoverypb.DiscoverRequest) ([]*discoverypb.Registrar, error)
	Get(ctx context.Context, id string) (*discoverypb.Registrar, error)
	Delete(ctx context.Context, ids ...string) error
	Save(ctx context.Context, req *discoverypb.Registrar) error
}

type discovery struct {
	cache *mem.Cache[*discoverypb.Registrar]
}

func (d *discovery) List(ctx context.Context,
	req *discoverypb.DiscoverRequest) ([]*discoverypb.Registrar, error) {

	return decor.WithContextReturn(ctx,
		func() ([]*discoverypb.Registrar, error) {
			resp, ok := d.cache.List()
			if !ok {
				zlog.Errorf("discovery.List: cannot list object in registrar")
				return nil, errors.ErrInvalidData
			}

			resp = utils.Filter(resp, func(r *discoverypb.Registrar) bool {
				return r.GetName() == req.GetName()
			})

			return resp, nil
		})
}

func (d *discovery) Get(ctx context.Context,
	id string) (*discoverypb.Registrar, error) {

	return decor.WithContextReturn(ctx,
		func() (*discoverypb.Registrar, error) {
			resp, ok := d.cache.Get(id)
			if !ok {
				zlog.Errorf("discovery.Get: cannot find object in registrar")
				return nil, errors.ErrNotFound
			}

			return resp, nil
		})
}

func (d *discovery) Delete(ctx context.Context, ids ...string) error {
	return decor.WithContext(ctx, func() error {
		for _, id := range ids {
			d.cache.Del(id)
		}

		return nil
	})
}

func (d *discovery) Save(ctx context.Context,
	req *discoverypb.Registrar) error {

	return decor.WithContext(ctx, func() error {
		d.cache.SetWithTTL(req.GetId(), req, req.GetTtl().AsDuration())
		return nil
	})
}
