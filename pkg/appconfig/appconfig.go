// Copyright 2026 Duc-Hung Ho.
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

// Package appconfig builds the config of a service: its flags, and its env
// from the environment (.env included) with the values Consul adds to it.
//
// A service only declares a Spec and calls Provider.Config. In Consul, the
// env lives under KVPrefix: a key shared by all the services, or one in the
// folder of a service (its service_key) that wins over the shared one. The
// environment wins over both. Values are read once, when the service starts.
package appconfig

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/sentinez/pkg/szconfig"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/sentinez/api/client/consul"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

// _consulTimeout bounds the read of the env from Consul, which runs before
// the service starts.
const _consulTimeout = 5 * time.Second

// Spec is what a service declares about its config.
type Spec struct {
	// Meta is the service; its service_key is its folder in Consul.
	Meta *typepb.XMeta
	// Flags parses the flags of the service.
	Flags func() (*settingpb.Flag, error)
	// Keys are the env keys the service reads; none means all of them.
	Keys []settingpb.Senz
}

// Provider loads the config of a service once.
type Provider struct {
	spec Spec
	// fetch reads the env of Consul at uri, keyed relative to KVPrefix.
	fetch func(ctx context.Context, uri string) (map[string]string, error)

	once sync.Once
	conf *settingpb.Config
	err  error
}

// New returns the Provider of the service of spec.
func New(spec Spec) *Provider {
	return &Provider{spec: spec, fetch: fetchConsul}
}

// Config loads the config on the first call and returns it every time.
func (p *Provider) Config() (*settingpb.Config, error) {
	p.once.Do(func() { p.conf, p.err = p.load() })

	return p.conf, p.err
}

func (p *Provider) load() (*settingpb.Config, error) {
	flag, err := p.spec.Flags()
	if err != nil {
		return nil, err
	}

	env, err := szconfig.LoadEnv(flag.GetEnvFile(), p.envKeys()...)
	if err != nil {
		return nil, err
	}

	if err = p.addConsul(env); err != nil {
		return nil, err
	}

	return &settingpb.Config{Meta: p.spec.Meta, Flag: flag, Env: env}, nil
}

// envKeys are the keys to load from the environment: those of the spec, and
// always SENZ_CONSUL_URI, or a service that lists its keys would never reach
// Consul. None means all of them.
func (p *Provider) envKeys() []settingpb.Senz {
	if len(p.spec.Keys) == 0 {
		return nil
	}

	keys := slices.Clone(p.spec.Keys)
	if !slices.Contains(keys, settingpb.Senz_SENZ_CONSUL_URI) {
		keys = append(keys, settingpb.Senz_SENZ_CONSUL_URI)
	}

	return keys
}

// addConsul sets in env what Consul gives for the keys the environment
// leaves empty. Without SENZ_CONSUL_URI there is nothing to read.
func (p *Provider) addConsul(env map[string]string) error {
	uri := env[settingpb.Senz_SENZ_CONSUL_URI.String()]
	if uri == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), _consulTimeout)
	defer cancel()

	// a service that cannot reach its Consul does not start: its secrets
	// may live there only, and would be missing later and elsewhere
	kv, err := p.fetch(ctx, uri)
	if err != nil {
		return fmt.Errorf("env from consul %s: %w", uri, err)
	}

	key := p.spec.Meta.GetServiceKey()
	found, skipped := overlay(env, kv, key, p.spec.Keys)
	for _, path := range skipped {
		szlog.Warnf("consul env: %s%s ignored: unknown key, or one that "+
			"only the environment may set", KVPrefix, path)
	}

	names := make([]string, 0, len(found))
	for name, value := range found {
		// env is the map of szconfig: this is what GetEnv gives too
		szconfig.SetEnv(settingpb.Senz(settingpb.Senz_value[name]), value)
		names = append(names, name)
	}

	slices.Sort(names)
	szlog.Infof("consul env: %d keys from %s: %v", len(names), uri, names)

	return nil
}

func fetchConsul(ctx context.Context, uri string) (map[string]string, error) {
	client, err := consul.New(uri)
	if err != nil {
		return nil, err
	}

	return consul.List[string](ctx, client, KVPrefix)
}
