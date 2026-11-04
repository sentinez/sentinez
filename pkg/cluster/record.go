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

package cluster

import (
	"encoding/binary"
	"errors"
	"sync"
	"time"
)

const _recordDeleted byte = 1

var errBadRecord = errors.New("cluster: malformed sync record")

// record is a write of a SyncMap key: its value, or a tombstone. Writes
// are ordered by version, a hybrid clock, then by origin: the last
// writer wins.
type record struct {
	version uint64
	origin  string
	deleted bool
	payload []byte
}

// newer reports whether r orders after o.
func (r record) newer(o record) bool {
	if r.version != o.version {
		return r.version > o.version
	}

	return r.origin > o.origin
}

// encode lays r out as version | flags | len(origin) | origin | payload.
func (r record) encode() []byte {
	buf := make([]byte, 0, 8+1+binary.MaxVarintLen16+len(r.origin)+
		len(r.payload))

	buf = binary.BigEndian.AppendUint64(buf, r.version)

	var flags byte
	if r.deleted {
		flags = _recordDeleted
	}

	buf = append(buf, flags)
	buf = binary.AppendUvarint(buf, uint64(len(r.origin)))
	buf = append(buf, r.origin...)

	return append(buf, r.payload...)
}

func decodeRecord(buf []byte) (record, error) {
	if len(buf) < 9 {
		return record{}, errBadRecord
	}

	r := record{
		version: binary.BigEndian.Uint64(buf),
		deleted: buf[8]&_recordDeleted != 0,
	}

	origin, rest, err := cutBytes(buf[9:])
	if err != nil {
		return record{}, err
	}

	r.origin = string(origin)
	r.payload = rest

	return r, nil
}

// encodeMessage prefixes the record of key with the key, for PubSub.
func encodeMessage(key string, r record) []byte {
	buf := binary.AppendUvarint(nil, uint64(len(key)))
	buf = append(buf, key...)

	return append(buf, r.encode()...)
}

func decodeMessage(buf []byte) (string, record, error) {
	key, rest, err := cutBytes(buf)
	if err != nil {
		return "", record{}, err
	}

	r, err := decodeRecord(rest)

	return string(key), r, err
}

// cutBytes splits a uvarint length-prefixed field off buf.
func cutBytes(buf []byte) (field, rest []byte, err error) {
	n, size := binary.Uvarint(buf)
	if size <= 0 || uint64(len(buf)-size) < n {
		return nil, nil, errBadRecord
	}

	return buf[size : size+int(n)], buf[size+int(n):], nil
}

// hlc is a hybrid logical clock: wall nanoseconds that never go back and
// move past every version seen, so a write orders after what this node
// already applied even when clocks drift.
type hlc struct {
	mu   sync.Mutex
	last uint64
	now  func() uint64
}

func newHLC() *hlc {
	return &hlc{now: func() uint64 { return uint64(time.Now().UnixNano()) }}
}

func (c *hlc) next() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.last = max(c.now(), c.last+1)

	return c.last
}

func (c *hlc) observe(version uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.last = max(c.last, version)
}
