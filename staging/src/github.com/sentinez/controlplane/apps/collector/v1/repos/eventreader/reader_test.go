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

package eventreader

import (
	"context"
	"testing"

	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// _event a row of the tests.
type _event struct{}

// _table a database of a table only: the tests never reach it.
type _table struct {
	dbx.Database[_event]
}

func (_table) Table() string { return "t" }

func newReader() *Reader[_event] {
	return &Reader[_event]{
		spec: Spec[_event]{
			Select: []string{"id", "status"},
			Columns: eventfilter.Columns{
				MatchedTime: "ts",
				Fields:      []eventfilter.Field{eventfilter.Column("path")},
			},
		},
		db: _table{},
	}
}

func TestReaderSelect(t *testing.T) {
	sql, _, err := newReader().Select().ToSql()

	require.NoError(t, err)
	assert.Equal(t, "SELECT id, status FROM t", sql)
}

func TestReaderListRejectsField(t *testing.T) {
	filter := &eventfilter.Filter{
		Matches: []eventfilter.Match{{Field: "headers", Value: "x"}},
	}

	_, err := newReader().List(context.Background(), filter)

	assert.ErrorIs(t, err, eventfilter.ErrUnknownField)
}

func TestReaderHistogramRejectsValue(t *testing.T) {
	events := &eventfilter.Filter{
		Matches: []eventfilter.Match{{Field: "headers", Value: "x"}},
	}

	_, err := newReader().Histogram(context.Background(), nil, events)

	assert.ErrorIs(t, err, eventfilter.ErrUnknownField)
}
