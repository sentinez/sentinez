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
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sentinez/sentinez/pkg/client/consul"
	"github.com/sentinez/sentinez/pkg/client/names"
	"github.com/sentinez/sentinez/pkg/client/options"
	"github.com/sentinez/sentinez/pkg/client/resolver"
)

var (
	dcvr *Discovery
	once sync.Once
)

func GetDiscovery(opt *options.Options) *Discovery {
	once.Do(func() {
		dcvr = New(opt)
	})

	return dcvr
}

// New create new instance
func New(opt *options.Options) *Discovery {
	csClient, _ := consul.New(opt.ConsulURL)
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

type RegisterRequest struct {
	Name    string
	Address string
	Port    int
	TTL     time.Duration
}

// Register registers the service to the service registry.
func (dcv *Discovery) Register(
	req *RegisterRequest) (serviceID string, err error) {

	id := uuid.New().String()
	err = dcv.client.RegisterWithTTL(id,
		req.Name,
		req.Address,
		req.Port,
		req.TTL,
	)

	if err != nil {
		return "", err
	}

	return id, nil
}

// Heartbeat is used to send heartbeat to the service registry.
func (dcv *Discovery) Heartbeat(serviceID string) error {
	if err := dcv.client.SendHeartbeat(serviceID); err != nil {
		return err
	}

	return nil
}

type DiscoverResponse struct {
	Address string
	Name    string
}

// Discover used to discover the service registry.
func (dcv *Discovery) Discover(
	serviceName names.Namespace) (*DiscoverResponse, error) {

	ans, err := dcv.resolver.PickInstance(serviceName.String())
	if err != nil {
		return nil, err
	}

	return &DiscoverResponse{
		Address: fmt.Sprintf("%s:%d", ans.Address, ans.Port),
		Name:    ans.Name,
	}, nil
}
