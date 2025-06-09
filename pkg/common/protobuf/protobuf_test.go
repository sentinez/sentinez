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
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestToTime(t *testing.T) {
	now := time.Now()
	ts := ToTime(now)
	if ts == nil {
		t.Fatalf("ToTime(%v) = nil, want non-nil", now)
	}
	if got := ts.AsTime(); !got.Equal(now) {
		t.Errorf("ToTime(%v).AsTime() = %v, want %v", now, got, now)
	}
}

func TestFromTime(t *testing.T) {
	now := time.Now()
	ts := timestamppb.New(now)
	got := FromTime(ts)
	if !got.Equal(now) {
		t.Errorf("FromTime(%v) = %v, want %v", ts, got, now)
	}
}

func TestCompare(t *testing.T) {
	msg1 := &timestamppb.Timestamp{Seconds: 1, Nanos: 1}
	msg2 := &timestamppb.Timestamp{Seconds: 1, Nanos: 1}
	msg3 := &timestamppb.Timestamp{Seconds: 2, Nanos: 2}

	if diff := cmp.Diff(msg1, msg2, Compare); diff != "" {
		t.Errorf("Compare(msg1, msg2) mismatch (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff(msg1, msg3, Compare); diff == "" {
		t.Errorf("Compare(msg1, msg3) = 0, want non-zero")
	}
}

func TestValidate(t *testing.T) {
	msg := &timestamppb.Timestamp{Seconds: 1, Nanos: 1}
	if err := Validate(msg); err != nil {
		t.Errorf("Validate(%v) = %v, want nil", msg, err)
	}
}
