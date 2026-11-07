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

package schema

import (
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/sztable"
	iampb "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
)

// IAMUsers the schema of the sztable.IAMUsers table.
func IAMUsers() dbx.Schema {
	return dbx.Schema{
		Table: sztable.IAMUsers,
		Columns: []dbx.Column{
			{Name: iampb.User_Id, Type: dbx.Text},
			{Name: iampb.User_Email, Type: dbx.Text},
			{Name: iampb.User_FullName, Type: dbx.Text},
			{Name: iampb.User_ControlPlane, Type: dbx.Int4},
		},
	}
}

// IAMAccounts the schema of the sztable.IAMAccounts table.
func IAMAccounts() dbx.Schema {
	return dbx.Schema{
		Table: sztable.IAMAccounts,
		Columns: []dbx.Column{
			{Name: iampb.Account_Id, Type: dbx.Text},
			{Name: iampb.Account_Email, Type: dbx.Text},
			{Name: iampb.Account_Username, Type: dbx.Text},
			{Name: iampb.Account_PasswordHash, Type: dbx.Text},
			{Name: iampb.Account_Credentials, Type: dbx.TextArr},
			{Name: iampb.Account_UserId, Type: dbx.Text},
			{Name: iampb.Account_Provider, Type: dbx.Int4},
			{Name: iampb.Account_ProviderUserId, Type: dbx.Text},
		},
	}
}
