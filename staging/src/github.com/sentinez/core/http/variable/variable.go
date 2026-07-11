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

package variable

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/sentinez/core/common/bytestr"
	corehttp "github.com/sentinez/core/http"
	"github.com/sentinez/shared/bytesconv"
)

// ParseProxyVal evaluates standard Nginx-like string variables.
func ParseProxyVal(ctx corehttp.Context, val string) (string, error) {
	if !strings.Contains(val, "$") {
		return val, nil
	}

	// Fast path for exact matches
	switch val {
	case VarHost:
		return bytesconv.B2s(ctx.Host()), nil
	case VarRemoteAddr:
		return bytesconv.B2s(ctx.RequestIP()), nil
	case VarScheme:
		return ctx.Scheme(), nil
	case VarRequestURI:
		return bytesconv.B2s(ctx.URI()), nil
	case VarProxyAddXForwardedFor:
		xff := ctx.Header(bytestr.HeaderXForwardedFor)
		ip := ctx.RequestIP()
		if xff == nil {
			return bytesconv.B2s(ip), nil
		}

		var buf bytes.Buffer
		buf.Write(xff)
		buf.WriteString(", ")
		buf.Write(ip)

		return buf.String(), nil
	}

	return "", fmt.Errorf("unsupported variable: %s", val)
}

// IsValidHeaderKey checks if a string is a
// valid HTTP header key according to RFC 7230.
func IsValidHeaderKey(key string) bool {
	if len(key) == 0 {
		return false
	}
	for i := 0; i < len(key); i++ {
		c := key[i]
		if !validHeaderFieldByte(c) {
			return false
		}
	}
	return true
}

func validHeaderFieldByte(c byte) bool {
	// tchar = "!" / "#" / "$" / "%" / "&" / "'" / "*" / "+" / "-" / "." /
	// "^" / "_" / "`" / "|" / "~" / DIGIT / ALPHA
	// (ASCII 33-126 excluding separators)
	if c <= 32 || c >= 127 {
		return false
	}
	switch c {
	case '(', ')', '<', '>', '@', ',', ';', ':',
		'\\', '"', '/', '[', ']', '?', '=', '{', '}':
		return false
	}
	return true
}
