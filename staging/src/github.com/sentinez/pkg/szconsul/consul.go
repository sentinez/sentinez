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

package szconsul

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/sentinez/pkg/szuri"
)

type Instance struct {
	Name    string
	Address string
	Port    int
}

type Client struct {
	consul *api.Client
}

// New returns the client of the Consul agent of consulURL (SZ_CONSUL_URI),
// "http://host:port", "https://host:port" or "unix:///path"; empty takes
// the agent of the CONSUL_HTTP_ADDR env, else 127.0.0.1:8500. https also
// reads the CONSUL_CACERT, CONSUL_CLIENT_CERT and CONSUL_CLIENT_KEY env.
func New(consulURL string) (*Client, error) {
	if err := checkURL(consulURL); err != nil {
		return nil, err
	}

	conf := api.DefaultConfig()
	if consulURL != "" {
		conf.Address = consulURL
	}

	c, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}
	return &Client{consul: c}, nil
}

func (c *Client) Register(
	id, name, address string,
	port int,
	healthPath string,
) error {
	url := fmt.Sprintf("http://%s:%d%s", address, port, healthPath)
	check := &api.AgentServiceCheck{
		HTTP:                           url,
		Interval:                       "10s",
		Timeout:                        "2s",
		DeregisterCriticalServiceAfter: "1m",
	}

	reg := &api.AgentServiceRegistration{
		ID:      id,
		Name:    name,
		Address: address,
		Port:    port,
		Check:   check,
	}

	return c.consul.Agent().ServiceRegister(reg)
}

func (c *Client) RegisterWithTTL(
	id, name, address string,
	port int,
	ttl time.Duration,
) error {
	check := &api.AgentServiceCheck{
		TTL:                            ttl.String(),
		DeregisterCriticalServiceAfter: "1m",
	}

	reg := &api.AgentServiceRegistration{
		ID:      id,
		Name:    name,
		Address: address,
		Port:    port,
		Check:   check,
	}

	return c.consul.Agent().ServiceRegister(reg)
}

func (c *Client) SendHeartbeat(checkId string) error {
	return c.consul.Agent().PassTTL("service:"+checkId, "alive")
}

func (c *Client) Discover(name string) ([]Instance, error) {
	services, _, err := c.consul.Health().Service(name, "", true, nil)
	if err != nil {
		return nil, err
	}

	var result []Instance
	for _, entry := range services {
		svc := entry.Service
		result = append(result, Instance{
			Name:    svc.Service,
			Address: svc.Address,
			Port:    svc.Port,
		})
	}
	return result, nil
}

// checkURL fails for a consulURL without a scheme Consul takes, e.g. a bare
// "host:port".
func checkURL(consulURL string) error {
	if consulURL == "" {
		return nil
	}

	u, err := szuri.Parse(consulURL)
	if err != nil {
		return fmt.Errorf("consul uri: %w", err)
	}

	if err := u.Want("http", "https", "unix"); err != nil {
		return fmt.Errorf("consul uri: %w", err)
	}

	if u.Scheme == "unix" {
		return nil
	}

	if _, err := u.Host(); err != nil {
		return fmt.Errorf("consul uri: %w", err)
	}

	return nil
}
