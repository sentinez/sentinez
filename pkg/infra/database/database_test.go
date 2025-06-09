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

package database

import (
	"context"
	"testing"

	"github.com/sentinez/sentinez/pkg/infra/database/internal"
	"github.com/sentinez/sentinez/pkg/infra/database/internal/mocks"

	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	storageLayer := mocks.NewIAuthors(t)
	storageLayer.On("Create", mock.Anything, mock.Anything).
		Return(internal.Authors{ID: 123}, nil)

	db := New[internal.Authors, int64](nil, storageLayer)
	resp, err := db.Create(context.Background(), internal.Authors{
		ID: 123,
	})
	if err != nil {
		t.Fatal(err)
	}

	if resp.ID != 123 {
		t.Errorf("got id %d, want 123", resp.ID)
	}
}

func TestGet(t *testing.T) {
	storageLayer := mocks.NewIAuthors(t)
	// Set up expectation for Get
	storageLayer.
		On("Get", mock.Anything, int64(123)).
		Return(internal.Authors{ID: 123}, nil)

	db := New[internal.Authors, int64](nil, storageLayer)
	resp, err := db.Get(context.Background(), 123)
	if err != nil {
		t.Fatal(err)
	}
	if resp.ID != 123 {
		t.Errorf("got id %d, want 123", resp.ID)
	}
	storageLayer.AssertExpectations(t)
}

func TestGetAll(t *testing.T) {
	storageLayer := mocks.NewIAuthors(t)
	expected := []internal.Authors{
		{ID: 123},
		{ID: 456},
	}
	// Set up expectation for GetAll
	storageLayer.
		On("GetAll", mock.Anything).
		Return(expected, nil)

	db := New[internal.Authors, int64](nil, storageLayer)
	resp, err := db.GetAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) != len(expected) {
		t.Errorf("got %d items, want %d", len(resp), len(expected))
	}
	// Additional checks on specific values can be added if necessary
	storageLayer.AssertExpectations(t)
}

func TestUpdate(t *testing.T) {
	storageLayer := mocks.NewIAuthors(t)
	// Assume updating Authors with ID = 123 with new information
	updatedAuthor := internal.Authors{ID: 123}
	// Set up expectation for
	// storageLayer.
	// 	On("Get", mock.Anything, int64(123)).
	// 	Return(internal.Authors{ID: 123}, nil)
	storageLayer.
		On("Update", mock.Anything, int64(123), updatedAuthor).
		Return(updatedAuthor, nil)

	db := New[internal.Authors, int64](nil, storageLayer)
	resp, err := db.Update(context.Background(), 123, updatedAuthor)
	if err != nil {
		t.Fatal(err)
	}
	if resp.ID != 123 {
		t.Errorf("got id %d, want 123", resp.ID)
	}
	storageLayer.AssertExpectations(t)
}

func TestDelete(t *testing.T) {
	storageLayer := mocks.NewIAuthors(t)
	// Set up expectation for Delete
	// storageLayer.
	// 	On("Get", mock.Anything, int64(123)).
	// 	Return(internal.Authors{ID: 123}, nil)
	storageLayer.
		On("Delete", mock.Anything, int64(123)).
		Return(nil)

	db := New[internal.Authors, int64](nil, storageLayer)
	err := db.Delete(context.Background(), 123)
	if err != nil {
		t.Fatal(err)
	}
	storageLayer.AssertExpectations(t)
}

func TestExists(t *testing.T) {
	storageLayer := mocks.NewIAuthors(t)
	// Assume the author with id 123 exists
	storageLayer.
		On("Exists", mock.Anything, int64(123)).
		Return(true, nil)

	db := New[internal.Authors, int64](nil, storageLayer)
	exists, err := db.Exists(context.Background(), 123)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Errorf("expected author to exist")
	}
	storageLayer.AssertExpectations(t)
}

func TestCount(t *testing.T) {
	storageLayer := mocks.NewIAuthors(t)
	// Assume there are 2 records
	storageLayer.
		On("Count", mock.Anything).
		Return(int64(2), nil)

	db := New[internal.Authors, int64](nil, storageLayer)
	count, err := db.Count(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("got count %d, want 2", count)
	}
	storageLayer.AssertExpectations(t)
}
