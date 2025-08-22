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

// Package protobuf provides utilities for working with protobuf messages.
package protobuf

import (
	"time"

	"buf.build/go/protovalidate"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToTime creates pb Timestamp from time.Time.
func ToTime(from time.Time) *timestamppb.Timestamp {
	return timestamppb.New(from)
}

// FromTime creates time.Time from pb Timestamp.
func FromTime(from *timestamppb.Timestamp) time.Time {
	return from.AsTime()
}

// ToDuration creates pb Duration from time.Duration.
func ToDuration(from time.Duration) *durationpb.Duration {
	return durationpb.New(from)
}

// FromDuration creates time.Duration from pb Duration.
func FromDuration(from *durationpb.Duration) time.Duration {
	return from.AsDuration()
}

// Compare compares two proto messages.
var Compare = cmp.FilterValues(
	func(x, y any) bool {
		_, xok := x.(proto.Message)
		_, yok := y.(proto.Message)
		return xok && yok
	},
	cmp.Comparer(func(x, y any) bool {
		vx, ok := x.(proto.Message)
		if !ok {
			return false
		}
		vy, ok := y.(proto.Message)
		if !ok {
			return false
		}
		return proto.Equal(vx, vy)
	}),
)

// Validate validates a proto message.
func Validate(msg proto.Message) error {
	v, err := protovalidate.New()
	if err != nil {
		return err
	}

	return v.Validate(msg)
}
