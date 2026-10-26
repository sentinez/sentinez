// Copyright 2026 Sentinéz Labs.
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

package network

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sentinez/core/szhttp/httpconst"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errH3Down = errors.New("h3 down")

type fakeTransport struct {
	calls  int
	altSvc string
	err    error
	// readBody reads one byte of the request body before failing.
	readBody bool
	// gotBody is the request body of the last successful call.
	gotBody string
}

func (f *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	f.calls++
	if f.err != nil {
		if f.readBody {
			_, _ = req.Body.Read(make([]byte, 1))
		}
		return nil, f.err
	}

	if req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		f.gotBody = string(body)
	}

	resp := &http.Response{StatusCode: http.StatusOK, Header: http.Header{}}
	if f.altSvc != "" {
		resp.Header.Set(httpconst.HeaderAltSvc, f.altSvc)
	}

	return resp, nil
}

func TestParseAltSvc(t *testing.T) {
	tests := []struct {
		name       string
		give       []string
		wantAddr   string
		wantMaxAge time.Duration
	}{
		{name: "h3 with ma", give: []string{`h3=":443"; ma=3600`},
			wantAddr: ":443", wantMaxAge: time.Hour},
		{name: "default ma", give: []string{`h3="alt.example:8443"`},
			wantAddr: "alt.example:8443", wantMaxAge: _altSvcDefaultMaxAge},
		{name: "h3 after others",
			give:     []string{`h2=":443"`, `h3-29=":443", h3=":9443"`},
			wantAddr: ":9443", wantMaxAge: _altSvcDefaultMaxAge},
		{name: "clear", give: []string{"clear"}},
		{name: "zero ma", give: []string{`h3=":443"; ma=0`}},
		{name: "bad ma", give: []string{`h3=":443"; ma=x`}},
		{name: "unquoted", give: []string{`h3=:443`}},
		{name: "no h3", give: []string{`h2=":443"`}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, maxAge := parseAltSvc(tt.give)
			assert.Equal(t, tt.wantAddr, addr)
			assert.Equal(t, tt.wantMaxAge, maxAge)
		})
	}
}

func TestAltDialAddr(t *testing.T) {
	tests := []struct {
		name     string
		giveAlt  string
		wantAddr string
	}{
		{name: "port only", giveAlt: ":8443", wantAddr: "up.example:8443"},
		{name: "host+port", giveAlt: "b.example:9", wantAddr: "b.example:9"},
		{name: "invalid", giveAlt: "nope", wantAddr: "up.example:443"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := altDialAddr("up.example:443", tt.giveAlt)
			assert.Equal(t, tt.wantAddr, got)
		})
	}
}

func TestAltSvcTransportSwitchesToH3(t *testing.T) {
	tcp := &fakeTransport{altSvc: `h3=":443"; ma=60`}
	h3 := &fakeTransport{}
	tr := newTestTransport(tcp, h3)

	roundTrip(t, tr, "https://up.example/")
	roundTrip(t, tr, "https://up.example/")
	assert.Equal(t, 1, tcp.calls)
	assert.Equal(t, 1, h3.calls)
}

func TestAltSvcTransportKeepsPlainHTTPOnTCP(t *testing.T) {
	tcp := &fakeTransport{altSvc: `h3=":443"`}
	h3 := &fakeTransport{}
	tr := newTestTransport(tcp, h3)

	roundTrip(t, tr, "http://up.example/")
	roundTrip(t, tr, "http://up.example/")
	assert.Equal(t, 2, tcp.calls)
	assert.Zero(t, h3.calls)
}

func TestAltSvcTransportFallsBackWhenH3Fails(t *testing.T) {
	tcp := &fakeTransport{altSvc: `h3=":443"`}
	h3 := &fakeTransport{err: errH3Down}
	tr := newTestTransport(tcp, h3)

	roundTrip(t, tr, "https://up.example/")
	roundTrip(t, tr, "https://up.example/")
	roundTrip(t, tr, "https://up.example/")
	assert.Equal(t, 3, tcp.calls)
	assert.Equal(t, 1, h3.calls, "h3 must stay off while broken")
}

func TestAltSvcTransportDoesNotReplayStreamedBody(t *testing.T) {
	tcp := &fakeTransport{altSvc: `h3=":443"`}
	h3 := &fakeTransport{err: errH3Down, readBody: true}
	tr := newTestTransport(tcp, h3)
	roundTrip(t, tr, "https://up.example/")

	_, err := tr.RoundTrip(newStreamRequest(t, "payload"))
	require.ErrorIs(t, err, errH3Down)
	assert.Equal(t, 1, tcp.calls)
}

// An HTTP/3 GET reaches the proxy with an open body of unknown length; a
// failed handshake never reads it, so it goes over TCP unchanged.
func TestAltSvcTransportReplaysUnreadStreamedBody(t *testing.T) {
	tcp := &fakeTransport{altSvc: `h3=":443"`}
	h3 := &fakeTransport{err: errH3Down}
	tr := newTestTransport(tcp, h3)
	roundTrip(t, tr, "https://up.example/")

	_, err := tr.RoundTrip(newStreamRequest(t, "payload"))
	require.NoError(t, err)
	assert.Equal(t, 2, tcp.calls)
	assert.Equal(t, "payload", tcp.gotBody)
}

func TestAltSvcTransportClearDropsH3(t *testing.T) {
	tcp := &fakeTransport{altSvc: `h3=":443"`}
	h3 := &fakeTransport{altSvc: "clear"}
	tr := newTestTransport(tcp, h3)

	roundTrip(t, tr, "https://up.example/")
	roundTrip(t, tr, "https://up.example/")
	tcp.altSvc = ""
	roundTrip(t, tr, "https://up.example/")
	assert.Equal(t, 2, tcp.calls)
	assert.Equal(t, 1, h3.calls)
}

func newTestTransport(tcp, h3 http.RoundTripper) *AltSvcTransport {
	return &AltSvcTransport{tcp: tcp, h3: h3}
}

func roundTrip(t *testing.T, tr http.RoundTripper, url string) {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	_, err = tr.RoundTrip(req)
	require.NoError(t, err)
}

// newStreamRequest returns a request whose body cannot be rewound, as
// httputil.ReverseProxy forwards a client's.
func newStreamRequest(t *testing.T, body string) *http.Request {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, "https://up.example/", nil)
	require.NoError(t, err)
	req.Body = io.NopCloser(strings.NewReader(body))
	req.ContentLength = -1

	return req
}
