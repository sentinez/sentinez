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

package topvalues

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	"github.com/sentinez/pkg/sztable"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// TopValuer counts the most frequent values of the fields of the events.
type TopValuer interface {
	// TopValues counts the limit most frequent values of each field of
	// names, all of them when empty, among the events of filter (but its
	// page), and the events; a name it does not count is an
	// eventfilter.ErrUnknownField, as a match of the filter on one.
	TopValues(ctx context.Context, filter *eventfilter.Filter,
		names []string, limit uint64) (*Result, error)
}

var _ TopValuer = (*Reader)(nil)

// Reader the TopValuer of an event table.
type Reader struct {
	db   dbx.Database[Row]
	cols eventfilter.Columns
}

// Open returns the Reader of the event table, which must exist (synced by
// its repo), filtered on cols, counting their Fields (in their order when
// the names are empty).
func Open(ctx context.Context, conf *settingpb.Config, table sztable.Table,
	cols eventfilter.Columns) (*Reader, error) {
	db, err := postgres.Open[Row](ctx, conf,
		dbx.Schema{Table: table}, postgres.NoSync)
	if err != nil {
		return nil, err
	}

	return &Reader{db: db, cols: cols}, nil
}

// TopValues implements TopValuer.
func (r *Reader) TopValues(ctx context.Context, filter *eventfilter.Filter,
	names []string, limit uint64) (*Result, error) {
	fields, err := pick(r.cols, names)
	if err != nil {
		return nil, err
	}

	b, err := filter.Where(sq.Select().From(r.db.Table()), r.cols)
	if err != nil {
		return nil, err
	}

	q, err := query(b, fields, limit)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.CollectRows(ctx, q, collect)
	if err != nil {
		return nil, err
	}

	return group(rows, fields), nil
}
