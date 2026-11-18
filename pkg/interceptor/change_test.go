// Copyright 2026 Duc-Hung Ho.
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

package interceptor

import (
	"context"
	"testing"

	"github.com/sentinez/controlplane/pkg/auditchange"
	"github.com/sentinez/controlplane/pkg/headers"
	iampb "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	auditpb "github.com/sentinez/sentinez/api/proto/sentinez/types/audit/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	_userField     = "email_or_username"
	_passwordField = "password"
)

var _changeCases = []struct {
	name       string
	giveBefore proto.Message
	giveAfter  proto.Message
	givePaths  []string
	want       *auditpb.Change
}{
	{
		name:       "nothing recorded",
		givePaths:  []string{_userField},
		giveBefore: nil,
		giveAfter:  nil,
		want:       nil,
	},
	{
		name:       "changed field",
		giveBefore: &iampb.LoginRequest{EmailOrUsername: "bob"},
		giveAfter:  &iampb.LoginRequest{EmailOrUsername: "alice"},
		givePaths:  []string{_userField},
		want: &auditpb.Change{Fields: []*auditpb.FieldChange{{
			Field:  _userField,
			Before: structpb.NewStringValue("bob"),
			After:  structpb.NewStringValue("alice"),
		}}},
	},
	{
		name:       "cleared field",
		giveBefore: &iampb.LoginRequest{EmailOrUsername: "bob"},
		giveAfter:  &iampb.LoginRequest{},
		givePaths:  []string{_userField},
		want: &auditpb.Change{Fields: []*auditpb.FieldChange{{
			Field:  _userField,
			Before: structpb.NewStringValue("bob"),
		}}},
	},
	{
		name:       "written but unchanged",
		giveBefore: &iampb.LoginRequest{EmailOrUsername: "bob"},
		giveAfter:  &iampb.LoginRequest{EmailOrUsername: "bob"},
		givePaths:  []string{_userField},
		want:       nil,
	},
	{
		name:       "changed but not written",
		giveBefore: &iampb.LoginRequest{EmailOrUsername: "bob"},
		giveAfter:  &iampb.LoginRequest{EmailOrUsername: "alice"},
		givePaths:  []string{_passwordField},
		want:       nil,
	},
	{
		name:       "redacted field",
		giveBefore: &iampb.LoginRequest{Password: "hunter22"},
		giveAfter:  &iampb.LoginRequest{Password: "hunter33"},
		givePaths:  []string{_passwordField},
		want: &auditpb.Change{Fields: []*auditpb.FieldChange{{
			Field:  _passwordField,
			Before: structpb.NewStringValue(_redacted),
			After:  structpb.NewStringValue(_redacted),
		}}},
	},
	{
		name:       "unknown path",
		giveBefore: &iampb.LoginRequest{EmailOrUsername: "bob"},
		giveAfter:  &iampb.LoginRequest{EmailOrUsername: "alice"},
		givePaths:  []string{"nope"},
		want:       nil,
	},
	{
		name:       "messages of other types",
		giveBefore: &iampb.LoginRequest{EmailOrUsername: "bob"},
		giveAfter:  &iampb.PasskeyLoginChallengeRequest{},
		givePaths:  []string{_userField},
		want:       nil,
	},
}

func TestChangeOf(t *testing.T) {
	for _, tt := range _changeCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx, change := auditchange.With(context.Background())
			auditchange.Set(ctx, tt.giveBefore, tt.giveAfter, tt.givePaths)

			got := changeOf(change)
			assert.True(t, proto.Equal(tt.want, got), "got %v", got)
		})
	}
}

func TestAuditRecordsChange(t *testing.T) {
	tests := []struct {
		name       string
		giveErr    error
		wantFields int
	}{
		{name: "update succeeded", giveErr: nil, wantFields: 1},
		{
			name:       "update failed",
			giveErr:    status.Error(codes.Internal, "db down"),
			wantFields: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &recorder{}
			ctx := headers.WithAuth(context.Background(), &headers.Auth{})
			info := &grpc.UnaryServerInfo{FullMethod: _loginMethod}

			_, err := Audit(r)(ctx, &iampb.LoginRequest{}, info,
				func(ctx context.Context, _ any) (any, error) {
					auditchange.Set(ctx,
						&iampb.LoginRequest{EmailOrUsername: "bob"},
						&iampb.LoginRequest{EmailOrUsername: "alice"},
						[]string{_userField})
					return nil, tt.giveErr
				})

			require.Equal(t, tt.giveErr, err)
			require.Len(t, r.entries, 1)
			assert.Len(t, r.entries[0].GetChange().GetFields(),
				tt.wantFields)
		})
	}
}
