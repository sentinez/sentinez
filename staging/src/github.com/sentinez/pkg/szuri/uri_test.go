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

package szuri_test

import (
	"net/url"
	"testing"

	"github.com/sentinez/pkg/szuri"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type parseCase struct {
	name string
	give string
	want szuri.URI
}

var _parseCases = []parseCase{
	{
		name: "one host",
		give: "https://otel:4317",
		want: szuri.URI{
			Scheme: "https",
			Hosts:  []string{"otel:4317"},
			Query:  url.Values{},
		},
	},
	{
		name: "several hosts",
		give: "KAFKAS://k1:9093, k2:9093,[::1]:9093",
		want: szuri.URI{
			Scheme: "kafkas",
			Hosts:  []string{"k1:9093", "k2:9093", "[::1]:9093"},
			Query:  url.Values{},
		},
	},
	{
		name: "user, path and query",
		give: "postgres://u:p%40ss@db:5432/app?sslmode=verify-full#x",
		want: szuri.URI{
			Scheme: "postgres",
			User:   url.UserPassword("u", "p@ss"),
			Hosts:  []string{"db:5432"},
			Path:   "/app",
			Query:  url.Values{"sslmode": {"verify-full"}},
		},
	},
	{
		name: "user without password",
		give: "kafka://alice@k1",
		want: szuri.URI{
			Scheme: "kafka",
			User:   url.User("alice"),
			Hosts:  []string{"k1"},
			Query:  url.Values{},
		},
	},
	{
		name: "no host",
		give: "unix:///run/consul.sock",
		want: szuri.URI{
			Scheme: "unix",
			Path:   "/run/consul.sock",
			Query:  url.Values{},
		},
	},
}

func TestParse(t *testing.T) {
	for _, tt := range _parseCases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := szuri.Parse(tt.give)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseError(t *testing.T) {
	tests := []struct {
		name    string
		give    string
		wantErr error
	}{
		{name: "bare address", give: "k1:9092", wantErr: szuri.ErrNoScheme},
		{name: "empty", give: "", wantErr: szuri.ErrNoScheme},
		{name: "no scheme name", give: "://k1", wantErr: szuri.ErrNoScheme},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := szuri.Parse(tt.give)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestParseBadHost(t *testing.T) {
	tests := []struct {
		name string
		give string
	}{
		{name: "empty host", give: "kafka://k1:9092,,k2:9092"},
		{name: "trailing comma", give: "olric://e1:3320,"},
		{name: "port not a number", give: "kafka://k1:abc"},
		{name: "port out of range", give: "kafka://k1:70000"},
		{name: "empty port", give: "kafka://k1:"},
		{name: "bad escape", give: "kafka://u%zz@k1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := szuri.Parse(tt.give)
			require.Error(t, err)
		})
	}
}

func TestURISecure(t *testing.T) {
	tests := []struct {
		name string
		give string
		want bool
	}{
		{name: "plain", give: "kafka://k1:9092", want: false},
		{name: "secure", give: "kafkas://k1:9093", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := szuri.Parse(tt.give)
			require.NoError(t, err)

			got, err := u.Secure("kafka", "kafkas")
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestURISecureOtherScheme(t *testing.T) {
	u, err := szuri.Parse("http://k1:9092")
	require.NoError(t, err)

	_, err = u.Secure("kafka", "kafkas")

	var schemeErr *szuri.SchemeError
	require.ErrorAs(t, err, &schemeErr)
	assert.Equal(t, "http", schemeErr.Scheme)
}

func TestURIRequire(t *testing.T) {
	u, err := szuri.Parse("unix:///run/consul.sock")
	require.NoError(t, err)

	require.ErrorIs(t, u.Require(), szuri.ErrNoHost)
}

func TestURIWant(t *testing.T) {
	u, err := szuri.Parse("unix:///run/consul.sock")
	require.NoError(t, err)

	require.NoError(t, u.Want("http", "https", "unix"))

	var schemeErr *szuri.SchemeError
	require.ErrorAs(t, u.Want("http", "https"), &schemeErr)
	assert.Equal(t, []string{"http", "https"}, schemeErr.Want)
}

func TestURIHost(t *testing.T) {
	tests := []struct {
		name    string
		give    string
		want    string
		wantErr bool
	}{
		{name: "one", give: "http://a:1", want: "a:1"},
		{name: "none", give: "http://", wantErr: true},
		{name: "several", give: "http://a:1,b:2", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := szuri.Parse(tt.give)
			require.NoError(t, err)

			got, err := u.Host()
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		give    string
		want    int
		wantErr bool
	}{
		{give: "3320", want: 3320},
		{give: "0", wantErr: true},
		{give: "65536", wantErr: true},
		{give: "x", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got, err := szuri.ParsePort(tt.give)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
