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
package discovery

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	discoverypb "github.com/sentinez/sentinez/api/gen/go/sentinez/common/discovery/v1"
	"github.com/sentinez/sentinez/client/consul"
	"github.com/sentinez/sentinez/client/resolver"

	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	dcvr *Discovery
	once sync.Once
)

func GetDiscovery(consulURL string) *Discovery {
	once.Do(func() {
		dcvr = New(consulURL)
	})

	return dcvr
}

// New create new instance
func New(consulAddr string) *Discovery {
	csClient, _ := consul.New(consulAddr)
	return &Discovery{
		client:   csClient,
		resolver: resolver.New(csClient),
	}
}

// Discovery is a service registry for the sentinez.
type Discovery struct {
	client   *consul.Client
	resolver *resolver.Resolver
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

	ans, err := dcv.resolver.PickInstance(req.GetName())
	if err != nil {
		return nil, err
	}

	return &discoverypb.DiscoverResponse{
		Name:    req.GetName(),
		Address: fmt.Sprintf("%s:%d", ans.Address, ans.Port),
	}, nil
}
