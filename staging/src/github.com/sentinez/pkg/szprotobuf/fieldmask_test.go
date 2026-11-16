// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package szprotobuf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type applyMaskTest struct {
	name      string
	giveSrc   *durationpb.Duration
	givePaths []string
	wantPaths []string
	wantDst   *durationpb.Duration
}

// applyMaskTests the masks of dst {Seconds: 1, Nanos: 2}, both fields
// updatable.
func applyMaskTests() []applyMaskTest {
	return []applyMaskTest{
		{
			name:      "implied mask: the fields set in src",
			giveSrc:   &durationpb.Duration{Seconds: 9},
			wantPaths: []string{"seconds"},
			wantDst:   &durationpb.Duration{Seconds: 9, Nanos: 2},
		},
		{
			name:      "named field unset in src is cleared",
			giveSrc:   &durationpb.Duration{},
			givePaths: []string{"nanos"},
			wantPaths: []string{"nanos"},
			wantDst:   &durationpb.Duration{Seconds: 1},
		},
		{
			name:      "every updatable field",
			giveSrc:   &durationpb.Duration{Seconds: 9},
			givePaths: []string{MaskAll},
			wantPaths: []string{"seconds", "nanos"},
			wantDst:   &durationpb.Duration{Seconds: 9},
		},
		{
			name:      "paths sorted, once each",
			giveSrc:   &durationpb.Duration{Seconds: 9, Nanos: 8},
			givePaths: []string{"seconds", "nanos", "seconds"},
			wantPaths: []string{"nanos", "seconds"},
			wantDst:   &durationpb.Duration{Seconds: 9, Nanos: 8},
		},
	}
}

func TestApplyMask(t *testing.T) {
	for _, tt := range applyMaskTests() {
		t.Run(tt.name, func(t *testing.T) {
			dst := &durationpb.Duration{Seconds: 1, Nanos: 2}
			mask := &fieldmaskpb.FieldMask{Paths: tt.givePaths}

			paths, err := ApplyMask(dst, tt.giveSrc, mask,
				"seconds", "nanos")

			require.NoError(t, err)
			assert.Equal(t, tt.wantPaths, paths)
			assert.Equal(t, tt.wantDst.GetSeconds(), dst.GetSeconds())
			assert.Equal(t, tt.wantDst.GetNanos(), dst.GetNanos())
		})
	}
}

func TestApplyMaskNotUpdatable(t *testing.T) {
	dst := &durationpb.Duration{Seconds: 1, Nanos: 2}
	src := &durationpb.Duration{Seconds: 9, Nanos: 8}
	mask := &fieldmaskpb.FieldMask{Paths: []string{"seconds", "nanos"}}

	_, err := ApplyMask(dst, src, mask, "seconds")

	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, int64(1), dst.GetSeconds(), "dst changed")
}

func TestApplyMaskImpliedKeepsUpdatable(t *testing.T) {
	dst := &durationpb.Duration{Seconds: 1, Nanos: 2}
	src := &durationpb.Duration{Seconds: 9, Nanos: 8}

	paths, err := ApplyMask(dst, src, nil, "nanos")

	require.NoError(t, err)
	assert.Equal(t, []string{"nanos"}, paths)
	assert.Equal(t, int64(1), dst.GetSeconds())
	assert.Equal(t, int32(8), dst.GetNanos())
}
