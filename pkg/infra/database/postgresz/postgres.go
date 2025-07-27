// Copyright 2025 Sentinez Labs.
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

package postgresz

import (
	"fmt"

	"github.com/sentinez/sentinez/pkg/infra/database"
)

func WithIndex(index ...string) database.Option {
	return func(ref *database.Table) {
		ref.Index = append(ref.Index, index...)
	}
}

func Field(field string) string {
	return fmt.Sprintf("data->>'%s'", field)
}

func Primary(field string) string {
	return field
}
