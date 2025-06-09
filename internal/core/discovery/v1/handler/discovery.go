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

// Package dcvrhandler provides a service discovery for the sentinez.
package dcvrhandler

import (
	"context"

	discoverypb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/discovery/v1"
	dcvrdomain "github.com/sentinez/sentinez/internal/core/discovery/v1/domain"
	"github.com/sentinez/sentinez/pkg/common/uuid"
	"github.com/sentinez/sentinez/pkg/std/errors"
	"github.com/sentinez/sentinez/pkg/std/zlog"

	"google.golang.org/protobuf/types/known/emptypb"
)

var _ discoverypb.DiscoveryServiceServer = (*Discovery)(nil)

// New create new instance
func New(svc dcvrdomain.Discovery) discoverypb.DiscoveryServiceServer {
	return &Discovery{
		service: svc,
	}
}

// Discovery is a service registry for the sentinez.
type Discovery struct {
	discoverypb.UnimplementedDiscoveryServiceServer
	service dcvrdomain.Discovery
}

// Register registers the service to the service registry.
func (dcv *Discovery) Register(ctx context.Context,
	req *discoverypb.RegisterRequest) (*discoverypb.RegisterResponse, error) {
	id := uuid.Generate()

	if err := dcv.service.RegisterService(ctx, id, req); err != nil {
		zlog.Errorf("discovery.Register service error: %s", err)
		return nil, errors.StatusInternalError
	}

	return &discoverypb.RegisterResponse{
		Id:      id,
		Name:    req.GetName(),
		Address: req.GetAddress(),
	}, nil
}

// Heartbeat is used to send heartbeat to the service registry.
func (dcv *Discovery) Heartbeat(ctx context.Context,
	request *discoverypb.HeartbeatRequest) (*emptypb.Empty, error) {
	_ = ctx
	_ = request
	//TODO implement me
	panic("implement me")
}

// Discover used to discover the service registry.
func (dcv *Discovery) Discover(ctx context.Context,
	req *discoverypb.DiscoverRequest) (*discoverypb.DiscoverResponse, error) {
	_ = ctx
	_ = req
	return &discoverypb.DiscoverResponse{
		Name: req.GetName(),
	}, nil
}
