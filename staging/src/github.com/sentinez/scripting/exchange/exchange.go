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

package exchange

import "strconv"

// Exchange is the request a hook reads and may change: the edge context
// satisfies it through a small adapter.
type Exchange interface {
	Method() []byte
	Path() []byte
	Host() []byte
	RequestIP() []byte
	Header(k []byte) []byte
	Query(k []byte) []byte
	// HasHeader reports whether the request has the header k, its name
	// as received.
	HasHeader(k []byte) bool
	// HasQuery reports whether the request has the query parameter k.
	HasQuery(k []byte) bool
	Body() []byte
	// JA4 the JA4 fingerprint of the TLS client hello, "" without one.
	JA4() string
	// TLS reports whether the request came over TLS.
	TLS() bool
	// SetHeader sets a header of the request sent upstream.
	SetHeader(k, v string)
	SetResponseHeader(k, v string)
	// String answers the request, ending it.
	String(status int, body []byte) error
	// Log writes msg, from log.debug or log.info of the script, to the
	// logs of the request.
	Log(level Level, script, msg string)
}

// Level is the level of a log line of a script.
type Level int

const (
	// LevelDebug log.debug: for the debug logs only.
	LevelDebug Level = iota + 1
	// LevelInfo log.info.
	LevelInfo
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	}

	return "Level(" + strconv.Itoa(int(l)) + ")"
}
