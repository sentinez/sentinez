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

// Package eventreader reads an event hypertable: the list of its events,
// their histogram and their top values, bound once to the table, its
// columns and its filters.
package eventreader

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/topvalues"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	"github.com/sentinez/pkg/szstorage/dbx/timescale"
	"github.com/sentinez/pkg/szstorage/dbx/timescale/histogram"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

var _ topvalues.TopValuer = (*Reader[struct{}])(nil)

// Spec describes the Reader of the events T of a hypertable.
type Spec[T any] struct {
	Hypertable timescale.Hypertable
	// Key an INTEGER column of the hypertable the histogram splits the
	// counts by, besides the resource, ex: the http status; none when
	// empty.
	Key string
	// Columns those the filters apply to; their Fields are the ones a
	// match filters and a top value counts, in this order.
	Columns eventfilter.Columns
	// Select the columns of an event, in the order Scan reads them: the
	// generated <Event>Columns().
	Select []string
	// Scan reads the rows of Select into events: the generated
	// Scan<Event>.
	Scan dbx.ScanFn[T]
	// ScanOne reads one row of Select into an event: the generated
	// ScanOne<Event>.
	ScanOne func(dbx.Row) (*T, error)
	// Values the values of an event, in the order of Select: the
	// generated <Event>Values.
	Values func(*T) ([]any, error)
	// IdColumn the column of the id of an event.
	IdColumn string
}

// Reader the events T of a hypertable, and their histogram and top
// values.
type Reader[T any] struct {
	spec      Spec[T]
	db        dbx.Database[T]
	histogram *histogram.Histogram
	topValues *topvalues.Reader
}

// New returns the Reader of spec, its hypertable and the views of its
// histogram synced on the way.
func New[T any](ctx context.Context, conf *settingpb.Config,
	spec Spec[T]) (*Reader[T], error) {
	db, err := timescale.New[T](ctx, conf, spec.Hypertable)
	if err != nil {
		return nil, err
	}

	hist, err := histogram.New(ctx, conf, histogram.Spec{
		Source: spec.Hypertable,
		Key:    spec.Key,
	})
	if err != nil {
		return nil, err
	}

	top, err := topvalues.Open(ctx, conf, spec.Hypertable.Schema.Table,
		spec.Columns)
	if err != nil {
		return nil, err
	}

	return &Reader[T]{spec: spec, db: db, histogram: hist, topValues: top}, nil
}

// WithTx returns r reading and writing through tx.
func (r *Reader[T]) WithTx(tx *postgres.TxSession) *Reader[T] {
	cp := *r
	cp.db = postgres.WithTx(tx, r.db)

	return &cp
}

// Select selects the columns of an event from the table.
func (r *Reader[T]) Select() sq.SelectBuilder {
	return sq.Select(r.spec.Select...).From(r.db.Table())
}

// Create inserts the events, which have ids, in one statement; an event
// whose id and time are stored already is skipped, so inserting a batch
// again is a no-op.
func (r *Reader[T]) Create(ctx context.Context, events ...*T) error {
	if len(events) == 0 {
		return nil
	}

	query, err := postgres.InsertRowsBuilder(r.db, r.spec.Select, events,
		r.spec.Values)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx,
		query.Suffix(r.spec.Hypertable.OnConflictDoNothing()))

	return err
}

// Get the event of id.
func (r *Reader[T]) Get(ctx context.Context, id string) (*T, error) {
	builder := r.Select().Where(sq.Eq{r.spec.IdColumn: id})

	return r.db.Select(ctx, builder, r.spec.ScanOne)
}

// Delete the event of id.
func (r *Reader[T]) Delete(ctx context.Context, id string) error {
	return r.db.Delete(ctx, id)
}

// List the events of filter, the latest first: its page of them. A match
// on a field the table does not have is an eventfilter.ErrUnknownField.
func (r *Reader[T]) List(ctx context.Context,
	filter *eventfilter.Filter) ([]*T, error) {
	builder, err := filter.Apply(r.Select(), r.spec.Columns)
	if err != nil {
		return nil, err
	}

	return r.db.CollectRows(ctx, builder, r.spec.Scan)
}

// Histogram counts the events per bucket, resource and key of filter (see
// histogram.Histogram.Points): from the views, or, when events is set,
// from the events it selects: its requests and matches, as the resources
// and the time range are those of filter.
func (r *Reader[T]) Histogram(ctx context.Context,
	filter *histogram.Filter,
	events *eventfilter.Filter) ([]*histogram.Point, error) {
	if events == nil {
		return r.histogram.Points(ctx, filter)
	}

	cond, err := events.Cond(r.spec.Columns)
	if err != nil {
		return nil, err
	}

	return r.histogram.PointsWhere(ctx, filter, cond)
}

// TopValues implements topvalues.TopValuer.
func (r *Reader[T]) TopValues(ctx context.Context,
	filter *eventfilter.Filter, names []string,
	limit uint64) (*topvalues.Result, error) {
	return r.topValues.TopValues(ctx, filter, names, limit)
}
