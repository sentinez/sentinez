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

package eventsrepo

import (
	"context"
	"testing"

	analyticpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	"github.com/stretchr/testify/assert"
)

func TestTopValuesRejectsKind(t *testing.T) {
	tests := []analyticpb.EventKind{
		analyticpb.EventKind_EVENT_KIND_UNSPECIFIED,
		analyticpb.EventKind(99),
	}

	for _, kind := range tests {
		t.Run(kind.String(), func(t *testing.T) {
			var events Events

			_, err := events.TopValues(context.Background(), kind, nil, nil, 5)

			assert.Error(t, err)
		})
	}
}
