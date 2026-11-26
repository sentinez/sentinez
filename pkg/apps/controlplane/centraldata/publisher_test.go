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

package centraldata

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/sentinez/pkg/szuri"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func confOf(env map[string]string) *settingpb.Config {
	return &settingpb.Config{Env: env}
}

func TestNewPublisherWithoutAddresses(t *testing.T) {
	pub, err := NewPublisher(confOf(nil))

	require.NoError(t, err)
	assert.IsType(t, nopPublisher{}, pub)
	assert.NoError(t, pub.Publish(context.Background(), nil))
}

func TestNewPublisherNeedsAKey(t *testing.T) {
	_, err := NewPublisher(confOf(map[string]string{
		"SZ_OLRIC_URI": "olric://127.0.0.1:3320",
	}))

	require.ErrorIs(t, err, errNoSyncKey)
}

func TestNewPublisherRefusesAShortKey(t *testing.T) {
	short := base64.StdEncoding.EncodeToString([]byte("short"))
	_, err := NewPublisher(confOf(map[string]string{
		"SZ_OLRIC_URI": "olric://127.0.0.1:3320",
		"SZ_SYNC_KEY":  short,
	}))

	require.Error(t, err)
}

func TestNewPublisherRefusesABareAddress(t *testing.T) {
	key := base64.StdEncoding.EncodeToString(make([]byte, 32))
	_, err := NewPublisher(confOf(map[string]string{
		"SZ_OLRIC_URI": "127.0.0.1:3320",
		"SZ_SYNC_KEY":  key,
	}))

	require.ErrorIs(t, err, szuri.ErrNoScheme)
}
