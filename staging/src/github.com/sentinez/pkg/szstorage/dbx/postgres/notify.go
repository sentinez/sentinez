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

package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/sentinez/pkg/szstorage/dbx"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// NotifyFunction the trigger function of NotifyChange.
const NotifyFunction = "sz_notify_change"

// _identifier what a channel and a table name of the DDL may be made of;
// postgres folds the unquoted ones to lower case.
var _identifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.]{0,62}$`)

// Change is the payload of the notification of a row change.
type Change struct {
	Table string `json:"t"`
	// Op is the TG_OP: INSERT, UPDATE or DELETE.
	Op string `json:"op"`
	Id string `json:"id"`
}

// NotifyChange makes the table tableName notify channel with a Change
// after every insert, update and delete of a row. It is idempotent, and
// the payload holds the id only, as a notification is limited to 8000
// bytes: the listener reads the row back.
func NotifyChange(ctx context.Context, client Client,
	tableName, channel string) error {

	if !_identifier.MatchString(tableName) ||
		!_identifier.MatchString(channel) {
		return fmt.Errorf("notify %s: invalid name or channel %q",
			tableName, channel)
	}

	// no arguments: the simple protocol, several statements, one
	// implicit transaction.
	sql := strings.Join(notifyStmts(tableName, channel), ";")
	if _, err := client.Exec(ctx, sql); err != nil {
		return fmt.Errorf("notify %s: %w", tableName, err)
	}

	return nil
}

func notifyStmts(tableName, channel string) []string {
	trigger := strings.ReplaceAll(tableName, ".", "_") + "_notify"

	return []string{
		`CREATE OR REPLACE FUNCTION ` + NotifyFunction + `() RETURNS trigger ` +
			`LANGUAGE plpgsql AS $$ BEGIN PERFORM pg_notify(TG_ARGV[0], ` +
			`json_build_object('t', TG_TABLE_NAME, 'op', TG_OP, 'id', ` +
			`CASE WHEN TG_OP = 'DELETE' THEN OLD.id ELSE NEW.id END)::text); ` +
			`RETURN NULL; END $$`,
		`DROP TRIGGER IF EXISTS ` + trigger + ` ON ` + tableName,
		`CREATE TRIGGER ` + trigger + ` AFTER INSERT OR UPDATE OR DELETE ON ` +
			tableName + ` FOR EACH ROW EXECUTE FUNCTION ` + NotifyFunction +
			`('` + channel + `')`,
	}
}

// NewNotifying is New whose table also notifies channel of its changes, see
// [NotifyChange].
func NewNotifying[T any](ctx context.Context, conf *settingpb.Config,
	schema dbx.Schema, channel string) (dbx.Database[T], error) {

	return Open[T](ctx, conf, schema,
		func(ctx context.Context, client Client, tableName string) error {
			if err := syncSchema(ctx, client, tableName, schema); err != nil {
				return err
			}

			return NotifyChange(ctx, client, tableName, channel)
		})
}

// Subscription receives the [Change]s of a channel, on a connection of its
// own: a connection that listens cannot go back to the pool.
type Subscription struct {
	conn *pgx.Conn
}

// Listen starts listening to channel on a connection taken from the pool
// shared by the tables.
func Listen(ctx context.Context, conf *settingpb.Config,
	channel string) (*Subscription, error) {

	if !_identifier.MatchString(channel) {
		return nil, fmt.Errorf("listen: invalid channel %q", channel)
	}

	pool, err := getConnPool(conf)
	if err != nil {
		return nil, err
	}

	acquired, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("listen: acquire: %w", err)
	}

	conn := acquired.Hijack()
	if _, err := conn.Exec(ctx, "LISTEN "+channel); err != nil {
		_ = conn.Close(ctx)

		return nil, fmt.Errorf("listen %s: %w", channel, err)
	}

	return &Subscription{conn: conn}, nil
}

// Next blocks until a change arrives or ctx ends; the error of a lost
// connection means the subscription is over, and changes may have been
// missed.
func (s *Subscription) Next(ctx context.Context) (Change, error) {
	n, err := s.conn.WaitForNotification(ctx)
	if err != nil {
		return Change{}, fmt.Errorf("wait for notification: %w", err)
	}

	var c Change
	if err := json.Unmarshal([]byte(n.Payload), &c); err != nil {
		return Change{}, fmt.Errorf("notification %q: %w", n.Payload, err)
	}

	return c, nil
}

// Close ends the subscription and its connection.
func (s *Subscription) Close(ctx context.Context) error {
	return s.conn.Close(ctx)
}
