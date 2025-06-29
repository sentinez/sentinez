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
package dischdl

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	discoverypb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/discovery/v1"
	consulclient "github.com/sentinez/sentinez/pkg/std/client/consul"

	"google.golang.org/protobuf/types/known/emptypb"
)

var _ discoverypb.DiscoveryServiceServer = (*Discovery)(nil)

// New create new instance
func New() discoverypb.DiscoveryServiceServer {
	client, _ := consulclient.New("http://localhost:8500")
	return &Discovery{
		client: client,
	}
}

// Discovery is a service registry for the sentinez.
type Discovery struct {
	discoverypb.UnimplementedDiscoveryServiceServer
	client *consulclient.Client
}

// Register registers the service to the service registry.
func (dcv *Discovery) Register(_ context.Context,
	req *discoverypb.RegisterRequest) (*discoverypb.RegisterResponse, error) {

	id := uuid.New().String()
	err := dcv.client.RegisterWithTTL(id,
		req.GetName(),
		req.GetAddress(),
		int(req.GetPort()),
		req.GetTtl().AsDuration(),
	)

	if err != nil {
		return nil, err
	}

	return &discoverypb.RegisterResponse{
		Id:      id,
		Name:    req.GetName(),
		Address: req.GetAddress(),
		Port:    req.GetPort(),
	}, nil
}

// Heartbeat is used to send heartbeat to the service registry.
func (dcv *Discovery) Heartbeat(_ context.Context,
	request *discoverypb.HeartbeatRequest) (*emptypb.Empty, error) {

	if err := dcv.client.SendHeartbeat(request.GetId()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// Discover used to discover the service registry.
func (dcv *Discovery) Discover(_ context.Context,
	req *discoverypb.DiscoverRequest) (*discoverypb.DiscoverResponse, error) {

	insts, err := dcv.client.Discover(req.GetName())
	if err != nil {
		return nil, err
	}

	if len(insts) == 0 {
		return nil, fmt.Errorf("service %s unavailable", req.GetName())
	}

	ans := insts[rand.Intn(len(insts))]

	return &discoverypb.DiscoverResponse{
		Name:    req.GetName(),
		Address: fmt.Sprintf("%s:%d", ans.Address, ans.Port),
	}, nil
}
