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

package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/corazawaf/libinjection-go"
	"github.com/google/uuid"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/httpconst"
	"github.com/sentinez/shared/szlog"
)

const (
	_attackSQLi = "sqli"
	_attackXSS  = "xss"

	_locationPath   = "path"
	_locationQuery  = "query"
	_locationHeader = "header"
	_locationCookie = "cookie"
	_locationBody   = "body"

	_mimeForm = "application/x-www-form-urlencoded"
	_mimeHTML = "text/html; charset=utf-8"
)

// _inspectedHeaders are the client-controlled headers worth inspecting.
// Credentials such as Authorization are opaque tokens and are left out.
var _inspectedHeaders = []string{"User-Agent", "Referer"}

// violation describes an injection payload found in a request.
type violation struct {
	attack   string
	location string
}

// Injection rejects requests carrying SQL injection or XSS payloads in the
// path, query string, selected headers, cookies or body. Offending requests
// get the HTML 403 page and never reach h.
func Injection(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read request body", http.StatusBadRequest)
			return
		}

		if v, ok := inspect(r, body); ok {
			forbidden(w, r, v)
			return
		}

		clonedR := r.Clone(r.Context())
		clonedR.Body = io.NopCloser(bytes.NewReader(body))

		h.ServeHTTP(w, clonedR)
	})
}

func inspect(r *http.Request, body []byte) (violation, bool) {
	if attack, ok := scanPath(r.URL.Path); ok {
		return violation{attack: attack, location: _locationPath}, true
	}

	if attack, ok := scanValues(r.URL.Query()); ok {
		return violation{attack: attack, location: _locationQuery}, true
	}

	if attack, ok := scanHeaders(r.Header); ok {
		return violation{attack: attack, location: _locationHeader}, true
	}

	if attack, ok := scanCookies(r.Cookies()); ok {
		return violation{attack: attack, location: _locationCookie}, true
	}

	ctype := r.Header.Get(httpconst.HeaderContentType)
	if attack, ok := scanBody(ctype, body); ok {
		return violation{attack: attack, location: _locationBody}, true
	}

	return violation{}, false
}

// scanPath checks the decoded path as a whole and segment by segment, since
// a payload is usually confined to a single path parameter.
func scanPath(path string) (string, bool) {
	if attack, ok := detect(path); ok {
		return attack, true
	}

	for seg := range strings.SplitSeq(path, "/") {
		if attack, ok := detect(seg); ok {
			return attack, true
		}
	}

	return "", false
}

func scanValues(values url.Values) (string, bool) {
	for key, vals := range values {
		if attack, ok := detect(key); ok {
			return attack, true
		}

		for _, val := range vals {
			if attack, ok := detect(val); ok {
				return attack, true
			}
		}
	}

	return "", false
}

func scanHeaders(header http.Header) (string, bool) {
	for _, name := range _inspectedHeaders {
		for _, val := range header.Values(name) {
			if attack, ok := detect(val); ok {
				return attack, true
			}
		}
	}

	return "", false
}

func scanCookies(cookies []*http.Cookie) (string, bool) {
	for _, cookie := range cookies {
		if attack, ok := detect(cookie.Value); ok {
			return attack, true
		}
	}

	return "", false
}

// scanBody inspects form fields and JSON strings individually; any other
// payload is inspected as raw text.
func scanBody(contentType string, body []byte) (string, bool) {
	if len(body) == 0 {
		return "", false
	}

	if mediaType, _, _ := mime.ParseMediaType(contentType); mediaType ==
		_mimeForm {
		if form, err := url.ParseQuery(string(body)); err == nil {
			return scanValues(form)
		}
	}

	var doc any
	if err := json.Unmarshal(body, &doc); err == nil {
		return scanJSON(doc)
	}

	return detect(string(body))
}

func scanJSON(node any) (string, bool) {
	switch v := node.(type) {
	case string:
		return detect(v)
	case []any:
		for _, item := range v {
			if attack, ok := scanJSON(item); ok {
				return attack, true
			}
		}
	case map[string]any:
		for key, item := range v {
			if attack, ok := detect(key); ok {
				return attack, true
			}

			if attack, ok := scanJSON(item); ok {
				return attack, true
			}
		}
	}

	return "", false
}

func detect(input string) (string, bool) {
	if input == "" {
		return "", false
	}

	if ok, _ := libinjection.IsSQLi(input); ok {
		return _attackSQLi, true
	}

	if libinjection.IsXSS(input) {
		return _attackXSS, true
	}

	return "", false
}

// forbidden writes the HTML 403 page for a request that tripped v.
func forbidden(w http.ResponseWriter, r *http.Request, v violation) {
	id := r.Header.Get(httpconst.HeaderXRequestId)
	if id == "" {
		id = uuid.NewString()
	}

	szlog.Warnf("deny http request: %s in %s method=%s path=%s id=%s",
		v.attack, v.location, r.Method, r.URL.Path, id)

	var page bytes.Buffer
	if err := szhttp.ForbiddenPage(id).Render(r.Context(), &page); err != nil {
		http.Error(w, http.StatusText(http.StatusForbidden),
			http.StatusForbidden)

		return
	}

	w.Header().Set(httpconst.HeaderContentType, _mimeHTML)
	w.Header().Set(httpconst.HeaderXRequestId, id)
	w.WriteHeader(http.StatusForbidden)

	_, _ = w.Write(page.Bytes())
}
