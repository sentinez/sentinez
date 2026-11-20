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

package httpx

import (
	"net/textproto"

	"github.com/sentinez/core/szhttp/httpconst"
	"github.com/sentinez/pkg/szcode"
)

// The header names and values the edge sets, as the strings a response
// header keeps: setting one of them takes the string of the table instead
// of a copy of its bytes. Read only.
var (
	// _internedKeys maps a name to its canonical form, the key of an
	// http.Header.
	_internedKeys = canonicalOf(
		httpconst.HeaderServer, httpconst.HeaderXRequestId,
		httpconst.HeaderContentType, httpconst.HeaderXCache,
		httpconst.HeaderCacheControl, httpconst.HeaderAltSvc,
		httpconst.HeaderSetCookie, httpconst.HeaderLocation,
		httpconst.HeaderETag, httpconst.HeaderContentLength,
		httpconst.HeaderAllow, httpconst.HeaderXContentTypeOpts,
	)

	_internedValues = identityOf(
		szcode.Name,
		httpconst.ValueTextPlain, httpconst.ValueTextHTML,
		httpconst.ValueAppJSON, httpconst.ValueNoSniff,
		httpconst.ValueAllowGetHead, httpconst.ValueNoStore,
		httpconst.ValueCacheImmutable, httpconst.ValueHitCache,
	)
)

// canonicalOf maps each name, as given and canonical (e.g. ETag, Etag, the
// form a copied header has), to its canonical form.
func canonicalOf(names ...string) map[string]string {
	m := make(map[string]string, 2*len(names))
	for _, name := range names {
		canonical := textproto.CanonicalMIMEHeaderKey(name)
		m[name] = canonical
		m[canonical] = canonical
	}

	return m
}

func identityOf(values ...string) map[string]string {
	m := make(map[string]string, len(values))
	for _, v := range values {
		m[v] = v
	}

	return m
}

// headerKey the header name k as a string: the interned one if known,
// else a copy (http.Header canonicalizes it).
func headerKey(k []byte) string { return intern(_internedKeys, k) }

// headerValue the header value v as a string: the interned one if known,
// else a copy.
func headerValue(v []byte) string { return intern(_internedValues, v) }

func intern(table map[string]string, b []byte) string {
	if s, ok := table[string(b)]; ok {
		return s
	}

	return string(b)
}
