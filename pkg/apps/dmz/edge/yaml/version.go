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

package edgeyaml

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"google.golang.org/protobuf/proto"
)

// _versionLen hex characters of the setting hash make its version.
const _versionLen = 12

// Version names setting by a hash of its content: two edges share a
// version iff they run the same setting.
func Version(setting *edgepb.Setting) (string, error) {
	raw, err := proto.MarshalOptions{Deterministic: true}.Marshal(setting)
	if err != nil {
		return "", fmt.Errorf("encode setting: %w", err)
	}

	sum := sha256.Sum256(raw)

	return hex.EncodeToString(sum[:])[:_versionLen], nil
}
