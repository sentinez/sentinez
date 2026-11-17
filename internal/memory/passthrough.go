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
	"net"

	"github.com/sentinez/pkg/sznormalize"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/sentinez/sentinez/pkg/network"
)

const (
	_passthrough = edgepb.TlsMode_TLS_MODE_PASSTHROUGH
	// _passthroughPort the port of the upstreams a passthrough connection
	// goes to: the one of HTTPS, as the upstream servers carry none.
	_passthroughPort = "443"
)

var _ network.PassthroughFunc = (*MemStore)(nil).Passthrough

// Passthrough implements network.PassthroughFunc: the upstreams of the
// domain serverName, on port 443, when its TLS mode is passthrough. A
// domain without upstreams is not passed through: its handshake fails at
// the edge, which holds no certificate of it.
func (m *MemStore) Passthrough(serverName string) ([]string, bool) {
	addrs, ok := m.passthrough.Load(sznormalize.ResourceDomain(serverName))

	return addrs, ok && len(addrs) != 0
}

// passthroughAddrs the upstreams of res on port 443, each once, in the
// order of its locations.
func passthroughAddrs(res *edgepb.Resource) []string {
	upstreams := upstreamServers(res)
	addrs := make([]string, 0, len(upstreams))
	for _, upstream := range upstreams {
		addrs = append(addrs,
			net.JoinHostPort(upstream.GetServer(), _passthroughPort))
	}

	return addrs
}
