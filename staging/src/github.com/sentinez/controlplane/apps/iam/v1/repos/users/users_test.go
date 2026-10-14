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

package usersrepo

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	iampb "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListConditions(t *testing.T) {
	tests := []struct {
		name      string
		give      *iampb.ListUsersRequest
		wantQuery string
		wantArgs  []any
	}{
		{
			name:      "no filter",
			give:      &iampb.ListUsersRequest{},
			wantQuery: "SELECT * FROM users",
		},
		{
			// several ids must match any of them, not all at once
			name:      "ids",
			give:      &iampb.ListUsersRequest{Ids: []string{"a", "b"}},
			wantQuery: "SELECT * FROM users WHERE " + iampb.User_Id + " IN (?,?)",
			wantArgs:  []any{"a", "b"},
		},
		{
			name: "ids and emails",
			give: &iampb.ListUsersRequest{
				Ids:    []string{"a"},
				Emails: []string{"x@example.com", "y@example.com"},
			},
			wantQuery: "SELECT * FROM users WHERE " + iampb.User_Id +
				" IN (?) AND " +
				iampb.User_Email + " IN (?,?)",
			wantArgs: []any{"a", "x@example.com", "y@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := sq.Select("*").From("users")
			for _, cond := range listConditions(tt.give) {
				builder = builder.Where(cond)
			}

			query, args, err := builder.ToSql()

			require.NoError(t, err)
			assert.Equal(t, tt.wantQuery, query)
			assert.Equal(t, tt.wantArgs, args)
		})
	}
}
