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

// Package discovery provides a client for the discovery service.
package discovery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	discoverypb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/discovery/v1"
)

// Discovery defines the interface for the discovery service client.
type Discovery interface {
	Heartbeat(ctx context.Context, req *discoverypb.HeartbeatRequest) error
	Register(ctx context.Context, req *discoverypb.RegisterRequest) error
	Discover(ctx context.Context,
		req *discoverypb.DiscoverRequest) (*discoverypb.DiscoverResponse, error)
}

// New creates a new Discovery client with the given base URL.
func New(baseURL string) Discovery {
	return &discovery{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

type discovery struct {
	baseURL string
	client  *http.Client
}

func (d *discovery) Heartbeat(
	_ context.Context, req *discoverypb.HeartbeatRequest) error {

	return d.post("/discovery/heartbeat", req)
}

func (d *discovery) Register(
	_ context.Context, req *discoverypb.RegisterRequest) error {

	return d.post("/discovery/register", RegisterRequest{
		Name:    req.Name,
		Address: req.Address,
		TTL:     fmt.Sprintf("%ds", req.GetTtl().Seconds),
	})

}

func (d *discovery) Discover(_ context.Context,
	req *discoverypb.DiscoverRequest) (*discoverypb.DiscoverResponse, error) {

	resp, err := d.client.Get(
		fmt.Sprintf("%s/discovery/discover?name=%s", d.baseURL, req.GetName()))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var result discoverypb.DiscoverResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	return &result, err
}

func (d *discovery) post(path string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := d.client.Post(
		d.baseURL+path, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error: %s", resp.Status)
	}
	return nil
}
