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

package api

import (
	"embed"
	"io/fs"
)

var (

	//go:embed docs/swagger
	swagger embed.FS

	//go:embed docs/v1
	apiDocsV1 embed.FS
)

func DocsV1() fs.FS {
	f, _ := fs.Sub(apiDocsV1, "docs/v1")
	return f
}

func Swagger() fs.FS {
	f, _ := fs.Sub(swagger, "docs/swagger")
	return f
}
