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

package bpf

import (
	"encoding/binary"
	"fmt"
	"net"

	sentinezbpf "github.com/sentinez/sentinez/bpf/sentinez"
	"github.com/sentinez/sentinez/internal/dmz/dataplane/driver"
)

func BlockCIDR(cidr string) error {
	ip, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}

	ip4 := ip.To4()
	if ip4 == nil {
		return fmt.Errorf("only IPv4 CIDR is supported")
	}

	ones, _ := network.Mask.Size()

	return driver.Exec(func(so *sentinezbpf.SenzObjects) error {
		key := sentinezbpf.SenzIpLpmKey{
			Prefixlen: uint32(ones),
			Ip:        binary.BigEndian.Uint32(ip4),
		}

		return so.Blocklist.Put(key, uint8(1))
	})
}
