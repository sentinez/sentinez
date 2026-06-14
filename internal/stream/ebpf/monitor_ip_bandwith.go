// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package streamebpf

import (
	"encoding/binary"
	"fmt"
	"net"
	"runtime"

	"github.com/cilium/ebpf"
	sentinezbpf "github.com/sentinez/sentinez/pkg/dmz/bpf/sentinez"
)

func ipToUint32(ipStr string) (uint32, error) {
	ip := net.ParseIP(ipStr).To4()
	if ip == nil {
		return 0, fmt.Errorf("invalid IPv4 address")
	}
	return binary.LittleEndian.Uint32(ip), nil
}

func LookupBandwidth(ip string) (uint64, error) {
	return Lookup(func(eo *sentinezbpf.SenzObjects) (uint64, error) {
		ipUint, err := ipToUint32(ip)
		if err != nil {
			return 0, err
		}

		values := make([]uint64, runtime.NumCPU())

		if errbw := eo.IpBandwidth.Lookup(&ipUint, &values); errbw != nil {
			if errbw == ebpf.ErrKeyNotExist {
				return 0, nil
			}
			return 0, errbw
		}

		var total uint64
		for _, v := range values {
			total += v
		}

		return total, nil
	})
}
