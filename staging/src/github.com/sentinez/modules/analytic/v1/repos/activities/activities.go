// Copyright 2025 Duc-Hung Ho.
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

package activitiesrepo

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/core/storage/dbx"
	"github.com/sentinez/core/storage/dbx/postgres"
	"github.com/sentinez/core/storage/utils/table"
	"github.com/sentinez/core/tables"
	analyticpb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/analytic/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/shared/rand"
	"github.com/sentinez/shared/zlog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	_ IActivity = (*Activities)(nil)
)

// nolint
type IActivity interface {
	Create(ctx context.Context, activity *analyticpb.Activity) (*analyticpb.Activity, error)
	Update(ctx context.Context, activity *analyticpb.Activity) error
	Get(ctx context.Context, id string) (*analyticpb.Activity, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) IActivity

	List(ctx context.Context, req *analyticpb.ListActivitiesRequest) (*analyticpb.ListActivitiesResponse, error)
}

func New(ctx context.Context, appConf *confpb.Config) (IActivity, error) {
	storage, err := postgres.New[analyticpb.Activity](ctx, appConf,
		dbx.WithTable(tables.AnalyticActivities),
		dbx.WithColumns(dbx.ColumnM{
			analyticpb.Activity_Id:            postgres.String,
			analyticpb.Activity_ResourceId:    postgres.String,
			analyticpb.Activity_UniqueVisitor: postgres.IntArr,
		}),
	)
	if err != nil {
		return nil, err
	}

	return &Activities{
		storage: storage,
	}, nil
}

type Activities struct {
	storage dbx.Database[analyticpb.Activity]
}

func (a *Activities) WithTX(tx *postgres.TxSession) IActivity {
	return &Activities{
		storage: postgres.WithTx(tx, a.storage),
	}
}

// nolint:funlen
func buildListQuery(builder sq.SelectBuilder,
	req *analyticpb.ListActivitiesRequest) sq.SelectBuilder {
	for _, id := range req.GetIds() {
		builder = builder.Where(sq.Eq{analyticpb.Activity_Id: id})
	}

	for _, resourceID := range req.GetResourceIds() {
		builder = builder.Where(
			sq.Eq{analyticpb.Activity_ResourceId: resourceID})
	}

	return builder
}

// List implements IActivity.
// nolint:funlen
func (a *Activities) List(ctx context.Context,
	req *analyticpb.ListActivitiesRequest,
) (*analyticpb.ListActivitiesResponse, error) {

	builder := postgres.SelectBuilder(a.storage, req.GetPage())
	builder = buildListQuery(builder, req)

	var total int64

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	zlog.Debug("[activities] query: ", query, " args: ", args)

	activities, err := a.storage.CollectRows(ctx, builder, scan)
	if err != nil {
		return nil, err
	}

	if req.GetPage().GetTotal() {
		total, err = a.storage.Total(ctx)
		if err != nil {
			return nil, err
		}
	}

	return &analyticpb.ListActivitiesResponse{
		Activities: activities, Total: total}, nil
}

// Create implements IActivity.
func (a *Activities) Create(ctx context.Context,
	activity *analyticpb.Activity) (*analyticpb.Activity, error) {

	activity.Id = rand.NewID(table.NewPrimaryKey(tables.AnalyticActivities))
	query := postgres.InsertBuilder(a.storage, postgres.M{
		analyticpb.Activity_Id:            activity.GetId(),
		analyticpb.Activity_ResourceId:    activity.GetResourceId(),
		analyticpb.Activity_UniqueVisitor: activity.GetUniqueVisitor(),
	})

	if _, err := a.storage.Insert(ctx, query); err != nil {
		return nil, err
	}

	return activity, nil
}

// Delete implements IActivity.
func (a *Activities) Delete(ctx context.Context, id string) error {
	return a.storage.Delete(ctx, id)
}

// Get implements IActivity.
func (a *Activities) Get(ctx context.Context,
	id string) (*analyticpb.Activity, error) {
	builder := a.selectQuery(nil).Where(sq.Eq{analyticpb.Activity_Id: id})
	return a.storage.Select(ctx, builder, scanOne)
}

// Update implements IActivity.
func (a *Activities) Update(ctx context.Context,
	activity *analyticpb.Activity) error {

	query := postgres.UpdateBuilder(a.storage, activity.GetId())

	if activity.GetResourceId() != "" {
		query = query.Set(
			analyticpb.Activity_ResourceId, activity.GetResourceId())
	}

	if len(activity.GetUniqueVisitor()) > 0 {
		query = query.Set(
			analyticpb.Activity_UniqueVisitor, activity.GetUniqueVisitor())
	}

	_, err := a.storage.Exec(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activities) selectQuery(page *typepb.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(a.storage, page,
		analyticpb.Activity_Id,
		analyticpb.Activity_ResourceId,
		analyticpb.Activity_UniqueVisitor,
		dbx.FieldCreatedAt,
		dbx.FieldUpdatedAt,
	)
}

func scan(rows dbx.Rows) ([]*analyticpb.Activity, error) {
	var activities []*analyticpb.Activity

	for rows.Next() {
		activity, err := scanOne(rows)
		if err != nil {
			return nil, err
		}

		activities = append(activities, activity)
	}

	return activities, rows.Err()
}

func scanOne(row dbx.Row) (*analyticpb.Activity, error) {
	var (
		createdAt, updatedAt time.Time
		activity             analyticpb.Activity
	)

	err := row.Scan(
		&activity.Id,
		&activity.ResourceId,
		&activity.UniqueVisitor,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	activity.Metadata = &typepb.Metadata{
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(updatedAt),
	}

	return &activity, nil
}
