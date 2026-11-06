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

package pools_test

import (
	"io"
	"testing"

	"github.com/sentinez/controlplane/pkg/pools"
	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ io.Closer = (*pools.Batch[httppb.IngressEvent, *httppb.IngressEvent])(nil)

func TestBatchCloseResets(t *testing.T) {
	p := pools.New(httppb.IngressEventFromVTPool)
	b := p.NewBatch(2)

	e := p.Acquire()
	e.Id, e.Host = "req_a", "a.test"
	b.Add(e)
	require.Len(t, b.Items(), 1)

	require.NoError(t, b.Close())

	assert.Empty(t, b.Items())
	assert.Empty(t, e.GetId())
	assert.Empty(t, e.GetHost())
}

func TestReleaseKeepsCapacity(t *testing.T) {
	p := pools.New(secrulepb.RuleEventFromVTPool)

	e := p.Acquire()
	e.Messages = make([]string, 3, 8)
	p.Release(e)

	assert.Empty(t, e.GetMessages())
	assert.Equal(t, 8, cap(e.GetMessages()))
}
