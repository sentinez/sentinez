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

package cluster

import (
	"errors"
	"fmt"
	"net"

	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szuri"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/internal/defaults"
)

// errNoPrivateIP is returned when a bind needs a private IP and the host
// has none: Olric would bind a public one in its place.
var errNoPrivateIP = errors.New("cluster: no private IP to bind, " +
	"set the host of SZ_OLRIC_ADDRESS and SZ_DISCOVERY_ADDRESS")

// privateIPFunc returns a private IP (RFC 6890) of the host, "" when it
// has none, e.g. sockaddr.GetPrivateIP.
type privateIPFunc func() (string, error)

// olricBind the Olric bind of SZ_OLRIC_ADDRESS, "host:port", where the
// other nodes and centraldata reach the node. Unset, an edge alone binds
// the loopback, out of reach, and an edge of a cluster a private IP.
func olricBind(appConf *settingpb.Config,
	privateIP privateIPFunc) (string, int, error) {
	addr := appConf.Get(settingpb.Sz_SZ_OLRIC_ADDRESS)
	if addr == "" && !appConf.Clustered() {
		return defaults.OlricLocalAddress, defaults.OlricPort, nil
	}

	if addr == "" {
		host, err := resolveHost("", privateIP)
		return host, defaults.OlricPort, err
	}

	host, port, err := bindOf(addr, privateIP)
	if err != nil {
		return "", 0, fmt.Errorf("cluster: SZ_OLRIC_ADDRESS %q: %w",
			addr, err)
	}

	return host, port, nil
}

// discoveryBind the memberlist bind of SZ_DISCOVERY_ADDRESS, "host:port";
// an invalid one takes the default port.
func discoveryBind(appConf *settingpb.Config,
	privateIP privateIPFunc) (string, int, error) {
	addr := appConf.Get(settingpb.Sz_SZ_DISCOVERY_ADDRESS)

	host, port, err := bindOf(addr, privateIP)
	if errors.Is(err, errNoPrivateIP) {
		return "", 0, err
	}

	if err != nil {
		szlog.Warnf("cluster: discovery address %q invalid, use port %d",
			addr, defaults.DiscoveryPort)

		host, err := resolveHost("", privateIP)

		return host, defaults.DiscoveryPort, err
	}

	return host, port, nil
}

// bindOf the host and port of addr, the host resolved by resolveHost.
func bindOf(addr string, privateIP privateIPFunc) (string, int, error) {
	host, rawPort, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, err
	}

	port, err := szuri.ParsePort(rawPort)
	if err != nil {
		return "", 0, err
	}

	host, err = resolveHost(host, privateIP)

	return host, port, err
}

// resolveHost the host to bind: an unspecified host ("", 0.0.0.0, ::)
// takes a private IP of the host. Olric would pick one too, but fall back
// to a public IP when there is none, exposing the port to anyone.
func resolveHost(host string, privateIP privateIPFunc) (string, error) {
	if !unspecified(host) {
		warnPublic(host)
		return host, nil
	}

	ip, err := privateIP()
	if err != nil {
		return "", fmt.Errorf("cluster: private ip: %w", err)
	}

	if ip == "" {
		return "", errNoPrivateIP
	}

	return ip, nil
}

// unspecified reports whether host binds every interface.
func unspecified(host string) bool {
	if host == "" {
		return true
	}

	ip := net.ParseIP(host)

	return ip != nil && ip.IsUnspecified()
}

// warnPublic warns of a public IP: what is set is bound, but the port of
// Olric and memberlist has no TLS nor password.
func warnPublic(host string) {
	ip := net.ParseIP(host)
	if ip == nil || ip.IsPrivate() || ip.IsLoopback() {
		return
	}

	szlog.Warnf("cluster: %s is a public IP: keep its port behind a "+
		"firewall, olric and memberlist have no TLS", host)
}
