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

package static

import (
	"path/filepath"
	"strings"

	httpxhz "github.com/sentinez/sentinez/pkg/core/net/httpx/hz"
)

var staticExts = map[string]struct{}{
	".css": {}, ".js": {}, ".png": {}, ".jpg": {}, ".jpeg": {},
	".gif": {}, ".ico": {}, ".svg": {}, ".woff": {}, ".woff2": {},
	".ttf": {}, ".eot": {}, ".map": {},
}

func IsStaticAsset(pathStr string) bool {
	ext := strings.ToLower(filepath.Ext(pathStr))
	_, ok := staticExts[ext]
	return ok
}

func HeaderCacheControlHandler(
	next httpxhz.RequestHandler) httpxhz.RequestHandler {

	return func(ctx *httpxhz.Context) error {
		err := next(ctx)

		if IsStaticAsset(string(ctx.Path())) {
			ctx.Response.Header.Set(
				"Cache-Control",
				"public, max-age=3600, immutable",
			)
		}

		return err
	}
}
