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

package auditchange

import (
	"context"
	"testing"

	iampb "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	"github.com/stretchr/testify/assert"
)

func TestSetRecordsOnChange(t *testing.T) {
	ctx, change := With(context.Background())
	before := &iampb.User{FullName: "Bob"}
	after := &iampb.User{FullName: "Alice"}
	paths := []string{"full_name"}

	Set(ctx, before, after, paths)
	paths[0] = "email"

	gotBefore, gotAfter, gotPaths := change.Get()
	assert.Same(t, before, gotBefore)
	assert.Same(t, after, gotAfter)
	assert.Equal(t, []string{"full_name"}, gotPaths, "paths copied")
}

func TestSetWithoutChange(t *testing.T) {
	assert.NotPanics(t, func() {
		Set(context.Background(), &iampb.User{}, &iampb.User{}, nil)
	})
}
