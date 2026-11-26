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
	"crypto/x509"
	"testing"
	"time"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/pkg/szseal"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	_upstreamNs     = "mtls.example.com"
	_upstreamServer = "origin.example.com"
)

// acceptServer an szhttp.Server keeping the TLS config of each proxy.
type acceptServer struct {
	szhttp.Server
	confs map[string]*tls.Config
}

func (s *acceptServer) AcceptReverse(target string,
	conf *tls.Config) (szhttp.ReverseProxy, error) {
	s.confs[target] = conf
	return noProxy{}, nil
}

type noProxy struct{}

func (noProxy) Serve(szhttp.Context) {}

// upstreamSetting a setting of _upstreamNs proxying to _upstreamServer
// over https, with the upstream certificates t.
func upstreamSetting(t *edgepb.Tls) *edgepb.Setting {
	st := newSetting("tnt_1", _upstreamNs)
	st.Ingress.Tls = t
	st.Ingress.Resource.Locations = []*edgepb.Location{{
		Location: "/",
		ProxyPass: []*edgepb.Upstream{{
			Server:   _upstreamServer,
			Protocol: edgepb.ProxyProtocol_PROXY_PROTOCOL_HTTPS,
		}},
	}}

	return st
}

func upstreamStore(t *testing.T, sealer *szseal.Sealer) *MemStore {
	t.Helper()

	m := newSyncStore()
	m.sealer = sealer
	t.Cleanup(func() { m.revProxy.Delete(_upstreamNs) })

	return m
}

func TestLoadReverseProxyWithoutUpstreamTLS(t *testing.T) {
	m := upstreamStore(t, nil)
	srv := &acceptServer{confs: map[string]*tls.Config{}}

	require.NoError(t, m.LoadReverseProxy(srv, upstreamSetting(nil)))

	conf, ok := srv.confs["https://"+_upstreamServer]
	require.True(t, ok)
	assert.Nil(t, conf, "the defaults of the transport")
	_, ok = m.revProxy.Load(_upstreamNs, _upstreamServer)
	assert.True(t, ok)
}

func TestLoadReverseProxyDialsWithUpstreamTLS(t *testing.T) {
	sealer := newTestSealer(t)
	ca := issue(t, caTmpl(1, "upstream root"), nil)
	root, inter, leaf := chain(t, "edge.example.com",
		time.Now().Add(time.Hour))
	m := upstreamStore(t, sealer)
	srv := &acceptServer{confs: map[string]*tls.Config{}}

	err := m.LoadReverseProxy(srv, upstreamSetting(&edgepb.Tls{
		UpstreamCas: []*edgepb.Certificate{
			sealed(t, sealer, "cacrt_ca", ca, false),
		},
		UpstreamCertificates: []*edgepb.Certificate{
			sealed(t, sealer, "cacrt_root", root, false),
			sealed(t, sealer, "cacrt_inter", inter, false),
			sealed(t, sealer, "cacrt_leaf", leaf, true),
		},
	}))
	require.NoError(t, err)

	conf := srv.confs["https://"+_upstreamServer]
	require.NotNil(t, conf)
	want := x509.NewCertPool()
	want.AddCert(ca.cert)
	assert.True(t, want.Equal(conf.RootCAs), "only the upstream CA")
	require.Len(t, conf.Certificates, 1)
	assert.Equal(t, [][]byte{leaf.cert.Raw, inter.cert.Raw},
		conf.Certificates[0].Certificate, "leaf then intermediate, no root")
}

// failClosedCase upstream certificates the edge can not dial with.
type failClosedCase struct {
	name       string
	giveSealer *szseal.Sealer
	giveTls    *edgepb.Tls
}

func failClosedCases(t *testing.T, sealer *szseal.Sealer) []failClosedCase {
	t.Helper()

	ca := issue(t, caTmpl(1, "upstream root"), nil)
	_, _, expired := chain(t, "edge.example.com",
		time.Now().Add(-time.Hour))
	// sealed for another record: it does not open as cacrt_a
	moved := sealed(t, sealer, "cacrt_b", ca, false)
	moved.Id = "cacrt_a"
	cas := func(c ...*edgepb.Certificate) *edgepb.Tls {
		return &edgepb.Tls{UpstreamCas: c}
	}
	clients := func(c ...*edgepb.Certificate) *edgepb.Tls {
		return &edgepb.Tls{UpstreamCertificates: c}
	}

	return []failClosedCase{
		{"no data key", nil, cas(sealed(t, sealer, "cacrt_a", ca, false))},
		{"ca of another aad", sealer, cas(moved)},
		{"expired client certificate", sealer,
			clients(sealed(t, sealer, "cacrt_a", expired, true))},
		{"client certificate without key", sealer,
			clients(sealed(t, sealer, "cacrt_a", ca, false))},
	}
}

func TestLoadReverseProxyFailsClosed(t *testing.T) {
	tests := failClosedCases(t, newTestSealer(t))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := upstreamStore(t, tt.giveSealer)
			srv := &acceptServer{confs: map[string]*tls.Config{}}
			require.NoError(t, m.LoadReverseProxy(srv, upstreamSetting(nil)))

			err := m.LoadReverseProxy(srv, upstreamSetting(tt.giveTls))

			require.Error(t, err)
			_, ok := m.revProxy.Load(_upstreamNs, _upstreamServer)
			assert.False(t, ok, "the previous proxy is dropped")
		})
	}
}

func TestUnloadDropsReverseProxy(t *testing.T) {
	m := upstreamStore(t, nil)
	m.server = &acceptServer{confs: map[string]*tls.Config{}}

	m.load("rsrc_1", upstreamSetting(nil))
	_, ok := m.revProxy.Load(_upstreamNs, _upstreamServer)
	require.True(t, ok)

	m.unload("rsrc_1", _upstreamNs)

	_, ok = m.revProxy.Load(_upstreamNs, _upstreamServer)
	assert.False(t, ok)
}
