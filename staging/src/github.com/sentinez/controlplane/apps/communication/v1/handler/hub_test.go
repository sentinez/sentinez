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

package communicationhdl

import (
	"testing"

	communicationpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/communication/v1"
	"github.com/stretchr/testify/assert"
)

func TestHubPublishesToMatchingSubscribers(t *testing.T) {
	hub := NewHub()
	a, unsubscribeA := hub.Subscribe(tenantFilter("tnt_a"))
	defer unsubscribeA()
	b, unsubscribeB := hub.Subscribe(tenantFilter("tnt_b"))
	defer unsubscribeB()

	n := &communicationpb.Notification{TenantId: "tnt_a"}
	hub.Publish(n)

	assert.Equal(t, n, <-a)
	assert.Empty(t, b)
}

func TestHubDropsForFullSubscriber(t *testing.T) {
	hub := NewHub()
	ch, unsubscribe := hub.Subscribe(tenantFilter())
	defer unsubscribe()

	// never blocks, even past the buffer
	for range _subscriberBuffer + 1 {
		hub.Publish(&communicationpb.Notification{})
	}

	assert.Len(t, ch, _subscriberBuffer)
}

func TestHubUnsubscribe(t *testing.T) {
	hub := NewHub()
	ch, unsubscribe := hub.Subscribe(tenantFilter())
	unsubscribe()

	hub.Publish(&communicationpb.Notification{})

	assert.Empty(t, ch)
}
