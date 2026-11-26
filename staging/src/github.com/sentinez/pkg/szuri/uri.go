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

// Package szuri parses the URIs of the SZ_*_URI settings: a scheme, an
// optional user, one or more comma-separated hosts, a path and a query,
// e.g. "kafkas://k1:9093,k2:9093" or "postgres://u:p@db:5432/app?x=y".
// net/url takes one host only, hence this package.
package szuri

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

const _maxPort = 65535

var (
	// ErrNoScheme is returned for a value without "scheme://", e.g. a
	// bare "host:port".
	ErrNoScheme = errors.New("szuri: no scheme")
	// ErrNoHost is returned by Require for a URI without a host.
	ErrNoHost = errors.New("szuri: no host")
)

// SchemeError is returned for a scheme the setting does not take.
type SchemeError struct {
	Scheme string
	Want   []string
}

func (e *SchemeError) Error() string {
	return fmt.Sprintf("szuri: scheme %q, want %s", e.Scheme,
		strings.Join(e.Want, " or "))
}

// URI is a parsed URI.
type URI struct {
	// Scheme is lowercased, e.g. "kafkas".
	Scheme string
	// User is nil without a user.
	User *url.Userinfo
	// Hosts are the host[:port] of the authority, in order; none for e.g.
	// "unix:///run/consul.sock".
	Hosts []string
	// Path is unescaped, "" or starting with "/".
	Path  string
	Query url.Values
}

// Parse parses s, "scheme://[user[:password]@]host[:port][,host[:port]...]
// [/path][?query]". The fragment is dropped.
func Parse(s string) (URI, error) {
	scheme, rest, ok := strings.Cut(strings.TrimSpace(s), "://")
	if !ok || scheme == "" {
		return URI{}, ErrNoScheme
	}

	rest, _, _ = strings.Cut(rest, "#")
	rest, rawQuery, _ := strings.Cut(rest, "?")

	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return URI{}, fmt.Errorf("szuri: query: %w", err)
	}

	authority, path := rest, ""
	if i := strings.IndexByte(rest, '/'); i >= 0 {
		authority, path = rest[:i], rest[i:]
	}

	if path, err = url.PathUnescape(path); err != nil {
		return URI{}, fmt.Errorf("szuri: path: %w", err)
	}

	u := URI{Scheme: strings.ToLower(scheme), Path: path, Query: query}
	if u.User, authority, err = parseUser(authority); err != nil {
		return URI{}, err
	}

	if u.Hosts, err = parseHosts(authority); err != nil {
		return URI{}, err
	}

	return u, nil
}

// Secure reports whether the scheme is secure (TLS) rather than plain,
// e.g. Secure("kafka", "kafkas"); any other scheme is a *SchemeError.
func (u URI) Secure(plain, secure string) (bool, error) {
	if err := u.Want(plain, secure); err != nil {
		return false, err
	}

	return u.Scheme == secure, nil
}

// Want returns a *SchemeError when the scheme is none of schemes.
func (u URI) Want(schemes ...string) error {
	if slices.Contains(schemes, u.Scheme) {
		return nil
	}

	return &SchemeError{Scheme: u.Scheme, Want: slices.Clone(schemes)}
}

// Host returns the one host of u, an error for none or several.
func (u URI) Host() (string, error) {
	if len(u.Hosts) != 1 {
		return "", fmt.Errorf("szuri: %d hosts, want one", len(u.Hosts))
	}

	return u.Hosts[0], nil
}

// Require returns ErrNoHost when u has no host.
func (u URI) Require() error {
	if len(u.Hosts) == 0 {
		return ErrNoHost
	}

	return nil
}

// parseUser splits the user off the authority.
func parseUser(authority string) (*url.Userinfo, string, error) {
	i := strings.LastIndexByte(authority, '@')
	if i < 0 {
		return nil, authority, nil
	}

	name, password, hasPassword := strings.Cut(authority[:i], ":")

	name, err := url.PathUnescape(name)
	if err != nil {
		return nil, "", fmt.Errorf("szuri: user: %w", err)
	}

	if !hasPassword {
		return url.User(name), authority[i+1:], nil
	}

	if password, err = url.PathUnescape(password); err != nil {
		return nil, "", fmt.Errorf("szuri: password: %w", err)
	}

	return url.UserPassword(name, password), authority[i+1:], nil
}

// parseHosts the comma-separated hosts of the authority.
func parseHosts(authority string) ([]string, error) {
	if authority == "" {
		return nil, nil
	}

	hosts := strings.Split(authority, ",")
	for i, host := range hosts {
		host = strings.TrimSpace(host)
		if err := checkHost(host); err != nil {
			return nil, err
		}

		hosts[i] = host
	}

	return hosts, nil
}

// checkHost fails for an empty host, or a bad port.
func checkHost(host string) error {
	if host == "" {
		return errors.New("szuri: empty host")
	}

	if !hasPort(host) {
		return nil
	}

	_, port, err := net.SplitHostPort(host)
	if err != nil {
		return fmt.Errorf("szuri: host %q: %w", host, err)
	}

	if _, err := ParsePort(port); err != nil {
		return fmt.Errorf("szuri: host %q: %w", host, err)
	}

	return nil
}

// ParsePort the port number of s, 1 to 65535.
func ParsePort(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > _maxPort {
		return 0, fmt.Errorf("bad port %q", s)
	}

	return n, nil
}

// hasPort reports whether host ends with a port: a ":" after the IPv6
// brackets, if any.
func hasPort(host string) bool {
	if strings.HasPrefix(host, "[") {
		return strings.Contains(host, "]:")
	}

	return strings.Contains(host, ":")
}
