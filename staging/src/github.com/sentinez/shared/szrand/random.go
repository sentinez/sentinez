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

package szrand

import (
	crand "crypto/rand"
	"math/big"
	"os"
	"strings"
	"sync"

	"github.com/google/uuid"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/oklog/ulid/v2"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/shared/szprefix"
)

const _charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomString(n int) (string, error) {
	result := make([]byte, n)
	charsetLen := int64(len(_charset))
	for i := range n {
		idxBig, err := crand.Int(crand.Reader, big.NewInt(charsetLen))
		if err != nil {
			return "", err
		}
		result[i] = _charset[idxBig.Int64()]
	}
	return string(result), nil
}

// _zone is "<SENZ_ZONE>-", the zone the service is deployed in, prefixing
// every generated id ("" when unset). It is read on the first New*Id call,
// once the env file is loaded.
var _zone = sync.OnceValue(loadZone)

// NewUUID returns a prefixed, never zoned UUID: the primary key of a
// record, matched by the proto id prefix rules.
func NewUUID(p szprefix.Prefix) string {
	return p.Key() + uuid.NewString()
}

// NewNanoId returns a zoned, prefixed nanoid.
func NewNanoId(p szprefix.Prefix) string {
	id, _ := gonanoid.New()
	return zoned(p, id)
}

// NewULID returns a zoned, prefixed ULID, sortable by creation time to the
// millisecond: the id of requests and the events raised while handling
// them. Entropy comes from crypto/rand, which is lock-free and never fails,
// unlike ulid.Make, whose monotonic reader takes one global mutex per id.
func NewULID(p szprefix.Prefix) string {
	id, _ := ulid.New(ulid.Now(), crand.Reader)
	return zoned(p, id.String())
}

func zoned(p szprefix.Prefix, id string) string {
	return _zone() + p.Key() + id
}

func loadZone() string {
	zone := os.Getenv(settingpb.Senz_SENZ_ZONE.String())
	if zone == "" {
		return ""
	}

	return strings.ToLower(zone) + "-"
}
