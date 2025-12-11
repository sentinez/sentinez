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

	corehttp "github.com/sentinez/core/http"
	"github.com/sentinez/sentinez/pkg/dmz/chains"
)

var _ chains.Handler = (*Static)(nil)

func NewStatic() chains.Handler {
	return &Static{
		BaseHandler: chains.New(),
		staticExits: map[string]struct{}{
			".css": {}, ".js": {}, ".png": {}, ".jpg": {}, ".jpeg": {},
			".gif": {}, ".ico": {}, ".svg": {}, ".woff": {}, ".woff2": {},
			".ttf": {}, ".eot": {}, ".map": {},
		},
	}
}

type Static struct {
	*chains.BaseHandler
	staticExits map[string]struct{}
}

func (s *Static) Handle(ctx corehttp.Context) error {
	// zlog.Debug("[edge] >>> visit static")

	err := s.HandleNext(ctx)

	if s.isStaticAsset(ctx.Path()) {
		ctx.SetResponseHeader(
			corehttp.HeaderCacheControl, "public, max-age=3600, immutable",
		)
	}

	return err
}

func (s *Static) isStaticAsset(pathStr string) bool {
	ext := strings.ToLower(filepath.Ext(pathStr))
	_, ok := s.staticExits[ext]
	return ok
}
