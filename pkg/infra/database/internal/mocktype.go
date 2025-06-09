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

// Package internal not share external module
package internal

import (
	"context"
)

// IAuthors define to mock test
type IAuthors interface {
	Create(ctx context.Context, author Authors) (Authors, error)
	Update(ctx context.Context, id int64, author Authors) (Authors, error)
	Get(ctx context.Context, id int64) (Authors, error)
	GetAll(ctx context.Context) ([]Authors, error)
	Delete(ctx context.Context, id int64) error
	Exists(ctx context.Context, id int64) (bool, error)
	Count(ctx context.Context) (int64, error)
}

// Authors mock test obj
type Authors struct {
	ID int64
}
