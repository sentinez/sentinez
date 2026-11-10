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
	// RequestId the request id of the events: the id of an http event.
	RequestId   string
	ResourceId  string
	MatchedTime string
	// Fields those a Match may filter, and their top values count, in
	// this order.
	Fields []Field
}

// Filter selects the events of a list, latest matched first; a zero
// field does not filter.
type Filter struct {
	RequestIds  []string
	ResourceIds []string
	// Matches all of them, on fields of the Columns.
	Matches []Match
	// From and To bound the matched time, From inclusive, To exclusive.
	From time.Time
	To   time.Time
	Page *typepb.Pages
}

// Apply adds the conditions, order and paging of f, nil included, to b;
// a Match of a field cols does not have is an ErrUnknownField.
func (f *Filter) Apply(b sq.SelectBuilder,
	cols Columns) (sq.SelectBuilder, error) {
	b, err := f.Where(b, cols)
	if err != nil {
		return b, err
	}

	b = b.OrderBy(cols.MatchedTime + " DESC")
	if f == nil {
		return b, nil
	}

	return postgres.LimitOffset(b, f.Page), nil
}

// Where adds the conditions of f, nil included, to b: its requests,
// resources, matches and time range, not its page; a Match of a field cols
// does not have is an ErrUnknownField, one of a value it cannot hold an
// ErrInvalidValue.
func (f *Filter) Where(b sq.SelectBuilder,
	cols Columns) (sq.SelectBuilder, error) {
	conds, err := f.conds(cols)
	if err != nil {
		return b, err
	}

	for _, cond := range conds {
		b = b.Where(cond)
	}

	return b, nil
}

// Cond the conditions of f, nil included, as one: none, a nil Sqlizer,
// without any; the errors are those of Where.
func (f *Filter) Cond(cols Columns) (sq.Sqlizer, error) {
	conds, err := f.conds(cols)
	if err != nil || len(conds) == 0 {
		return nil, err
	}

	return sq.And(conds), nil
}

// conds the conditions of f, nil included: its matches, requests,
// resources and time range.
func (f *Filter) conds(cols Columns) ([]sq.Sqlizer, error) {
	if f == nil {
		return nil, nil
	}

	conds := make([]sq.Sqlizer, 0, len(f.Matches)+4)
	for _, m := range f.Matches {
		field, err := cols.Field(m.Field)
		if err != nil {
			return nil, err
		}

		cond, err := m.condition(field)
		if err != nil {
			return nil, err
		}
		conds = append(conds, cond)
	}

	return append(conds, f.scope(cols)...), nil
}

// scope the conditions of the requests, resources and time range of f.
func (f *Filter) scope(cols Columns) []sq.Sqlizer {
	var conds []sq.Sqlizer
	if len(f.RequestIds) > 0 {
		conds = append(conds, sq.Eq{cols.RequestId: f.RequestIds})
	}
	if len(f.ResourceIds) > 0 {
		conds = append(conds, sq.Eq{cols.ResourceId: f.ResourceIds})
	}
	if !f.From.IsZero() {
		conds = append(conds,
			sq.GtOrEq{cols.MatchedTime: f.From.UnixMilli()})
	}
	if !f.To.IsZero() {
		conds = append(conds, sq.Lt{cols.MatchedTime: f.To.UnixMilli()})
	}

	return conds
}
