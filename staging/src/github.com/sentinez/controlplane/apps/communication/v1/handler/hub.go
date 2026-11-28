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
	"sync"

	communicationpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/communication/v1"
)

// _subscriberBuffer is the number of notifications a subscriber may lag
// behind before the next ones are dropped for it.
const _subscriberBuffer = 16

// Hub fans the published notifications out to the open streams, each
// those its filter matches. A slow stream loses notifications rather than
// block the publisher.
type Hub struct {
	mu   sync.Mutex
	subs map[chan *communicationpb.Notification]filter
}

// filter tells whether a subscriber receives a notification.
type filter func(*communicationpb.Notification) bool

// NewHub returns a hub without subscribers.
func NewHub() *Hub {
	return &Hub{subs: make(map[chan *communicationpb.Notification]filter)}
}

// Publish sends n to every subscriber whose filter matches it. It never
// blocks.
func (h *Hub) Publish(n *communicationpb.Notification) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for ch, match := range h.subs {
		if !match(n) {
			continue
		}

		select {
		case ch <- n:
		default:
		}
	}
}

// Subscribe returns the channel of the notifications published from now
// on that match, and the function that ends the subscription.
func (h *Hub) Subscribe(
	match filter) (<-chan *communicationpb.Notification, func()) {
	ch := make(chan *communicationpb.Notification, _subscriberBuffer)

	h.mu.Lock()
	h.subs[ch] = match
	h.mu.Unlock()

	return ch, func() {
		h.mu.Lock()
		delete(h.subs, ch)
		h.mu.Unlock()
	}
}

// tenantFilter matches the notifications of tenantIds; of every tenant
// when there are none.
func tenantFilter(tenantIds ...string) filter {
	if len(tenantIds) == 0 {
		return func(*communicationpb.Notification) bool { return true }
	}

	set := make(map[string]struct{}, len(tenantIds))
	for _, id := range tenantIds {
		set[id] = struct{}{}
	}

	return func(n *communicationpb.Notification) bool {
		_, ok := set[n.GetTenantId()]
		return ok
	}
}
