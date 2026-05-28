// Copyright 2025 Sentinéz Labs.
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

package copier

import (
	"github.com/sentinez/shared/jsonx"
	"github.com/sentinez/shared/protobuf/protox"

	google "google.golang.org/protobuf/proto"
)

// CopyProtoMessage copies the src message to the dst message.
func CopyProtoMessage(src, dst google.Message) error {
	bytes, err := protox.Marshal(src)
	if err != nil {
		return err
	}

	return protox.Unmarshal(bytes, dst)
}

// CopyJSON copies the src object to the dst object.
func CopyJSON(src, dst any) error {
	bytes, err := jsonx.Marshal(src)
	if err != nil {
		return err
	}

	return jsonx.Unmarshal(bytes, dst)
}
