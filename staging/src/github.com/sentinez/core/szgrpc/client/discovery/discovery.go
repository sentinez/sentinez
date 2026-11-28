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
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sentinez/core/szgrpc/client/resolver"
	"github.com/sentinez/pkg/szconsul"
)

var (
	_dcvr    *Discovery
	_dcvrErr error
	_once    sync.Once
)

// GetDiscovery returns the Discovery of the process, made by New with the
// opts of the first call; later opts are ignored, an error is that of New.
func GetDiscovery(opts ...Option) (*Discovery, error) {
	_once.Do(func() {
		_dcvr, _dcvrErr = New(opts...)
	})

	return _dcvr, _dcvrErr
}

// New returns the Discovery of the consul of opts, an error when its URL
// is invalid (see szconsul.New).
func New(opts ...Option) (*Discovery, error) {
	o := newOptions(opts)

	csClient, err := szconsul.New(o.consulURL)
	if err != nil {
		return nil, fmt.Errorf("discovery: %w", err)
	}

	return &Discovery{
		client:   csClient,
		resolver: resolver.New(csClient),
	}, nil
}

// Discovery is a service registry for the sentinez.
type Discovery struct {
	client   *szconsul.Client
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
	req *RegisterRequest) (serviceId string, err error) {

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
func (dcv *Discovery) Heartbeat(serviceId string) error {
	if err := dcv.client.SendHeartbeat(serviceId); err != nil {
		return err
	}

	return nil
}

// Instances returns the host:port of the healthy instances of serviceKey,
// looked up in consul now, an error when there is none.
func (dcv *Discovery) Instances(serviceKey string) ([]string, error) {
	instances, err := dcv.client.Discover(serviceKey)
	if err != nil {
		return nil, fmt.Errorf("discover %s: %w", serviceKey, err)
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("discover %s: no instance", serviceKey)
	}

	addrs := make([]string, 0, len(instances))
	for _, inst := range instances {
		addrs = append(addrs, hostPort(inst))
	}

	return addrs, nil
}

type DiscoverResponse struct {
	Address string
	Name    string
}

// Discover used to discover the service registry.
func (dcv *Discovery) Discover(
	serviceKey string) (*DiscoverResponse, error) {

	ans, err := dcv.resolver.PickInstance(serviceKey)
	if err != nil {
		return nil, err
	}

	return &DiscoverResponse{
		Address: hostPort(*ans),
		Name:    ans.Name,
	}, nil
}

// hostPort returns the address of inst, IPv6 hosts bracketed.
func hostPort(inst szconsul.Instance) string {
	return net.JoinHostPort(inst.Address, strconv.Itoa(inst.Port))
}
