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

// Package database provides the database interface.
package database

import (
	"context"

	"github.com/sentinez/sentinez/pkg/std/zlog"

	"github.com/sentinez/sentinez/pkg/std/errors"
)

// Repository provides the interface for the database.
type Repository[T any, ID comparable] interface {
	Create(ctx context.Context, entity T) (T, error)
	Get(ctx context.Context, id ID) (T, error)
	GetAll(ctx context.Context) ([]T, error)
	Update(ctx context.Context, id ID, entity T) (T, error)
	Delete(ctx context.Context, id ID) error
	Exists(ctx context.Context, id ID) (bool, error)
	Count(ctx context.Context) (int64, error)
}

// New create multi-layer database instance
func New[T any, ID comparable](searchLayer Repository[T, ID],
	storageLayer Repository[T, ID]) Repository[T, ID] {

	return &Database[T, ID]{
		search:  searchLayer,
		storage: storageLayer,
	}
}

// Database multilayer with postgres & elasticsearch
type Database[T any, ID comparable] struct {
	search  Repository[T, ID]
	storage Repository[T, ID]
}

// Create inserts a new entity into both storage (PostgresSQL) and
// search (Elasticsearch). If Elasticsearch fails after PostgresSQL succeeds,
func (db *Database[T, ID]) Create(ctx context.Context, entity T) (T, error) {
	var empty T
	if db.storage == nil {
		zlog.Debug("[db.Create] storage layer is nil")
		return empty, errors.F("db.Create err: storage is nil")
	}

	// insert into PostgresSQL first
	createdEntity, err := db.storage.Create(ctx, entity)
	if err != nil {
		zlog.Debugf("[db.Create] storage err: %v", err)
		return empty, errors.F("db.Create err: %v", err)
	}

	// if search layer is nil, return created entity
	if db.search == nil {
		zlog.Infof("[db.Create] search => pass")
		return createdEntity, nil
	}

	// insert into Elasticsearch
	// ignore error if exist
	_, err = db.search.Create(ctx, entity)
	if err != nil {
		zlog.Warnf("[db.Create] search err: %v", err)
	}

	return createdEntity, nil
}

// Update modifies an existing entity in both storage (PostgresSQL) and
// search (Elasticsearch).If Elasticsearch fails after PostgresSQL succeeds,
func (db *Database[T, ID]) Update(
	ctx context.Context, id ID, entity T) (T, error) {
	var empty T
	if db.storage == nil {
		zlog.Debug("[db.Update] storage layer is nil")
		return empty, errors.F("db.Update err: storage is nil")
	}

	// update in PostgresSQL
	updatedEntity, err := db.storage.Update(ctx, id, entity)
	if err != nil {
		zlog.Debugf("[db.Update] storage err: %v", err)
		return empty, errors.F("db.Update storage err: %v", err)
	}

	// if search layer is nil, return updated entity
	if db.search == nil {
		zlog.Info("[db.Update] search => pass")
		return updatedEntity, nil
	}

	// update in Elasticsearch if they are exist
	if existed, err := db.search.Exists(ctx, id); err == nil && existed {
		_, err = db.search.Update(ctx, id, updatedEntity)
		if err != nil {
			zlog.Warnf("[db.Update] search err: %v", err)
		}
	}

	return updatedEntity, nil
}

// Delete removes an entity from both storage (PostgresSQL) and search
// (Elasticsearch). If Elasticsearch fails after PostgresSQL succeeds,
func (db *Database[T, ID]) Delete(ctx context.Context, id ID) error {
	if db.storage == nil {
		zlog.Debug("[db.Delete] err: storage is nil")
		return errors.F("db.Delete err: storage is nil")
	}

	if db.search == nil {
		zlog.Info("[db.Delete] search => pass")
	}

	if db.search != nil {
		// delete from Elasticsearch, if error, return error
		if esErr := db.search.Delete(ctx, id); esErr != nil {
			zlog.Debugf("[db.Delete][search] err: %v", esErr)
			return errors.F("db.Delete search err: %v", esErr)
		}
	}

	// after delete from search, delete from storage
	if err := db.storage.Delete(ctx, id); err != nil {
		zlog.Debugf("[db.Delete][storage] err: %v", err)
		return errors.F("db.Delete storage err: %v", err)
	}

	return nil
}

// Get retrieves an entity by ID from the search layer (Elasticsearch).
func (db *Database[T, ID]) Get(ctx context.Context, id ID) (T, error) {
	var empty T

	// if storage layer is nil, return error
	if db.storage == nil {
		zlog.Debug("[db.Get] storage layer is nil")
		return empty, errors.F("db.Get err: storage is nil")
	}

	if db.search == nil {
		zlog.Info("[db.Get] search => pass")
	}

	if db.search != nil {
		result, err := db.search.Get(ctx, id)
		if err == nil {
			return result, nil
		}

		zlog.Warnf("[db.Get][search] err: %v", err)
	}

	data, err := db.storage.Get(ctx, id)
	if err != nil {
		zlog.Debugf("[db.Get][storage] err: %v", err)
		return empty, errors.F("db.Get storage err: %v", err)
	}

	return data, nil
}

// GetAll retrieves all entities from the search layer (Elasticsearch).
func (db *Database[T, ID]) GetAll(ctx context.Context) ([]T, error) {
	if db.storage == nil {
		zlog.Debug("[db.GetAll] err: storage is nil")
		return nil, errors.F("db.GetAll err: storage is nil")
	}

	if db.search == nil {
		zlog.Info("[db.GetAll] search => pass")
	}

	if db.search != nil {
		ts, err := db.search.GetAll(ctx)
		if err == nil {
			return ts, nil
		}

		zlog.Warnf("[db.GetAll][search] err: %v", err)
	}

	resp, err := db.storage.GetAll(ctx)
	if err != nil {
		zlog.Debugf("[db.GetAll][search] err: %v", err)
		return nil, errors.F("db.GetAll search err: %v", err)
	}

	return resp, nil
}

// Exists checks if an entity exists in the search layer (Elasticsearch).
func (db *Database[T, ID]) Exists(ctx context.Context, id ID) (bool, error) {
	if db.storage == nil {
		zlog.Debug("[db.Exists] storage layer is nil")
		return false, errors.F("db.Exists storage layer is nil")
	}

	if db.search == nil {
		zlog.Info("[db.Exists] search => pass")
	}

	if db.search != nil {
		existed, err := db.search.Exists(ctx, id)
		if err == nil {
			return existed, nil
		}

		zlog.Warnf("[db.Exists] search err: %v", err)
	}

	existed, err := db.storage.Exists(ctx, id)
	if err != nil {
		zlog.Debugf("[db.Exists] storage err: %v", err)
		return false, errors.F("db.Exists storage err: %v", err)
	}

	return existed, nil
}

// Count returns the total number of entities from the search layer
func (db *Database[T, ID]) Count(ctx context.Context) (int64, error) {
	if db.storage == nil {
		zlog.Debug("[db.Count][storage] layer is nil")
		return -1, errors.F("db.Count storage layer is nil")
	}

	if db.search == nil {
		zlog.Info("[db.Exists] search => pass")
	}

	if db.search != nil {
		count, err := db.search.Count(ctx)
		if err == nil {
			return count, nil
		}

		zlog.Warnf("[db.Count][search] err: %v", err)
	}

	count, err := db.storage.Count(ctx)
	if err != nil {
		zlog.Debugf("[db.Count][storage] err: %v", err)
		return -1, errors.F("db.Count storage err: %v", err)
	}

	return count, nil
}
