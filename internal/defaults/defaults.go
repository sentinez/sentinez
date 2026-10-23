// Copyright 2026 Sentinéz Labs.
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

package defaults

// Listen addresses used when the matching setting is unset.
const (
	HTTPSAddress     = "0.0.0.0:443"
	DiscoveryAddress = "0.0.0.0"
	DiscoveryPort    = 3322
	EdgeAddress      = "0.0.0.0:8443"
	APIServerAddress = "0.0.0.0:9000"
	RealtimeAddress  = "0.0.0.0:7778"
	PprofAddress     = "0.0.0.0:6060"
	DataplaneAddress = "0.0.0.0:9001"
	GreeterAddress   = "0.0.0.0:9001"
)

const (
	NamespaceSetting = "sentinez.dmz.edge.setting"
)
