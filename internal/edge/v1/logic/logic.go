// Copyright 2025 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package logic ...
package logic

import (
	"strings"

	httpxf1 "github.com/sentinez/sentinez/pkg/core/net/httpx/f1"
)

func DomainHandler(hostname string,
) func(httpxf1.RequestHandler) httpxf1.RequestHandler {

	return func(next httpxf1.RequestHandler) httpxf1.RequestHandler {
		return func(ctx *httpxf1.Context) error {
			if !isValidSingleLevelSubdomain(string(ctx.Host()), hostname) {
				httpxf1.Forbidden(ctx)
				return nil
			}

			return next(ctx)
		}
	}
}

func isValidSingleLevelSubdomain(subdomain, root string) bool {
	subLabels := strings.Split(subdomain, ".")
	rootLabels := strings.Split(root, ".")

	if len(subLabels) != len(rootLabels)+1 {
		return false
	}

	rootMatch := strings.Join(subLabels[len(subLabels)-len(rootLabels):], ".")
	rootMatch = strings.Split(rootMatch, ":")[0]
	return rootMatch == root
}
