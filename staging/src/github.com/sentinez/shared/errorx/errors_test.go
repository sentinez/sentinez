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

// Package stderr provide all type of error in sentinez universal
package errorx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _tests = []struct {
	name     string
	err      error
	code     codes.Code
	expected string
}{
	{
		name:     "ErrUnspecified",
		err:      StatusUnspecified,
		code:     codes.Unknown,
		expected: _unspecified,
	},
	{
		name:     "ErrInternalError",
		err:      StatusInternalError,
		code:     codes.Internal,
		expected: _internalError,
	},
	{
		name:     "ErrNotFound",
		err:      StatusNotFound,
		code:     codes.NotFound,
		expected: _notFound,
	},
	{
		name:     "ErrUnauthorized",
		err:      StatusUnauthorized,
		code:     codes.Unauthenticated,
		expected: _unauthorized,
	},
	{
		name:     "ErrForbidden",
		err:      StatusForbidden,
		code:     codes.PermissionDenied,
		expected: _forbidden,
	},
	{
		name:     "ErrInvalidData",
		err:      StatusInvalidData,
		code:     codes.InvalidArgument,
		expected: _invalidData,
	},
}

func TestErrors(t *testing.T) {

	for _, tt := range _tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.code, status.Code(tt.err))
			assert.Equal(t, tt.expected, status.Convert(tt.err).Message())
		})
	}
}
