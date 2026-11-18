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

// Package auditchange carries what an audited update changed, from the
// service that writes it to its audit log: the request names only the
// new values, the service reads the old ones from the database.
package auditchange

import (
	"context"
	"slices"
	"sync"

	"google.golang.org/protobuf/proto"
)

type changeKey struct{}

// Change a message an update wrote, before and after, and the fields it
// wrote, named as in the .proto.
type Change struct {
	mu            sync.Mutex
	before, after proto.Message
	paths         []string
}

// With returns ctx carrying a new Change, which Set fills, and it.
func With(ctx context.Context) (context.Context, *Change) {
	change := &Change{}
	return context.WithValue(ctx, changeKey{}, change), change
}

// Set records on the Change of ctx that an update wrote the fields paths
// of a message, before and after it, which are messages of the same type
// it must not change anymore; nothing without a Change.
func Set(ctx context.Context, before, after proto.Message, paths []string) {
	change, ok := ctx.Value(changeKey{}).(*Change)
	if !ok {
		return
	}

	change.mu.Lock()
	defer change.mu.Unlock()

	change.before, change.after = before, after
	change.paths = slices.Clone(paths)
}

// Get returns what Set recorded on c, nil messages when nothing.
func (c *Change) Get() (before, after proto.Message, paths []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.before, c.after, c.paths
}
