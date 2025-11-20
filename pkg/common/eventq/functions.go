// Copyright 2025 Duc-Hung Ho.
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

package eventq

import (
	"context"

	"github.com/sentinez/shared/zlog"
)

// QueueSpace is the size of the queue
const QueueSpace int = 2

var queue = New[string](QueueSpace)

// Subscribe to the event queue with a names and a handler function
func Subscribe(ctx context.Context, ns string, hdl func(value string) error) {
	go func() {
		ch := queue.Get(ns)

		for {
			select {
			case <-ctx.Done():
				zlog.Debugf("subscription to names %s stopped", ns)
				return
			case event, ok := <-ch:
				zlog.Debugf("received event of names: %s", ns)
				if !ok {
					zlog.Warnf("channel for names %s closed", ns)
					return
				}

				func() {
					defer func() {
						if r := recover(); r != nil {
							zlog.Errorf("panic in handler: %v", r)
						}
					}()

					if err := hdl(event); err != nil {
						zlog.Errorf("failed to handle event: %v", err)
					}
				}()
			}
		}
	}()
}

// Publish to the event queue with a names and a value
func Publish(ns string, value string) {
	ch := queue.Get(ns)

	select {
	case ch <- value:
		zlog.Debugf("published event to names: %s", ns)
	default:
		zlog.Warnf("channel for names %s is full, dropping event", ns)
	}
}
