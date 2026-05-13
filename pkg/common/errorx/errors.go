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

package errorx

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/sentinez/core"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	forbidden = fmt.Sprintf("%s-%d: %s", core.Code,
		typepb.Errors_ERRORS_FORBIDDEN,
		typepb.Errors_ERRORS_FORBIDDEN.String())

	unspecified = fmt.Sprintf("%s-%d: %s", core.Code,
		typepb.Errors_ERRORS_UNSPECIFIED,
		typepb.Errors_ERRORS_UNSPECIFIED.String())

	internalError = fmt.Sprintf("%s-%d: %s", core.Code,
		typepb.Errors_ERRORS_INTERNAL_ERROR,
		typepb.Errors_ERRORS_INTERNAL_ERROR.String())

	notFound = fmt.Sprintf("%s-%d: %s", core.Code,
		typepb.Errors_ERRORS_NOT_FOUND,
		typepb.Errors_ERRORS_NOT_FOUND.String())

	unauthorized = fmt.Sprintf("%s-%d: %s", core.Code,
		typepb.Errors_ERRORS_UNAUTHORIZED,
		typepb.Errors_ERRORS_UNAUTHORIZED.String())

	invalidData = fmt.Sprintf("%s-%d: %s", core.Code,
		typepb.Errors_ERRORS_INVALID_DATA,
		typepb.Errors_ERRORS_INVALID_DATA.String())

	unimplemented = fmt.Sprintf("%s-%d: %s", core.Code,
		typepb.Errors_ERRORS_UNIMPLEMENTED,
		typepb.Errors_ERRORS_UNIMPLEMENTED.String())
)

var (
	// StatusUnspecified is a generic error
	StatusUnspecified = status.Error(codes.Unknown, unspecified)

	// StatusInternalError is an internal error
	StatusInternalError = status.Error(codes.Internal, internalError)

	// StatusNotFound is a not found error
	StatusNotFound = status.Error(codes.NotFound, notFound)

	// StatusUnauthorized is an unauthorized error
	StatusUnauthorized = status.Error(codes.Unauthenticated, unauthorized)

	// StatusForbidden is a forbidden error
	StatusForbidden = status.Error(codes.PermissionDenied, forbidden)

	// StatusInvalidData is an invalid data error
	StatusInvalidData = status.Error(codes.InvalidArgument, invalidData)

	// StatusUnimplemented is a not implemented error
	StatusUnimplemented = status.Error(codes.Unimplemented, unimplemented)
)

var (
	// ErrServerClosed is a server closed error
	ErrServerClosed = http.ErrServerClosed

	// ErrUnimplemented is a generic error
	ErrUnimplemented = errors.New(unimplemented)

	// ErrInvalidData is an invalid data error
	ErrInvalidData = errors.New(invalidData)

	// ErrNotFound is a not found error
	ErrNotFound = errors.New(notFound)
)

// F wrapped error with format template
func F(template string, args ...any) error {
	return fmt.Errorf(template, args...)
}

func StatusUnspecifiedF(format string, args ...any) error {
	return status.Error(codes.Unknown, fmt.Sprintf(format, args...))
}

func StatusInternalErrorF(format string, args ...any) error {
	return status.Error(codes.Internal, fmt.Sprintf(format, args...))
}

func StatusNotFoundF(format string, args ...any) error {
	return status.Error(codes.NotFound, fmt.Sprintf(format, args...))
}

func StatusUnauthorizedF(format string, args ...any) error {
	return status.Error(codes.Unauthenticated, fmt.Sprintf(format, args...))
}

func StatusForbiddenF(format string, args ...any) error {
	return status.Error(codes.PermissionDenied, fmt.Sprintf(format, args...))
}

func StatusInvalidDataF(format string, args ...any) error {
	return status.Error(codes.InvalidArgument, fmt.Sprintf(format, args...))
}

func StatusUnimplementedF(format string, args ...any) error {
	return status.Error(codes.Unimplemented, fmt.Sprintf(format, args...))
}

func StatusAlreadyExistsF(format string, args ...any) error {
	return status.Error(codes.AlreadyExists, fmt.Sprintf(format, args...))
}

func StatusInvalidArgumentF(format string, args ...any) error {
	return status.Error(codes.InvalidArgument, fmt.Sprintf(format, args...))
}

// Is checks if the error is a specific error
func Is(err error, target error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, target) {
		return true
	}
	if errors.Is(err, ErrUnimplemented) {
		return true
	}
	return false
}

func NotRowsNotFound(err error) bool {
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return true
	}

	return false
}
