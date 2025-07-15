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

const alterQuery = `
	ALTER TABLE %s
	ADD CONSTRAINT %s FOREIGN KEY (%s)
	REFERENCES %s(%s)
	ON DELETE CASCADE
	ON UPDATE CASCADE;
`

const checkConstraint = `
	SELECT 1
	FROM pg_constraint
	WHERE conname = $1;
`
