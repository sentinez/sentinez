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

package corerule

import (
	"regexp"
	"sync"
)

var (
	regexCache sync.Map
)

func matchRegex(pattern, value string) bool {
	if pattern == "" {
		return false
	}

	// Try to get from cache
	if v, ok := regexCache.Load(pattern); ok {
		if re, ok := v.(*regexp.Regexp); ok {
			return re.MatchString(value)
		}
	}

	// Compile and cache
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	regexCache.Store(pattern, re)
	return re.MatchString(value)
}
