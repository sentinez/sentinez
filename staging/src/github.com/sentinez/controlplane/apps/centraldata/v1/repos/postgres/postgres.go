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

// Package pgrepo reads from postgres the tables centraldata syncs, and makes
// them notify their changes.
package pgrepo

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/controlplane/pkg/schema"
	"github.com/sentinez/pkg/szerror"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	"github.com/sentinez/pkg/sztable"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"google.golang.org/protobuf/proto"
)

// _idColumn the primary key every table has.
const _idColumn = "id"

// Table is a synced table of the database.
type Table interface {
	// Name the table, ex: sztable.SecuritySecRules.
	Name() sztable.Table
	// Relation the name of the table in the database, as the
	// notifications carry it (lower case).
	Relation() string
	// List all the rows, as pointers to the model of the table.
	List(ctx context.Context) ([]proto.Message, error)
	// Get the row of id; szerror.ErrNotFound when there is none.
	Get(ctx context.Context, id string) (proto.Message, error)
}

// Defaults the tables synced when none is configured.
func Defaults() []sztable.Table {
	return []sztable.Table{
		sztable.SecuritySecRules,
		sztable.SecurityRateLimits,
		sztable.SecurityCoreRulesets,
		sztable.TenantResources,
		sztable.TenantCaCertificates,
	}
}

// Parse the comma-separated names of tables, [Defaults] when empty. It
// fails on a table that cannot be synced.
func Parse(list string) ([]sztable.Table, error) {
	if strings.TrimSpace(list) == "" {
		return Defaults(), nil
	}

	var tables []sztable.Table
	for _, name := range strings.Split(list, ",") {
		table := sztable.Table(strings.TrimSpace(name))
		if _, ok := _openers[table]; !ok {
			return nil, fmt.Errorf("table %q cannot be synced", table)
		}

		tables = append(tables, table)
	}

	return tables, nil
}

// Open the tables, creating them and their notify trigger on channel.
func Open(ctx context.Context, conf *settingpb.Config,
	tables []sztable.Table, channel string) ([]Table, error) {

	opened := make([]Table, 0, len(tables))
	for _, name := range tables {
		open, ok := _openers[name]
		if !ok {
			return nil, fmt.Errorf("table %q cannot be synced", name)
		}

		table, err := open(ctx, conf, channel)
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", name, err)
		}

		opened = append(opened, table)
	}

	return opened, nil
}

type opener func(context.Context, *settingpb.Config, string) (Table, error)

// _openers the tables that can be synced, each with the scanners generated
// for its model.
var _openers = map[sztable.Table]opener{
	sztable.SecuritySecRules: openerOf(schema.SecuritySecRules,
		securitypb.SecRuleColumns, securitypb.ScanSecRule,
		securitypb.ScanOneSecRule),
	sztable.SecurityRateLimits: openerOf(schema.SecurityRateLimits,
		securitypb.RateLimitColumns, securitypb.ScanRateLimit,
		securitypb.ScanOneRateLimit),
	sztable.SecurityCoreRulesets: openerOf(schema.SecurityCoreRulesets,
		securitypb.CoreRulesetColumns, securitypb.ScanCoreRuleset,
		securitypb.ScanOneCoreRuleset),
	sztable.TenantResources: openerOf(schema.TenantResources,
		tenantpb.ResourceColumns, tenantpb.ScanResource,
		tenantpb.ScanOneResource),
	sztable.TenantCaCertificates: openerOf(schema.TenantCaCertificates,
		tenantpb.CaCertificateColumns, tenantpb.ScanCaCertificate,
		tenantpb.ScanOneCaCertificate),
}

// openerOf the opener of the table of sch, whose model is T.
func openerOf[T any, P rowPtr[T]](sch func() dbx.Schema,
	columns func() []string, scan dbx.ScanFn[T],
	scanOne dbx.ScanOneFn[T]) opener {

	return func(ctx context.Context, conf *settingpb.Config,
		channel string) (Table, error) {

		return open[T, P](ctx, conf, channel, sch(), columns(), scan,
			scanOne)
	}
}

type rowPtr[T any] interface {
	*T
	proto.Message
}

// table is a Table of the model T.
type table[T any, P rowPtr[T]] struct {
	name     sztable.Table
	relation string
	storage  dbx.Database[T]
	columns  []string
	scan     dbx.ScanFn[T]
	scanOne  dbx.ScanOneFn[T]
}

func open[T any, P rowPtr[T]](ctx context.Context, conf *settingpb.Config,
	channel string, sch dbx.Schema, columns []string,
	scan dbx.ScanFn[T], scanOne dbx.ScanOneFn[T]) (Table, error) {

	storage, err := postgres.NewNotifying[T](ctx, conf, sch, channel)
	if err != nil {
		return nil, err
	}

	return &table[T, P]{
		name:     sch.Table,
		relation: strings.ToLower(storage.Table()),
		storage:  storage,
		columns:  columns,
		scan:     scan,
		scanOne:  scanOne,
	}, nil
}

func (t *table[T, P]) Name() sztable.Table { return t.name }

func (t *table[T, P]) Relation() string { return t.relation }

func (t *table[T, P]) List(ctx context.Context) ([]proto.Message, error) {
	rows, err := t.storage.CollectRows(ctx, t.selectQ(), t.scan)
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", t.name, err)
	}

	list := make([]proto.Message, 0, len(rows))
	for _, row := range rows {
		list = append(list, P(row))
	}

	return list, nil
}

func (t *table[T, P]) Get(ctx context.Context,
	id string) (proto.Message, error) {

	q := t.selectQ().Where(sq.Eq{_idColumn: id})

	row, err := t.storage.CollectOneRow(ctx, q, t.scanOne)
	if szerror.IsNoRows(err) {
		return nil, fmt.Errorf("%s %s: %w", t.name, id, szerror.ErrNotFound)
	}

	if err != nil {
		return nil, fmt.Errorf("get %s %s: %w", t.name, id, err)
	}

	return P(row), nil
}

func (t *table[T, P]) selectQ() sq.SelectBuilder {
	return postgres.SelectBuilder(t.storage, nil, t.columns...)
}
