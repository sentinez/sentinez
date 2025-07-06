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

package coretest

import (
	"testing"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	"github.com/sentinez/sentinez/pkg/common/protobuf"
)

//nolint:funlen
func TestValidateUserModel(t *testing.T) {
	type Test struct {
		data   *iam.Users
		result bool
	}

	tests := []*Test{
		{
			data: &iam.Users{
				Id:          "stnz.users.xxx",
				FullName:    "Name x y z",
				PhoneNumber: "+84875470160",
				Email:       "email@email.com",
			},
			result: true,
		},
		{
			data: &iam.Users{
				Id:          "stnz.usxers.xxx",
				FullName:    "Name x y z",
				PhoneNumber: "+84875470160",
				Email:       "email@email.com",
			},
			result: false,
		},
		{
			data: &iam.Users{
				Id:          "stnz.users.xxx",
				FullName:    "",
				PhoneNumber: "+84875470160",
				Email:       "email@email.com",
			},
			result: false,
		},
		{
			data: &iam.Users{
				Id:          "stnz.users.xxx",
				FullName:    "Name x y z",
				PhoneNumber: "+84875470160",
				Email:       "email@email.com",
			},
			result: false,
		},
		{
			data: &iam.Users{
				Id:          "",
				FullName:    "Name x y z",
				PhoneNumber: "+84875470160z",
				Email:       "email@email.com",
			},
			result: false,
		},
		{
			data: &iam.Users{
				Id:          "",
				FullName:    "Name x y z",
				PhoneNumber: "+84875470160z",
				Email:       "",
			},
			result: false,
		},
	}

	for _, test := range tests {
		if err := protobuf.Validate(test.data); err != nil {
			if test.result {
				t.Error(err)
			}
		}
	}
}
