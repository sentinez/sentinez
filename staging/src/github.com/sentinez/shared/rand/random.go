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

package rand

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"

	"github.com/cloudresty/ulid"
	"github.com/google/uuid"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/xid"
	"github.com/sentinez/shared/bytesconv"
)

const _charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomString(n int) (string, error) {
	result := make([]byte, n)
	charsetLen := int64(len(_charset))
	for i := range n {
		idxBig, err := rand.Int(rand.Reader, big.NewInt(charsetLen))
		if err != nil {
			return "", err
		}
		result[i] = _charset[idxBig.Int64()]
	}
	return string(result), nil
}

var _bufPool = sync.Pool{
	New: func() any {
		buf := bytes.NewBuffer([]byte{})
		buf.Grow(4096)
		return buf
	},
}

// getBuffer returns an empty buffer from the pool.
func getBuffer() *bytes.Buffer {
	buf, ok := _bufPool.Get().(*bytes.Buffer)
	if !ok {
		return &bytes.Buffer{}
	}
	buf.Reset()

	return buf
}

// NewId generates a new UUID and returns it as a string.
func NewId(prefix string) string {
	id := uuid.New()
	return fmt.Sprintf("%s%s", prefix, id.String())
}

func NewNanoId(prefix string) string {
	id, _ := gonanoid.New()
	return fmt.Sprintf("%s%s", prefix, id)
}

func NewXId(prefix []byte) string {
	guid := xid.New()
	guid.Bytes()

	buf := getBuffer()

	buf.Write(prefix)
	buf.Write(bytesconv.S2b(guid.String()))

	// Copy: the buffer goes back to the pool and is overwritten later.
	res := buf.String()
	_bufPool.Put(buf)

	return res
}

func NewTimeId(prefix []byte, unixtime uint64) string {
	ulidStr, _ := ulid.NewTime(unixtime)

	buf := getBuffer()

	buf.Write(prefix)
	buf.WriteString(ulidStr)

	// Copy: the buffer goes back to the pool and is overwritten later.
	res := buf.String()
	_bufPool.Put(buf)

	return res
}
