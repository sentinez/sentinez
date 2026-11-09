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

package example

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var errRows = errors.New("rows")

// fakeRow scans the value of each column, named by columns, into the
// dest at its index.
type fakeRow struct {
	columns []string
	values  map[string]any
}

func (r fakeRow) Scan(dest ...any) error {
	if len(dest) != len(r.columns) {
		return fmt.Errorf("scan %d dest, want %d", len(dest), len(r.columns))
	}
	for i, col := range r.columns {
		v, ok := r.values[col]
		if !ok {
			continue
		}
		setDest(reflect.ValueOf(dest[i]).Elem(), reflect.ValueOf(v))
	}
	return nil
}

// setDest sets dst to val, a pointer to it when dst is a pointer, as the
// driver scans a value into a nullable destination.
func setDest(dst, val reflect.Value) {
	if dst.Kind() == reflect.Pointer && val.Type() != dst.Type() {
		ptr := reflect.New(val.Type())
		ptr.Elem().Set(val)
		val = ptr
	}
	dst.Set(val)
}

// fakeRows yields rows, then err.
type fakeRows struct {
	rows []fakeRow
	err  error
	next int
}

func (r *fakeRows) Next() bool {
	r.next++
	return r.next <= len(r.rows)
}

func (r *fakeRows) Scan(dest ...any) error {
	return r.rows[r.next-1].Scan(dest...)
}

func (r *fakeRows) Err() error { return r.err }

func TestScanOneScanned(t *testing.T) {
	at := time.Date(2026, 11, 9, 1, 2, 3, 0, time.UTC)
	attrs, err := structpb.NewStruct(map[string]any{"k": "v"})
	require.NoError(t, err)
	parent := &Scanned{Id: "parent"}
	priority := int32(7)
	note := "note"
	level := typepb.Status(2)
	snapshot := &Scanned_Audit{Actor: "snap", At: timestamppb.New(at)}

	row := fakeRow{
		columns: ScannedColumns(),
		values: map[string]any{
			Scanned_OwnerCreatedAt:  at,
			Scanned_OwnerUpdatedAt:  at.Add(time.Hour),
			Scanned_OwnerResourceId: "rsrc_1",
			Scanned_OwnerTenantId:   "tnt_1",
			Scanned_Id:              "id_1",
			Scanned_Status:          typepb.Status(1),
			Scanned_Tags:            []string{"a", "b"},
			Scanned_AuditActor:      "actor",
			Scanned_AuditAt:         at,
			Scanned_AuditParent:     mustMarshal(t, parent),
			Scanned_Ttl:             time.Minute,
			Scanned_Attrs:           mustMarshal(t, attrs),
			Scanned_Priority:        &priority,
			Scanned_Note:            &note,
			Scanned_Snapshot:        mustMarshal(t, snapshot),
			Scanned_AuditJson:       mustMarshalJSON(t, snapshot),
			Scanned_Level:           &level,
		},
	}

	got, err := ScanOneScanned(row)
	require.NoError(t, err)

	want := &Scanned{
		Owner: &typepb.Owner{
			CreatedAt:  timestamppb.New(at),
			UpdatedAt:  timestamppb.New(at.Add(time.Hour)),
			ResourceId: "rsrc_1",
			TenantId:   "tnt_1",
		},
		Id:     "id_1",
		Status: typepb.Status(1),
		Tags:   []string{"a", "b"},
		Audit: &Scanned_Audit{
			Actor:  "actor",
			At:     timestamppb.New(at),
			Parent: parent,
		},
		Ttl:       durationpb.New(time.Minute),
		Attrs:     attrs,
		Priority:  &priority,
		Note:      note,
		Snapshot:  snapshot,
		AuditJson: snapshot,
		Level:     level,
	}
	assert.Empty(t, cmpDiff(want, got))
}

func TestScanOneScannedLeavesNullsZero(t *testing.T) {
	row := fakeRow{columns: ScannedColumns(), values: map[string]any{}}

	got, err := ScanOneScanned(row)
	require.NoError(t, err)

	assert.Nil(t, got.GetAudit().GetParent())
	assert.Nil(t, got.GetAttrs())
	assert.Nil(t, got.GetSnapshot())
	assert.Nil(t, got.GetAuditJson())
	assert.Empty(t, got.GetNote())
	assert.Zero(t, got.GetLevel())
}

func TestScannedColumnsLeaveIgnoredOut(t *testing.T) {
	assert.NotContains(t, ScannedColumns(), Scanned_Preview)
}

func TestScanScanned(t *testing.T) {
	tests := []struct {
		name    string
		giveIds []string
		giveErr error
		wantIds []string
		wantErr error
	}{
		{name: "no rows"},
		{
			name:    "rows",
			giveIds: []string{"a", "b"},
			wantIds: []string{"a", "b"},
		},
		{name: "rows error", giveErr: errRows, wantErr: errRows},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := &fakeRows{err: tt.giveErr}
			for _, id := range tt.giveIds {
				rows.rows = append(rows.rows, fakeRow{
					columns: ScannedColumns(),
					values:  map[string]any{Scanned_Id: id},
				})
			}

			got, err := ScanScanned(rows)

			require.ErrorIs(t, err, tt.wantErr)
			var gotIds []string
			for _, m := range got {
				gotIds = append(gotIds, m.GetId())
			}
			assert.Equal(t, tt.wantIds, gotIds)
		})
	}
}

func mustMarshal(t *testing.T, m proto.Message) []byte {
	t.Helper()
	b, err := proto.Marshal(m)
	require.NoError(t, err)
	return b
}

func mustMarshalJSON(t *testing.T, m proto.Message) []byte {
	t.Helper()
	b, err := protojson.Marshal(m)
	require.NoError(t, err)
	return b
}

func cmpDiff(want, got proto.Message) string {
	if proto.Equal(want, got) {
		return ""
	}
	return fmt.Sprintf("want %v\ngot  %v", want, got)
}

func TestScannedValuesRoundTrip(t *testing.T) {
	at := time.Date(2026, 11, 9, 1, 2, 3, 0, time.UTC)
	priority := int32(3)
	give := &Scanned{
		Owner: &typepb.Owner{
			CreatedAt:  timestamppb.New(at),
			UpdatedAt:  timestamppb.New(at),
			ResourceId: "rsrc_1",
		},
		Id:        "id_1",
		Tags:      []string{"a"},
		Audit:     &Scanned_Audit{Actor: "actor", At: timestamppb.New(at)},
		Ttl:       durationpb.New(time.Second),
		Priority:  &priority,
		Snapshot:  &Scanned_Audit{Actor: "snap"},
		AuditJson: &Scanned_Audit{Actor: "json"},
	}

	values, err := ScannedValues(give)
	require.NoError(t, err)
	require.Len(t, values, len(ScannedColumns()))

	row := fakeRow{columns: ScannedColumns(), values: map[string]any{}}
	for i, col := range ScannedColumns() {
		if values[i] != nil {
			row.values[col] = values[i]
		}
	}

	got, err := ScanOneScanned(row)
	require.NoError(t, err)
	assert.Empty(t, cmpDiff(give, got))
}
