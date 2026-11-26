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
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/olric-data/olric/config"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const _password = "s3cret"

var _key32 = base64.StdEncoding.EncodeToString(
	[]byte(strings.Repeat("k", 32)))

func TestSecure(t *testing.T) {
	conf, err := newConfig(&settingpb.Config{Env: clusteredWith(
		map[string]string{
			"SZ_OLRIC_PASSWORD": _password,
			"SZ_GOSSIP_KEY":     _key32,
		})}, func() {}, privateIPOf("10.0.0.2"))
	require.NoError(t, err)

	assert.Equal(t, _password, conf.Authentication.Password)
	assert.Equal(t, []byte(strings.Repeat("k", 32)),
		conf.MemberlistConfig.SecretKey)
}

func TestSecureUnset(t *testing.T) {
	conf, err := newConfig(&settingpb.Config{Env: _clustered},
		func() {}, privateIPOf("10.0.0.2"))
	require.NoError(t, err)

	assert.False(t, conf.Authentication.Enabled())
	assert.Nil(t, conf.MemberlistConfig.SecretKey)
}

func TestSecureAloneIgnoresGossipKey(t *testing.T) {
	conf, err := newConfig(&settingpb.Config{Env: map[string]string{
		"SZ_GOSSIP_KEY": "not base64",
	}}, func() {}, privateIPOf("10.0.0.2"))
	require.NoError(t, err)

	assert.Nil(t, conf.MemberlistConfig.SecretKey)
}

func TestGossipKeyError(t *testing.T) {
	tests := []struct {
		name string
		give string
	}{
		{name: "not base64", give: "not base64"},
		{
			name: "bad length",
			give: base64.StdEncoding.EncodeToString([]byte("short")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newConfig(&settingpb.Config{Env: clusteredWith(
				map[string]string{"SZ_GOSSIP_KEY": tt.give})},
				func() {}, privateIPOf("10.0.0.2"))
			require.Error(t, err)
		})
	}
}

// withPassword secures the Olric port of a test node.
func withPassword(conf *config.Config) {
	conf.Authentication = &config.Authentication{Password: _password}
}

// A node with a password: its own subscriber and a publisher holding the
// password get through, a publisher without it does not.
func TestPasswordGuardsThePort(t *testing.T) {
	c, addr, _ := startNodeWith(t, "edge", "", withPassword)
	sub := subscribeTo(t, c)
	ctx := context.Background()

	_, err := NewPublisher([]string{addr})
	require.Error(t, err, "no password")

	pub, err := NewPublisher([]string{addr}, WithPassword(_password))
	require.NoError(t, err)
	defer func() { _ = pub.Close(ctx) }()
	require.NoError(t, pub.Publish(ctx, "test.changes", []byte("hello")))

	assert.Equal(t, "hello", receive(t, sub))
}

// Two nodes sharing the password and the gossip key join, and a message
// published on one reaches the subscriber of the other: the nodes
// authenticate to each other.
func TestSecuredNodesJoin(t *testing.T) {
	secured := func(conf *config.Config) {
		withPassword(conf)
		conf.MemberlistConfig.SecretKey = []byte(strings.Repeat("k", 32))
	}
	a, addrA, memberA := startNodeWith(t, "edge-a", "", secured)
	b, _, _ := startNodeWith(t, "edge-b", memberA, secured)
	waitMembers(t, a, 2)
	sub := subscribeTo(t, b)
	ctx := context.Background()

	pub, err := NewPublisher([]string{addrA}, WithPassword(_password))
	require.NoError(t, err)
	defer func() { _ = pub.Close(ctx) }()
	require.NoError(t, pub.Publish(ctx, "test.changes", []byte("hello")))

	assert.Equal(t, "hello", receive(t, sub))
}
