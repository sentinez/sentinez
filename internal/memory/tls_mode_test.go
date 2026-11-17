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

package memory

import (
	"crypto/tls"
	"testing"

	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// modeSetting the setting of upstreamSetting in TLS mode, its upstream
// of protocol.
func modeSetting(mode edgepb.TlsMode,
	protocol edgepb.ProxyProtocol) *edgepb.Setting {
	st := upstreamSetting(nil)
	st.Ingress.Resource.TlsMode = mode
	st.Ingress.Resource.Locations[0].ProxyPass[0].Protocol = protocol

	return st
}

func TestLoadReverseProxyTargetOfTlsMode(t *testing.T) {
	const (
		http  = edgepb.ProxyProtocol_PROXY_PROTOCOL_HTTP
		https = edgepb.ProxyProtocol_PROXY_PROTOCOL_HTTPS
	)

	tests := []struct {
		name         string
		giveMode     edgepb.TlsMode
		giveProtocol edgepb.ProxyProtocol
		wantTarget   string
	}{
		{"unspecified, http", edgepb.TlsMode_TLS_MODE_UNSPECIFIED, http,
			"http://" + _upstreamServer},
		{"unspecified, https", edgepb.TlsMode_TLS_MODE_UNSPECIFIED, https,
			"https://" + _upstreamServer},
		{"terminate", edgepb.TlsMode_TLS_MODE_TERMINATE, https,
			"http://" + _upstreamServer},
		{"re-encrypt", edgepb.TlsMode_TLS_MODE_REENCRYPT, http,
			"https://" + _upstreamServer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := upstreamStore(t, nil)
			srv := &acceptServer{confs: map[string]*tls.Config{}}

			err := m.LoadReverseProxy(srv,
				modeSetting(tt.giveMode, tt.giveProtocol))

			require.NoError(t, err)
			assert.Contains(t, srv.confs, tt.wantTarget)
			assert.Len(t, srv.confs, 1)
		})
	}
}

func TestLoadReverseProxyPassthroughHasNone(t *testing.T) {
	m := upstreamStore(t, nil)
	srv := &acceptServer{confs: map[string]*tls.Config{}}
	require.NoError(t, m.LoadReverseProxy(srv, upstreamSetting(nil)))

	err := m.LoadReverseProxy(srv, modeSetting(_passthrough,
		edgepb.ProxyProtocol_PROXY_PROTOCOL_HTTPS))

	require.NoError(t, err)
	_, ok := m.revProxy.Load(_upstreamNs, _upstreamServer)
	assert.False(t, ok, "the previous proxy is dropped")
}

// loadMode loads st into a fresh store through LoadReverseProxy.
func loadMode(t *testing.T, st *edgepb.Setting) *MemStore {
	t.Helper()

	m := upstreamStore(t, nil)
	srv := &acceptServer{confs: map[string]*tls.Config{}}
	require.NoError(t, m.LoadReverseProxy(srv, st))

	return m
}

func TestPassthroughResolvesUpstreams(t *testing.T) {
	st := modeSetting(_passthrough, edgepb.ProxyProtocol_PROXY_PROTOCOL_HTTPS)
	st.Ingress.Resource.Locations = append(st.Ingress.Resource.Locations,
		&edgepb.Location{Location: "/api", ProxyPass: []*edgepb.Upstream{
			{Server: _upstreamServer}, {Server: "10.0.0.2"}, {Server: ""},
		}})
	m := loadMode(t, st)

	got, ok := m.Passthrough("MTLS.example.com")

	require.True(t, ok)
	assert.Equal(t, []string{_upstreamServer + ":443", "10.0.0.2:443"}, got)
}

func TestPassthroughSkipsOtherDomains(t *testing.T) {
	tests := []struct {
		name     string
		giveMode edgepb.TlsMode
		giveName string
	}{
		{"terminate", edgepb.TlsMode_TLS_MODE_TERMINATE, _upstreamNs},
		{"re-encrypt", edgepb.TlsMode_TLS_MODE_REENCRYPT, _upstreamNs},
		{"unspecified", edgepb.TlsMode_TLS_MODE_UNSPECIFIED, _upstreamNs},
		{"unknown domain", _passthrough, "other.example.com"},
		{"no server name", _passthrough, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := loadMode(t, modeSetting(tt.giveMode,
				edgepb.ProxyProtocol_PROXY_PROTOCOL_HTTPS))

			_, ok := m.Passthrough(tt.giveName)

			assert.False(t, ok)
		})
	}
}

func TestPassthroughDroppedWhenModeChanges(t *testing.T) {
	m := loadMode(t, modeSetting(_passthrough,
		edgepb.ProxyProtocol_PROXY_PROTOCOL_HTTPS))
	srv := &acceptServer{confs: map[string]*tls.Config{}}

	require.NoError(t, m.LoadReverseProxy(srv, modeSetting(
		edgepb.TlsMode_TLS_MODE_REENCRYPT,
		edgepb.ProxyProtocol_PROXY_PROTOCOL_HTTPS)))

	_, ok := m.Passthrough(_upstreamNs)
	assert.False(t, ok)
}

func TestLoadReverseProxySharesPlainProxies(t *testing.T) {
	m := upstreamStore(t, nil)
	srv := &acceptServer{confs: map[string]*tls.Config{}}
	other := upstreamSetting(nil)
	other.Ingress.Resource.Domain = "other.example.com"
	t.Cleanup(func() { m.revProxy.Delete("other.example.com") })

	require.NoError(t, m.LoadReverseProxy(srv, upstreamSetting(nil)))
	clear(srv.confs)
	require.NoError(t, m.LoadReverseProxy(srv, upstreamSetting(nil)))
	require.NoError(t, m.LoadReverseProxy(srv, other))

	assert.Empty(t, srv.confs, "built once, then reused")
	_, ok := m.revProxy.Load("other.example.com", _upstreamServer)
	assert.True(t, ok)
}
