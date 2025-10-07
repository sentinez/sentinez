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
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
)

func Add(perms int32, flag common.Permission) int32 {
	return perms | int32(flag)
}

func Remove(perms int32, flag common.Permission) int32 {
	return perms &^ int32(flag)
}

func Has(perms int32, flag common.Permission) bool {
	return (perms & int32(flag)) != 0
}

func HasLeastOne(perms int32, flags int32) bool {
	return perms&flags != 0
}

func DefaultOwner() int32 {
	return int32(common.Permission_PERMISSION_CREATE_OWN |
		common.Permission_PERMISSION_VIEW_OWN |
		common.Permission_PERMISSION_DELETE_OWN |
		common.Permission_PERMISSION_UPDATE_OWN)
}

func DefaultRoot() int32 {
	return int32(common.Permission_PERMISSION_ROOT)
}

func DefaultViewAny() int32 {
	return int32(common.Permission_PERMISSION_ROOT |
		common.Permission_PERMISSION_VIEW_ANY)
}
