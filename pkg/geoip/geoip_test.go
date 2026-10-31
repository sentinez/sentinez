// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package geoip

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fullLog asks for every location field.
type fullLog struct {
	country, city, org string
	asn                int64
}

func (l *fullLog) SetCountry(isoCode string) { l.country = isoCode }
func (l *fullLog) SetCity(name string)       { l.city = name }
func (l *fullLog) SetAsn(number int64)       { l.asn = number }
func (l *fullLog) SetAsnOrg(org string)      { l.org = org }

// countryLog asks for the country only.
type countryLog struct {
	country string
}

func (l *countryLog) SetCountry(isoCode string) { l.country = isoCode }

type fillTest struct {
	name      string
	giveFiles []string
	giveIp    string
	want      fullLog
}

func TestDBFill(t *testing.T) {
	for _, tt := range fillTests() {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(testFS(tt.giveFiles...))
			require.NoError(t, err)

			var got fullLog
			db.Fill(tt.giveIp, &got)
			assert.Equal(t, tt.want, got)
		})
	}
}

// fillTests are the cases of TestDBFill: the test databases map every
// address of 0.0.0.0/1, and none of 128.0.0.0/1.
//
//nolint:funlen
func fillTests() []fillTest {
	return []fillTest{
		{name: "no database", giveIp: "10.0.0.1"},
		{
			name:      "country",
			giveFiles: []string{CountryFile},
			giveIp:    "10.0.0.1",
			want:      fullLog{country: "VN"},
		},
		{
			name:      "city falls back for the country",
			giveFiles: []string{CityFile},
			giveIp:    "10.0.0.1",
			want:      fullLog{country: "VN", city: "Hanoi"},
		},
		{
			name:      "asn",
			giveFiles: []string{AsnFile},
			giveIp:    "10.0.0.1",
			want:      fullLog{asn: 7552, org: "Viettel Group"},
		},
		{
			name:      "ipv4-mapped",
			giveFiles: []string{CountryFile, CityFile, AsnFile},
			giveIp:    "::ffff:10.0.0.1",
			want: fullLog{
				country: "VN", city: "Hanoi",
				asn: 7552, org: "Viettel Group",
			},
		},
		{
			name:      "not found",
			giveFiles: []string{CountryFile, CityFile, AsnFile},
			giveIp:    "200.0.0.1",
		},
		{
			name:      "ipv6 in an ipv4 database",
			giveFiles: []string{CountryFile, CityFile, AsnFile},
			giveIp:    "2001:db8::1",
		},
		{
			name:      "invalid ip",
			giveFiles: []string{CountryFile, CityFile, AsnFile},
			giveIp:    "not-an-ip",
		},
	}
}

func TestDBFillRequestedOnly(t *testing.T) {
	db, err := New(testFS(CityFile, AsnFile))
	require.NoError(t, err)

	var got countryLog
	db.Fill("10.0.0.1", &got)
	assert.Equal(t, countryLog{country: "VN"}, got)

	// a log asking for nothing is left as is
	assert.NotPanics(t, func() { db.Fill("10.0.0.1", struct{}{}) })
}

func TestDBLoaded(t *testing.T) {
	db, err := New(testFS(AsnFile, CountryFile))
	require.NoError(t, err)
	assert.Equal(t, []string{CountryFile, AsnFile}, db.Loaded())

	var none *DB
	assert.Nil(t, none.Loaded())
	assert.NotPanics(t, func() { none.Fill("10.0.0.1", &fullLog{}) })
}

func TestNewInvalidDatabase(t *testing.T) {
	_, err := New(fstest.MapFS{
		CityFile: {Data: []byte("not a database")},
	})
	require.ErrorContains(t, err, "geoip: open "+CityFile)
}

// testFS returns the test databases of names.
func testFS(names ...string) fstest.MapFS {
	records := map[string]mmdbMap{
		CountryFile: {
			{"country", mmdbMap{{"iso_code", "VN"}}},
		},
		CityFile: {
			{"city", mmdbMap{{"names", mmdbMap{{"en", "Hanoi"}}}}},
			{"country", mmdbMap{{"iso_code", "VN"}}},
		},
		AsnFile: {
			{"autonomous_system_number", uint32(7552)},
			{"autonomous_system_organization", "Viettel Group"},
		},
	}

	fsys := make(fstest.MapFS, len(names))
	for _, name := range names {
		fsys[name] = &fstest.MapFile{Data: mmdbOf(records[name])}
	}

	return fsys
}

// mmdbMap is an ordered MaxMind DB map, of string, uint32 or mmdbMap
// values.
type mmdbMap []struct {
	key   string
	value any
}

// mmdbOf returns an IPv4 database of one node: rec is the record of
// 0.0.0.0/1, and 128.0.0.0/1 has none.
func mmdbOf(rec mmdbMap) []byte {
	const nodeCount = 1

	// 24-bit records: a data pointer is node count + 16 + offset, and
	// the node count itself is the empty record
	b := []byte{0, 0, nodeCount + 16, 0, 0, nodeCount}
	b = append(b, make([]byte, 16)...)
	b = encode(b, rec)
	b = append(b, "\xab\xcd\xefMaxMind.com"...)

	return encode(b, mmdbMap{
		{"binary_format_major_version", uint32(2)},
		{"database_type", "Test"},
		{"ip_version", uint32(4)},
		{"node_count", uint32(nodeCount)},
		{"record_size", uint32(24)},
	})
}

// encode appends v to b in the MaxMind DB data format; strings are
// shorter than 285 bytes and maps than 29 entries.
func encode(b []byte, v any) []byte {
	switch v := v.(type) {
	case string:
		if len(v) < 29 {
			b = append(b, 2<<5|byte(len(v)))
		} else {
			b = append(b, 2<<5|29, byte(len(v)-29))
		}
		return append(b, v...)
	case uint32:
		return append(b, 6<<5|4,
			byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
	case mmdbMap:
		b = append(b, 7<<5|byte(len(v)))
		for _, kv := range v {
			b = encode(encode(b, kv.key), kv.value)
		}
	}

	return b
}
