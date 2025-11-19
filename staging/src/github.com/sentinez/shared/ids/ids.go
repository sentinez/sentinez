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

package sids

import (
	"fmt"

	"github.com/google/uuid"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

// NewID generates a new UUID and returns it as a string.
func NewID(prefix string) string {
	id := uuid.New()
	return fmt.Sprintf("%s%s", prefix, id.String())
}

func NewNanoID(prefix string) string {
	id, _ := gonanoid.New()
	return fmt.Sprintf("%s%s", prefix, id)
}
