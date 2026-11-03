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
	analyticpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	"github.com/sentinez/shared/szstorage/dbx"
	"github.com/sentinez/shared/sztable"
)

// AnalyticActivities the schema of the sztable.AnalyticActivities table.
func AnalyticActivities() dbx.Schema {
	return dbx.Schema{
		Table: sztable.AnalyticActivities,
		Columns: []dbx.Column{
			{Name: analyticpb.Activity_Id, Type: dbx.Text},
			{Name: analyticpb.Activity_OwnerResourceId, Type: dbx.Text},
			{Name: analyticpb.Activity_OwnerTenantId, Type: dbx.Text},
			{Name: analyticpb.Activity_UniqueVisitor, Type: dbx.IntArr},
		},
	}
}
