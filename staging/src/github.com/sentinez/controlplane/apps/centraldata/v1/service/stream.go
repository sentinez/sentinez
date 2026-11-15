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

package centraldatasvc

import (
	"context"
	"fmt"
	"slices"

	"github.com/sentinez/pkg/szerror"
	"github.com/sentinez/pkg/sztable"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Sync streams every row of the requested tables, then a Snapshot: the
// version and epoch the rows are those of. The changes published after
// that version are not in them.
func (srv *CentralDataService) Sync(ctx context.Context,
	req *centraldatapb.SyncRequest,
	send func(*centraldatapb.SyncResponse) error) error {

	tables, err := srv.requested(req.GetTables())
	if err != nil {
		return err
	}

	version, epoch, err := srv.memory.Snapshot(tables,
		func(table sztable.Table, row any) error {
			if err := ctx.Err(); err != nil {
				return err
			}

			return sendRow(table, row, send)
		})
	if err != nil {
		return err
	}

	return send(&centraldatapb.SyncResponse{
		Kind: &centraldatapb.SyncResponse_Done{
			Done: &centraldatapb.Snapshot{Version: version, Epoch: epoch},
		},
	})
}

// requested the tables of names, all those synced when none is named.
func (srv *CentralDataService) requested(
	names []string) ([]sztable.Table, error) {

	if len(names) == 0 {
		return srv.tables, nil
	}

	tables := make([]sztable.Table, 0, len(names))
	for _, name := range names {
		table := sztable.Table(name)
		if !slices.Contains(srv.tables, table) {
			return nil, szerror.StatusInvalidArgumentF(
				"table %q is not synced", name)
		}

		tables = append(tables, table)
	}

	return tables, nil
}

func sendRow(table sztable.Table, row any,
	send func(*centraldatapb.SyncResponse) error) error {

	msg, ok := row.(proto.Message)
	if !ok {
		return fmt.Errorf("%s holds %T", table, row)
	}

	change, err := upsertChange(table, msg)
	if err != nil {
		return err
	}

	return send(&centraldatapb.SyncResponse{
		Kind: &centraldatapb.SyncResponse_Row{Row: change},
	})
}

// upsertChange the UPSERT of row, without version nor epoch.
func upsertChange(table sztable.Table,
	row proto.Message) (*centraldatapb.Change, error) {

	packed, err := anypb.New(row)
	if err != nil {
		return nil, fmt.Errorf("pack %s: %w", table, err)
	}

	return &centraldatapb.Change{
		Table: table.String(),
		Op:    centraldatapb.Change_OP_UPSERT,
		Id:    idOf(row),
		Row:   packed,
	}, nil
}
