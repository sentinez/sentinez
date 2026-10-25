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
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/sentinez/core/szhttp/httpconst"
)

const (
	// _altSvcDefaultMaxAge is the freshness of an entry without "ma"
	// (RFC 7838 section 3.1).
	_altSvcDefaultMaxAge = 24 * time.Hour
	// _h3BrokenFor keeps a failing upstream on TCP before retrying HTTP/3.
	_h3BrokenFor = 5 * time.Minute
	// _h3HandshakeTimeout bounds the QUIC handshake, so an upstream whose
	// UDP port is filtered falls back to TCP quickly instead of hanging.
	_h3HandshakeTimeout = 3 * time.Second
	// _h3SessionCacheSize lets redials resume TLS (and use 0-RTT).
	_h3SessionCacheSize = 16
)

var _ http.RoundTripper = (*AltSvcTransport)(nil)

// AltSvcTransport sends requests over TCP until the upstream advertises
// HTTP/3 through Alt-Svc, then switches to HTTP/3. It is meant for a single
// upstream (one per reverse proxy target).
type AltSvcTransport struct {
	tcp http.RoundTripper
	h3  http.RoundTripper

	mu          sync.Mutex
	altAddr     string
	expires     time.Time
	brokenUntil time.Time
}

// NewAltSvcTransport returns an [AltSvcTransport] dialing TCP with
// [StandardTransporter].
func NewAltSvcTransport() *AltSvcTransport {
	t := &AltSvcTransport{tcp: StandardTransporter()}
	t.h3 = &http3.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
			ClientSessionCache: tls.NewLRUClientSessionCache(
				_h3SessionCacheSize),
		},
		QUICConfig: &quic.Config{
			HandshakeIdleTimeout: _h3HandshakeTimeout,
		},
		Dial: t.dialH3,
	}

	return t
}

// RoundTrip implements [http.RoundTripper].
func (t *AltSvcTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme == httpconst.SchemeSecure && t.h3Ready() {
		resp, err := t.h3.RoundTrip(req)
		if err == nil {
			return t.learn(resp), nil
		}

		t.markBroken()

		// The body may be partly consumed; only replay what can be rewound.
		if req, err = rewind(req, err); err != nil {
			return nil, err
		}
	}

	resp, err := t.tcp.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	return t.learn(resp), nil
}

func (t *AltSvcTransport) h3Ready() bool {
	now := time.Now()

	t.mu.Lock()
	defer t.mu.Unlock()

	return t.altAddr != "" && now.Before(t.expires) &&
		!now.Before(t.brokenUntil)
}

func (t *AltSvcTransport) markBroken() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.brokenUntil = time.Now().Add(_h3BrokenFor)
}

// learn records the upstream's HTTP/3 alternative.
func (t *AltSvcTransport) learn(resp *http.Response) *http.Response {
	values := resp.Header.Values(httpconst.HeaderAltSvc)
	if len(values) == 0 {
		return resp
	}

	// No usable h3 entry ("clear", ma=0, other protocols) drops HTTP/3.
	addr, maxAge := parseAltSvc(values)
	expires := time.Now().Add(maxAge)

	t.mu.Lock()
	defer t.mu.Unlock()

	t.altAddr, t.expires = addr, expires

	return resp
}

// dialH3 dials the advertised alternative while keeping the origin as SNI,
// as RFC 7838 requires the alternative to present the origin's certificate.
func (t *AltSvcTransport) dialH3(
	ctx context.Context,
	addr string,
	tlsConf *tls.Config,
	conf *quic.Config,
) (*quic.Conn, error) {
	t.mu.Lock()
	alt := t.altAddr
	t.mu.Unlock()

	return quic.DialAddrEarly(ctx, altDialAddr(addr, alt), tlsConf, conf)
}

// altDialAddr resolves an alt-authority such as ":8443" against the origin.
func altDialAddr(origin, alt string) string {
	host, port, err := net.SplitHostPort(alt)
	if err != nil {
		return origin
	}

	if host == "" {
		if host, _, err = net.SplitHostPort(origin); err != nil {
			return origin
		}
	}

	return net.JoinHostPort(host, port)
}

// parseAltSvc returns the first h3 alt-authority and its max age, or zero
// values for "clear" or when no h3 entry is usable.
func parseAltSvc(values []string) (string, time.Duration) {
	for _, value := range values {
		for entry := range strings.SplitSeq(value, ",") {
			if addr, maxAge, ok := parseAltSvcEntry(entry); ok {
				return addr, maxAge
			}
		}
	}

	return "", 0
}

func parseAltSvcEntry(entry string) (string, time.Duration, bool) {
	params := strings.Split(entry, ";")

	proto, authority, found := strings.Cut(strings.TrimSpace(params[0]), "=")
	if !found || proto != http3.NextProtoH3 {
		return "", 0, false
	}

	authority, err := strconv.Unquote(authority)
	if err != nil || authority == "" {
		return "", 0, false
	}

	maxAge := _altSvcDefaultMaxAge
	for _, param := range params[1:] {
		key, val, _ := strings.Cut(strings.TrimSpace(param), "=")
		if key != "ma" {
			continue
		}

		secs, err := strconv.ParseUint(strings.Trim(val, `"`), 10, 32)
		if err != nil {
			return "", 0, false
		}

		maxAge = time.Duration(secs) * time.Second
	}

	return authority, maxAge, maxAge > 0
}

// rewind returns a request whose body can be sent again, or cause when the
// body cannot be replayed.
func rewind(req *http.Request, cause error) (*http.Request, error) {
	if req.Body == nil || req.Body == http.NoBody {
		return req, nil
	}

	if req.GetBody == nil {
		return nil, cause
	}

	body, err := req.GetBody()
	if err != nil {
		return nil, cause
	}

	req = req.Clone(req.Context())
	req.Body = body

	return req, nil
}
