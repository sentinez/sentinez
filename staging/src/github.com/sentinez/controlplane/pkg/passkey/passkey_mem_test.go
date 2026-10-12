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

package passkey

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTakeRegistration_OnlyOnce(t *testing.T) {
	store := NewMemoryStorage(time.Minute)
	give := &Registration{Email: "bob@x.io", Handle: "handle-1"}
	store.SaveRegistration("sid", give)

	got, ok := store.TakeRegistration("sid")
	require.True(t, ok)
	assert.Equal(t, give, got)

	_, ok = store.TakeRegistration("sid")
	assert.False(t, ok)
}
