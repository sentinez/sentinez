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

package appconfig

import (
	"maps"
	"slices"
	"strings"

	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// KVPrefix is the Consul folder of the env: KVPrefix+"SENZ_KEY" is shared
// by every service, KVPrefix+"<service_key>/SENZ_KEY" is for one service.
const KVPrefix = "sentinez/env/"

// _envOnly are the keys that only the environment may set: the address of
// Consul itself, and what differs from one instance to the next. The list is
// mirrored by ENV_ONLY in hack/consul-env-put.sh and by the docs
// (docs/02-getting-started.md): keep them equal.
var _envOnly = []settingpb.Senz{
	settingpb.Senz_SENZ_CONSUL_URI,
	settingpb.Senz_SENZ_ADDRESS,
	settingpb.Senz_SENZ_MEMBERSHIP_ADDRESS,
	settingpb.Senz_SENZ_DISCOVERY_ADDRESS,
	settingpb.Senz_SENZ_ZONE,
}

// judge looks at the key at path, relative to KVPrefix, for the service of
// serviceKey that reads keys (none meaning all of them). It returns the env
// key name and whether it is in the folder of the service, or an empty name
// when the key is not for this service. refused is true for a key to warn
// about: an unknown one, or one that only the env may set.
func judge(path, serviceKey string, keys []settingpb.Senz,
) (name string, scoped, refused bool) {
	// without a "/" the whole path is the key, and is returned as folder
	folder, name, scoped := strings.Cut(path, "/")
	if !scoped {
		name = folder
	}

	if scoped && folder != serviceKey {
		return "", false, false
	}

	key, known := settingpb.Senz_value[name]
	senz := settingpb.Senz(key)

	if !known || senz == settingpb.Senz_SENZ_UNSPECIFIED ||
		slices.Contains(_envOnly, senz) {

		return "", false, true
	}

	if len(keys) != 0 && !slices.Contains(keys, senz) {
		return "", false, false
	}

	return name, scoped, false
}

// overlay returns the env values to take from Consul. kv is the content of
// KVPrefix, keyed relative to it. A value of the folder of serviceKey wins
// over the shared one, and a value the environment already has wins over
// both. skipped names what is refused, to be warned about: an unknown key,
// or one that only the env may set.
func overlay(env, kv map[string]string, serviceKey string,
	keys []settingpb.Senz) (found map[string]string, skipped []string) {

	found = make(map[string]string, len(kv))
	own := make(map[string]string, len(kv))

	for path, value := range kv {
		if value == "" {
			continue
		}

		name, scoped, refused := judge(path, serviceKey, keys)

		switch {
		case refused:
			skipped = append(skipped, path)
		case name == "" || env[name] != "":
			// not for this service, or the environment has it
		case scoped:
			own[name] = value
		default:
			found[name] = value
		}
	}

	maps.Copy(found, own)
	slices.Sort(skipped)

	return found, skipped
}
