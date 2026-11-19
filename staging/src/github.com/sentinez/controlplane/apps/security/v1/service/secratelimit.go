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

package securitysvc

import (
	"context"
	"errors"
	"time"

	"github.com/sentinez/controlplane/pkg/auditchange"
	"github.com/sentinez/controlplane/pkg/authz"
	"github.com/sentinez/pkg/szerror"
	"github.com/sentinez/pkg/szprotobuf"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	"google.golang.org/protobuf/proto"
)

var errNonPositiveDuration = errors.New("must be greater than 0")

// mapRateLimit converts a DB SecRateLimit into an API SecRateLimit response.
func (srv *SecurityService) mapRateLimit(_ context.Context,
	dbRL *securitypb.SecRateLimit,
) *securitypb.SecRateLimit {
	if dbRL == nil {
		return nil
	}

	return &securitypb.SecRateLimit{
		Id:          dbRL.GetId(),
		Name:        dbRL.GetName(),
		Description: dbRL.GetDescription(),
		Expr:        dbRL.GetExpr(),
		Action:      dbRL.GetAction(),
		Status:      dbRL.GetStatus(),
		Priority:    dbRL.GetPriority(),
		TimeWindow:  dbRL.GetTimeWindow(),
		MaxRequests: dbRL.GetMaxRequests(),
		Timeout:     dbRL.GetTimeout(),
		Owner:       dbRL.GetOwner(),
	}
}

// ── SecRateLimit ──────────────────────────────────────────────────────────

// CreateSecRateLimit creates a new rate limit rule
func (srv *SecurityService) CreateSecRateLimit(ctx context.Context,
	req *securitypb.CreateSecRateLimitRequest,
) (*securitypb.CreateSecRateLimitResponse, error) {
	if err := validateRateLimit(req.GetSecratelimit()); err != nil {
		return nil, err
	}

	owner, err := srv.newOwner(ctx, req.GetSecratelimit().GetOwner())
	if err != nil {
		return nil, err
	}

	req.GetSecratelimit().Owner = owner

	created, err := srv.secRateLimitRepo.Create(ctx, req.GetSecratelimit())
	if err != nil {
		return nil, err
	}

	return &securitypb.CreateSecRateLimitResponse{Id: created.GetId()}, nil
}

func (srv *SecurityService) GetSecRateLimit(ctx context.Context,
	req *securitypb.GetSecRateLimitRequest,
) (*securitypb.GetSecRateLimitResponse, error) {
	dbRL, err := owned(ctx, srv, srv.secRateLimitRepo.Get, "rate limit",
		req.GetId(), authz.Read)
	if err != nil {
		return nil, err
	}

	return &securitypb.
		GetSecRateLimitResponse{Secratelimit: srv.mapRateLimit(ctx, dbRL)}, nil
}

// _rateLimitUpdatable the fields of a rate limit an update writes.
var _rateLimitUpdatable = []string{
	securitypb.SecRateLimit_Name, securitypb.SecRateLimit_Description,
	securitypb.SecRateLimit_Expr, securitypb.SecRateLimit_Status,
	securitypb.SecRateLimit_Priority, securitypb.SecRateLimit_Action,
	securitypb.SecRateLimit_TimeWindow, securitypb.SecRateLimit_MaxRequests,
	securitypb.SecRateLimit_Timeout,
}

// UpdateSecRateLimit updates the fields of a rate limit rule its
// update_mask names; the rule they make must be valid.
func (srv *SecurityService) UpdateSecRateLimit(ctx context.Context,
	req *securitypb.UpdateSecRateLimitRequest,
) (*securitypb.UpdateSecRateLimitResponse, error) {
	if req.GetSecratelimit() == nil {
		return nil, szerror.StatusInvalidArgumentF("secratelimit is required")
	}

	dbRL, err := owned(ctx, srv, srv.secRateLimitRepo.Get, "rate limit",
		req.GetId(), authz.Write)
	if err != nil {
		return nil, err
	}

	before := proto.CloneOf(dbRL)
	paths, err := szprotobuf.ApplyMask(dbRL, req.GetSecratelimit(),
		req.GetUpdateMask(), _rateLimitUpdatable...)
	if err != nil {
		return nil, err
	}

	if err = validateRateLimit(dbRL); err != nil {
		return nil, err
	}

	if err = srv.secRateLimitRepo.Update(ctx, dbRL, paths); err != nil {
		return nil, err
	}

	auditchange.Set(ctx, before, dbRL, paths)

	return &securitypb.
		UpdateSecRateLimitResponse{
		Secratelimit: srv.mapRateLimit(ctx, dbRL),
	}, nil
}

func (srv *SecurityService) DeleteSecRateLimit(ctx context.Context,
	req *securitypb.DeleteSecRateLimitRequest,
) (*securitypb.DeleteSecRateLimitResponse, error) {
	_, err := owned(ctx, srv, srv.secRateLimitRepo.Get, "rate limit",
		req.GetId(), authz.Write)
	if err != nil {
		return nil, err
	}

	if err = srv.secRateLimitRepo.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &securitypb.DeleteSecRateLimitResponse{}, nil
}

func (srv *SecurityService) ListSecRateLimits(ctx context.Context,
	req *securitypb.ListSecRateLimitsRequest,
) (*securitypb.ListSecRateLimitsResponse, error) {
	req, ok, err := authz.Scoped(ctx, srv.authz, req)
	if err != nil || !ok {
		return &securitypb.ListSecRateLimitsResponse{}, err
	}

	dbRLs, total, err := srv.secRateLimitRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	rls := make([]*securitypb.SecRateLimit, 0, len(dbRLs))
	for _, dbRL := range dbRLs {
		rls = append(rls, srv.mapRateLimit(ctx, dbRL))
	}

	return &securitypb.ListSecRateLimitsResponse{
		Secratelimits: rls,
		Total:         total,
	}, nil
}

// validateRateLimit checks the limiter settings the edge relies on.
func validateRateLimit(rl *securitypb.SecRateLimit) error {
	if rl == nil {
		return szerror.StatusInvalidArgumentF("secratelimit is required")
	}
	if err := validateExpr(rl.GetExpr()); err != nil {
		return err
	}
	if rl.GetMaxRequests() <= 0 {
		return szerror.StatusInvalidArgumentF(
			"max_requests must be greater than 0")
	}
	if err := validatePositiveDuration(rl.GetTimeWindow()); err != nil {
		return szerror.StatusInvalidArgumentF("time_window: %v", err)
	}
	if rl.GetTimeout() == "" {
		return nil
	}
	if err := validatePositiveDuration(rl.GetTimeout()); err != nil {
		return szerror.StatusInvalidArgumentF("timeout: %v", err)
	}
	return nil
}

func validatePositiveDuration(s string) error {
	d, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	if d <= 0 {
		return errNonPositiveDuration
	}
	return nil
}
