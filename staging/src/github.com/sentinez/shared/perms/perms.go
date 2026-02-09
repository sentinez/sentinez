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

package perms

import (
	"errors"

	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
)

type Claim int32

func New(p int32) Claim {
	return Claim(p)
}

func (c Claim) Int32() int32 {
	return int32(c)
}

func (c Claim) Add(p typepb.Permission) Claim {
	return c | Claim(p)
}

func (c Claim) Remove(p typepb.Permission) Claim {
	return c &^ Claim(p)
}

func (c Claim) Has(p typepb.Permission) bool {
	return (c & Claim(p)) != 0
}

func (c Claim) HasAny(flags Claim) bool {
	return (c & flags) != 0
}

func (c Claim) HasAll(flags Claim) bool {
	return (c & flags) == flags
}

// Check verifies if the claim and role satisfy the requirements.
func (c Claim) Check(requires []*typepb.XRequire, role typepb.Role) error {
	return Allow(requires, role, c)
}

// Allow checks if the given role and permissions satisfy any of the requirements.
// The logic follows:
//   - If requires is empty, access is allowed.
//   - It iterates through the list of requirements (OR logic).
//   - For each requirement:
//   - If Role is specified, the user's role must match (AND logic).
//   - If Permission is specified, the user must have that permission (AND logic).
//   - If any requirement is fully satisfied, returns nil.
//   - If no requirement is satisfied, returns an error.
func Allow(requires []*typepb.XRequire, role typepb.Role, c Claim) error {
	if len(requires) == 0 {
		return nil
	}

	for _, req := range requires {
		// Check Role constraint
		roleMatch := true
		if req.Role != typepb.Role_ROLE_UNSPECIFIED {
			if role != req.Role {
				roleMatch = false
			}
		}

		// Check Permission constraint
		permMatch := true
		if req.Permission != typepb.Permission_PERMISSION_UNSPECIFIED {
			if !c.Has(req.Permission) {
				permMatch = false
			}
		}

		if roleMatch && permMatch {
			return nil
		}
	}

	return errors.New("permission denied")
}
