// Copyright 2025 Duc-Hung Ho.
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

package settingpb

func (x *Config) Get(key Senz) string {
	if x.GetEnv() == nil {
		return ""
	}

	value, ok := x.GetEnv()[key.String()]
	if !ok {
		return ""
	}

	return value
}

// Clustered reports whether the service joins peers: both its membership
// and discovery addresses are set.
func (x *Config) Clustered() bool {
	return x.Get(Senz_SENZ_MEMBERSHIP_ADDRESS) != "" &&
		x.Get(Senz_SENZ_DISCOVERY_ADDRESS) != ""
}

func (x *Config) GetDefault(key Senz, defaultValue string) string {
	val := x.Get(key)

	if val == "" {
		return defaultValue
	}

	return val
}
