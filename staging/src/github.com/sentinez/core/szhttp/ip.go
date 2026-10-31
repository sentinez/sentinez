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

package szhttp

import "bytes"

// StripPort returns the host of addr without its port: "1.2.3.4:80" and
// "[::1]:80" give "1.2.3.4" and "::1", while a bare IPv4 or IPv6 address
// is returned as is. It slices addr, without allocating, and leaves the
// spaces around it.
func StripPort(addr []byte) []byte {
	if len(addr) > 0 && addr[0] == '[' {
		if end := bytes.IndexByte(addr, ']'); end > 0 {
			return addr[1:end]
		}

		return addr
	}

	// one colon: host:port; more: a bare IPv6 address
	i := bytes.IndexByte(addr, ':')
	if i < 0 || bytes.IndexByte(addr[i+1:], ':') >= 0 {
		return addr
	}

	return addr[:i]
}
