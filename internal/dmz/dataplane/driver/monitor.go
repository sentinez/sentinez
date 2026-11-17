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
	"errors"
	"fmt"
	"net"

	"github.com/cilium/ebpf"
)

// errNotIPv4 an address the maps, keyed by IPv4, cannot hold.
var errNotIPv4 = errors.New("not an IPv4 address")

// LookupBandwidth the bytes ip sent, summed over the CPUs; 0 when the XDP
// program saw none.
func (ctx *Context) LookupBandwidth(ip string) (uint64, error) {
	ip4 := net.ParseIP(ip).To4()
	if ip4 == nil {
		return 0, fmt.Errorf("driver: %s: %w", ip, errNotIPv4)
	}

	// a per-CPU map holds a value for every possible CPU, not only the
	// online ones
	cpus, err := ebpf.PossibleCPU()
	if err != nil {
		return 0, fmt.Errorf("driver: possible cpus: %w", err)
	}

	key := binary.NativeEndian.Uint32(ip4)
	values := make([]uint64, cpus)

	err = ctx.obj.IpBandwidth.Lookup(&key, &values)
	if errors.Is(err, ebpf.ErrKeyNotExist) {
		return 0, nil
	}

	if err != nil {
		return 0, fmt.Errorf("driver: bandwidth of %s: %w", ip, err)
	}

	var total uint64
	for _, v := range values {
		total += v
	}

	return total, nil
}
