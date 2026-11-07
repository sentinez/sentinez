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

package ingest

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/sentinez/controlplane/pkg/pools"
	crspool "github.com/sentinez/controlplane/pkg/pools/coreruleset"
	httpegresspool "github.com/sentinez/controlplane/pkg/pools/httpegress"
	httpingresspool "github.com/sentinez/controlplane/pkg/pools/httpingress"
	secrulepool "github.com/sentinez/controlplane/pkg/pools/secrule"
	crspb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szprefix"
	"github.com/twmb/franz-go/pkg/kgo"
)

// _insertRows caps the rows of an insert statement, far below the 65535
// parameters of a postgres statement.
const _insertRows = 500

// Sink inserts the records of a topic in its table.
type Sink interface {
	Insert(ctx context.Context, records []*kgo.Record) error
}

// CreateFn inserts events, skipping the ones stored already. The events
// go back to their pool once it returns: it must not keep them.
type CreateFn[E any] func(ctx context.Context, events ...*E) error

// event is the pointer to an event message E, decoded from a record.
type event[E any] interface {
	pools.Message[E]
	UnmarshalVT(b []byte) error
	GetId() string
}

// NewSecRuleEventSink returns the sink of the secrulepb.SecRuleEvent records,
// inserted by create.
func NewSecRuleEventSink(create CreateFn[secrulepb.SecRuleEvent]) Sink {
	return newSink(secrulepool.Pool(), szprefix.SecRuleEvent, create,
		func(e *secrulepb.SecRuleEvent, id string) { e.Id = id })
}

// NewCoreRuleEventSink returns the sink of the corerulesetpb.CoreRuleEvent
// records, inserted by create.
func NewCoreRuleEventSink(create CreateFn[crspb.CoreRuleEvent]) Sink {
	return newSink(crspool.Pool(), szprefix.CoreRulesetEvent, create,
		func(e *crspb.CoreRuleEvent, id string) { e.Id = id })
}

// NewHttpIngressEventSink returns the sink of the httppb.IngressEvent
// records, inserted by create.
func NewHttpIngressEventSink(create CreateFn[httppb.IngressEvent]) Sink {
	return newSink(httpingresspool.Pool(), szprefix.Request, create,
		func(e *httppb.IngressEvent, id string) { e.Id = id })
}

// NewHttpEgressEventSink returns the sink of the httppb.EgressEvent
// records, inserted by create.
func NewHttpEgressEventSink(create CreateFn[httppb.EgressEvent]) Sink {
	return newSink(httpegresspool.Pool(), szprefix.Request, create,
		func(e *httppb.EgressEvent, id string) { e.Id = id })
}

// sink decodes the records in events of pool; an event without id gets
// the recordId of prefix, set by setId.
type sink[E any, P event[E]] struct {
	pool   *pools.Pool[E, P]
	prefix szprefix.Prefix
	setId  func(e P, id string)
	create CreateFn[E]
}

func newSink[E any, P event[E]](pool *pools.Pool[E, P],
	prefix szprefix.Prefix, create CreateFn[E],
	setId func(e P, id string)) *sink[E, P] {
	return &sink[E, P]{
		pool:   pool,
		prefix: prefix,
		setId:  setId,
		create: create,
	}
}

// Insert decodes records, dropping the malformed ones, and inserts them
// _insertRows at a time; the events go back to the pool after.
func (s *sink[E, P]) Insert(ctx context.Context, records []*kgo.Record) error {
	b := s.decode(records)
	defer b.Close()

	for chunk := range slices.Chunk(b.Items(), _insertRows) {
		if err := s.create(ctx, chunk...); err != nil {
			return fmt.Errorf("insert %d events: %w", len(chunk), err)
		}
	}

	return nil
}

// decode returns the batch of the events of records, each from the pool.
func (s *sink[E, P]) decode(records []*kgo.Record) *pools.Batch[E, P] {
	b := s.pool.NewBatch(len(records))
	for _, r := range records {
		e := s.pool.Acquire()
		if err := s.read(e, r); err != nil {
			szlog.Warnf("ingest: drop %s[%d]@%d: %v",
				r.Topic, r.Partition, r.Offset, err)
			s.pool.Release(e)
			continue
		}
		b.Add(e)
	}

	return b
}

// read decodes r in e, giving it the recordId of r when it has no id.
func (s *sink[E, P]) read(e *E, r *kgo.Record) error {
	p := P(e)
	if err := p.UnmarshalVT(r.Value); err != nil {
		return err
	}
	if p.GetId() == "" {
		s.setId(p, recordId(s.prefix, r))
	}

	return nil
}

// recordId the id of the event of r without one, the same each time r is
// consumed so inserting it again is a no-op.
func recordId(prefix szprefix.Prefix, r *kgo.Record) string {
	return prefix.Key() + r.Topic + "-" +
		strconv.FormatInt(int64(r.Partition), 10) + "-" +
		strconv.FormatInt(r.Offset, 10)
}
