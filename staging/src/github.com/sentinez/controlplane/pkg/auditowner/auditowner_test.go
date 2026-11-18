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

package auditowner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetKeepsTheFirst(t *testing.T) {
	ctx, owner := With(context.Background())

	Set(ctx, "tnt_1", "")
	Set(ctx, "tnt_2", "res_1")
	Set(ctx, "", "res_2")

	tenantId, resourceId := owner.Ids()
	assert.Equal(t, "tnt_1", tenantId)
	assert.Equal(t, "res_1", resourceId)
}

func TestSetWithoutOwner(t *testing.T) {
	assert.NotPanics(t, func() {
		Set(context.Background(), "tnt_1", "res_1")
	})
}
