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

// Package eventfilter filters the lists of the event repos.
package eventfilter

import (
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

// Columns the columns of an event table the Filter applies to.
type Columns struct {
	ResourceId  string
	MatchedTime string
}

// Filter selects the events of a list, latest matched first; a zero
// field does not filter.
type Filter struct {
	ResourceIds []string
	// From and To bound the matched time, From inclusive, To exclusive.
	From time.Time
	To   time.Time
	Page *typepb.Pages
}

// Apply adds the conditions, order and paging of f, nil included, to b.
func (f *Filter) Apply(b sq.SelectBuilder, cols Columns) sq.SelectBuilder {
	if f == nil {
		return b.OrderBy(cols.MatchedTime + " DESC")
	}

	if len(f.ResourceIds) > 0 {
		b = b.Where(sq.Eq{cols.ResourceId: f.ResourceIds})
	}
	if !f.From.IsZero() {
		b = b.Where(sq.GtOrEq{cols.MatchedTime: f.From.UnixMilli()})
	}
	if !f.To.IsZero() {
		b = b.Where(sq.Lt{cols.MatchedTime: f.To.UnixMilli()})
	}

	return postgres.LimitOffset(b.OrderBy(cols.MatchedTime+" DESC"), f.Page)
}
