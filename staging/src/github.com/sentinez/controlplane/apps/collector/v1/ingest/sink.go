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

	crspb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	"github.com/sentinez/shared/szlog"
	"github.com/sentinez/shared/szprefix"
	"github.com/twmb/franz-go/pkg/kgo"
)

// _insertRows caps the rows of an insert statement, far below the 65535
// parameters of a postgres statement.
const _insertRows = 500

// Sink inserts the records of a topic in its table.
type Sink interface {
	Insert(ctx context.Context, records []*kgo.Record) error
}

// CreateFn inserts events, skipping the ones stored already.
type CreateFn[E any] func(ctx context.Context, events ...*E) error

// NewRuleEventSink returns the sink of the secrulepb.RuleEvent records,
// inserted by create.
func NewRuleEventSink(create CreateFn[secrulepb.RuleEvent]) Sink {
	return &sink[secrulepb.RuleEvent]{
		decode: func(r *kgo.Record) (*secrulepb.RuleEvent, error) {
			var event secrulepb.RuleEvent
			if err := event.UnmarshalVT(r.Value); err != nil {
				return nil, err
			}
			if event.GetId() == "" {
				event.Id = recordId(szprefix.SecRuleEvent, r)
			}

			return &event, nil
		},
		create: create,
	}
}

// NewCoreRuleEventSink returns the sink of the corerulesetpb.CoreRuleEvent
// records, inserted by create.
func NewCoreRuleEventSink(create CreateFn[crspb.CoreRuleEvent]) Sink {
	return &sink[crspb.CoreRuleEvent]{
		decode: func(r *kgo.Record) (*crspb.CoreRuleEvent, error) {
			var event crspb.CoreRuleEvent
			if err := event.UnmarshalVT(r.Value); err != nil {
				return nil, err
			}
			if event.GetId() == "" {
				event.Id = recordId(szprefix.CoreRulesetEvent, r)
			}

			return &event, nil
		},
		create: create,
	}
}

type sink[E any] struct {
	decode func(r *kgo.Record) (*E, error)
	create CreateFn[E]
}

// Insert decodes records, dropping the malformed ones, and inserts them
// _insertRows at a time.
func (s *sink[E]) Insert(ctx context.Context, records []*kgo.Record) error {
	events := make([]*E, 0, len(records))
	for _, r := range records {
		event, err := s.decode(r)
		if err != nil {
			szlog.Warnf("ingest: drop %s[%d]@%d: %v",
				r.Topic, r.Partition, r.Offset, err)
			continue
		}
		events = append(events, event)
	}

	for chunk := range slices.Chunk(events, _insertRows) {
		if err := s.create(ctx, chunk...); err != nil {
			return fmt.Errorf("insert %d events: %w", len(chunk), err)
		}
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
