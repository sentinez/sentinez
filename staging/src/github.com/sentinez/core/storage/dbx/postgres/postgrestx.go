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
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/sentinez/core/storage/dbx"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	ssync "github.com/sentinez/shared/sync"
)

var (
	_        dbx.TxSession = (*TxSession)(nil)
	onceTxSS sync.Once
	txSSPool *ssync.Pool[TxSession]
)

func NewTX(conf *confpb.Config) *Tx {
	return &Tx{conf: conf.GetEnv()}
}

func WithTx[T any](ss *TxSession, db dbx.Database[T]) dbx.Database[T] {
	return &postgres[T]{client: ss.tx, tableName: db.Table(), tx: true}
}

type Tx struct {
	conf *confpb.EnvConfig
	mock pgx.Tx
}

func (t *Tx) Begin(ctx context.Context) (*TxSession, error) {
	conn, err := getConnPool(t.conf)
	if err != nil {
		return nil, err
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, err
	}

	onceTxSS.Do(func() {
		txSSPool = ssync.NewPool[TxSession]()
	})

	txSS := txSSPool.Get()
	txSS.tx = tx

	return txSS, nil
}

func NewTXMock(tx pgx.Tx) *Tx {
	return &Tx{mock: tx}
}

type TxSession struct {
	tx pgx.Tx
}

func (ts *TxSession) release() {
	ts.tx = nil
	txSSPool.Put(ts)
}

// Commit implements database.Transaction.
func (ts *TxSession) Commit(ctx context.Context) error {
	defer ts.release()

	if err := ts.tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// Rollback implements database.Transaction.
func (ts *TxSession) Rollback(ctx context.Context) error {
	defer ts.release()

	if err := ts.tx.Rollback(ctx); err != nil {
		return err
	}

	return nil
}
