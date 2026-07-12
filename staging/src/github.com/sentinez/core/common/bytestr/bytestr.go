// Copyright 2026 Duc-Hung Ho.
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

package bytestr

import (
	"github.com/sentinez/core"
	httpconst "github.com/sentinez/core/http/const"
)

var (
	DefaultServerName = []byte(core.Name)
	DefaultUserAgent  = []byte(core.BaseName)
)

var (
	AccessDenied        = []byte("Access denied")
	InternalServerError = []byte("Internal server error")
	NotFound            = []byte("Not found")
	TooManyRequests     = []byte("Too many requests")
)

var (
	HeaderContentType      = []byte(httpconst.HeaderContentType)
	HeaderXForwardedFor    = []byte(httpconst.HeaderXForwardedFor)
	HeaderXForwardedPrefix = []byte(httpconst.HeaderXForwardedPrefix)
	HeaderServer           = []byte(httpconst.HeaderServer)
	HeaderTransferEncoding = []byte(httpconst.HeaderTransferEncoding)
	HeaderXRealIP          = []byte(httpconst.HeaderXRealIP)
	HeaderUserAgent        = []byte(httpconst.HeaderUserAgent)
	HeaderCacheControl     = []byte(httpconst.HeaderCacheControl)
	HeaderUpgrade          = []byte(httpconst.HeaderUpgrade)

	ValueAppJSON   = []byte(httpconst.ValueAppJSON)
	ValueTextPlain = []byte(httpconst.ValueTextPlain)
	ValueTextHTML  = []byte(httpconst.ValueTextHTML)
)
