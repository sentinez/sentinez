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

import typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"

func DefaultOwner() Claim {
	return Claim(typepb.Permission_PERMISSION_CREATE_OWN |
		typepb.Permission_PERMISSION_VIEW_OWN |
		typepb.Permission_PERMISSION_DELETE_OWN |
		typepb.Permission_PERMISSION_UPDATE_OWN)
}

func DefaultRoot() Claim {
	return Claim(typepb.Permission_PERMISSION_ROOT)
}

func DefaultViewAny() Claim {
	return Claim(typepb.Permission_PERMISSION_ROOT |
		typepb.Permission_PERMISSION_VIEW_ANY)
}
