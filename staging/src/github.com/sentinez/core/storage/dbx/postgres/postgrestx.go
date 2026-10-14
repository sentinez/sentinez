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
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	ssync "github.com/sentinez/shared/sync"
)

var (
	_         dbx.TxSession = (*TxSession)(nil)
	_onceTxSS sync.Once
	_txSSPool *ssync.Pool[TxSession]
)

func NewTX(conf *settingpb.Config) *Tx {
	return &Tx{conf: conf}
}

func WithTx[T any](ss *TxSession, db dbx.Database[T]) dbx.Database[T] {
	return &postgres[T]{client: ss.tx, tableName: db.Table(), tx: true}
}

type Tx struct {
	conf *settingpb.Config
	mock pgx.Tx
}

func (t *Tx) Begin(ctx context.Context) (*TxSession, error) {
	tx, err := t.begin(ctx)
	if err != nil {
		return nil, err
	}

	_onceTxSS.Do(func() {
		_txSSPool = ssync.NewPool[TxSession]()
	})

	txSS := _txSSPool.Get()
	txSS.tx = tx

	return txSS, nil
}

// begin starts a transaction on the pool, or returns the mock one.
func (t *Tx) begin(ctx context.Context) (pgx.Tx, error) {
	if t.mock != nil {
		return t.mock, nil
	}

	conn, err := getConnPool(t.conf)
	if err != nil {
		return nil, err
	}

	return conn.Begin(ctx)
}

// NewTXMock returns a Tx whose Begin hands out tx, for tests.
func NewTXMock(tx pgx.Tx) *Tx {
	return &Tx{mock: tx}
}

type TxSession struct {
	tx pgx.Tx
}

func (ts *TxSession) release() {
	ts.tx = nil
	_txSSPool.Put(ts)
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
