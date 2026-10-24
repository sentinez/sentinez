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

import (
	"fmt"

	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
)

func Upstream2Target(upstream *edgepb.Upstream) (string, error) {
	var target string
	switch upstream.GetProtocol() {
	case edgepb.ProxyProtocol_PROXY_PROTOCOL_HTTP:
		target = "http://" + upstream.GetServer()
	case edgepb.ProxyProtocol_PROXY_PROTOCOL_HTTPS:
		target = "https://" + upstream.GetServer()
	default:
		return "",
			fmt.Errorf("unsupported protocol: %v", upstream.GetProtocol())
	}

	return target, nil
}
