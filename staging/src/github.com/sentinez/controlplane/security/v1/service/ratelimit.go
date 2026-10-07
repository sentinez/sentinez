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

	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	"github.com/sentinez/shared/errorx"
)

var errNonPositiveDuration = errors.New("must be greater than 0")

// mapRateLimit converts a DB RateLimit into an API RateLimit response.
func (srv *SecurityService) mapRateLimit(_ context.Context,
	dbRL *securitypb.RateLimit,
) *securitypb.RateLimit {
	if dbRL == nil {
		return nil
	}

	return &securitypb.RateLimit{
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
		Metadata:    dbRL.GetMetadata(),
	}
}

// ── RateLimit ──────────────────────────────────────────────────────────

// CreateRateLimit creates a new rate limit rule
func (srv *SecurityService) CreateRateLimit(ctx context.Context,
	req *securitypb.CreateRateLimitRequest,
) (*securitypb.CreateRateLimitResponse, error) {
	if err := validateRateLimit(req.GetRateLimit()); err != nil {
		return nil, err
	}

	created, err := srv.rateLimitRepo.Create(ctx, req.GetRateLimit())
	if err != nil {
		return nil, err
	}

	return &securitypb.CreateRateLimitResponse{Id: created.GetId()}, nil
}

func (srv *SecurityService) GetRateLimit(ctx context.Context,
	req *securitypb.GetRateLimitRequest,
) (*securitypb.GetRateLimitResponse, error) {
	dbRL, err := srv.rateLimitRepo.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &securitypb.
		GetRateLimitResponse{RateLimit: srv.mapRateLimit(ctx, dbRL)}, nil
}

// UpdateRateLimit updates an existing rate limit rule
func (srv *SecurityService) UpdateRateLimit(ctx context.Context,
	req *securitypb.UpdateRateLimitRequest,
) (*securitypb.UpdateRateLimitResponse, error) {
	rl := req.GetRateLimit()
	if err := validateRateLimit(rl); err != nil {
		return nil, err
	}

	dbRL, err := srv.rateLimitRepo.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	dbRL.Name = rl.GetName()
	dbRL.Description = rl.GetDescription()
	dbRL.Expr = rl.GetExpr()
	dbRL.Status = rl.GetStatus()
	dbRL.Priority = rl.GetPriority()
	dbRL.Action = rl.GetAction()
	dbRL.TimeWindow = rl.GetTimeWindow()
	dbRL.MaxRequests = rl.GetMaxRequests()
	dbRL.Timeout = rl.GetTimeout()

	if err := srv.rateLimitRepo.Update(ctx, dbRL); err != nil {
		return nil, err
	}

	return &securitypb.
		UpdateRateLimitResponse{RateLimit: srv.mapRateLimit(ctx, dbRL)}, nil
}

func (srv *SecurityService) DeleteRateLimit(ctx context.Context,
	req *securitypb.DeleteRateLimitRequest,
) (*securitypb.DeleteRateLimitResponse, error) {
	if err := srv.rateLimitRepo.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &securitypb.DeleteRateLimitResponse{}, nil
}

func (srv *SecurityService) ListRateLimits(ctx context.Context,
	req *securitypb.ListRateLimitsRequest,
) (*securitypb.ListRateLimitsResponse, error) {
	dbRLs, total, err := srv.rateLimitRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	rls := make([]*securitypb.RateLimit, 0, len(dbRLs))
	for _, dbRL := range dbRLs {
		rls = append(rls, srv.mapRateLimit(ctx, dbRL))
	}

	return &securitypb.ListRateLimitsResponse{
		RateLimits: rls,
		Total:      total,
	}, nil
}

// validateRateLimit checks the limiter settings the edge relies on.
func validateRateLimit(rl *securitypb.RateLimit) error {
	if rl == nil {
		return errorx.StatusInvalidArgumentF("rate_limit is required")
	}
	if rl.GetMaxRequests() <= 0 {
		return errorx.StatusInvalidArgumentF(
			"max_requests must be greater than 0")
	}
	if err := validatePositiveDuration(rl.GetTimeWindow()); err != nil {
		return errorx.StatusInvalidArgumentF("time_window: %v", err)
	}
	if rl.GetTimeout() == "" {
		return nil
	}
	if err := validatePositiveDuration(rl.GetTimeout()); err != nil {
		return errorx.StatusInvalidArgumentF("timeout: %v", err)
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
