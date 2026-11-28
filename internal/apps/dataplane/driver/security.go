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

package driver

import (
	"encoding/binary"
	"fmt"
	"net"

	sentinezbpf "github.com/sentinez/bpf/sentinez"
)

// BlockCIDR adds cidr to the XDP blocklist. The key ip is kept in network
// byte order in memory, as XDP looks it up with the raw saddr.
func (ctx *Context) BlockCIDR(cidr string) error {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("driver: %w", err)
	}

	ip4 := ipNet.IP.To4()
	if ip4 == nil {
		return fmt.Errorf("driver: %s: %w", cidr, errNotIPv4)
	}

	ones, _ := ipNet.Mask.Size()
	key := sentinezbpf.SzIpLpmKey{
		Prefixlen: uint32(ones),
		Ip:        binary.NativeEndian.Uint32(ip4),
	}

	if err := ctx.obj.Blocklist.Put(key, uint8(1)); err != nil {
		return fmt.Errorf("driver: block %s: %w", cidr, err)
	}

	return nil
}
