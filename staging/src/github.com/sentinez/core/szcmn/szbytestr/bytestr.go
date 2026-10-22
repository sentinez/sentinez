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

package szbytestr

import (
	"github.com/sentinez/core"
	"github.com/sentinez/core/szhttp/szhttpconst"
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
	BadRequest          = []byte("Bad request")
	HitCache            = []byte("HIT")
	MethodNotAllowed    = []byte(szhttpconst.ValueMethodNotAllowed)
)

var (
	HeaderContentType      = []byte(szhttpconst.HeaderContentType)
	HeaderXForwardedFor    = []byte(szhttpconst.HeaderXForwardedFor)
	HeaderXForwardedPrefix = []byte(szhttpconst.HeaderXForwardedPrefix)
	HeaderServer           = []byte(szhttpconst.HeaderServer)
	HeaderTransferEncoding = []byte(szhttpconst.HeaderTransferEncoding)
	HeaderXRealIP          = []byte(szhttpconst.HeaderXRealIP)
	HeaderUserAgent        = []byte(szhttpconst.HeaderUserAgent)
	HeaderCacheControl     = []byte(szhttpconst.HeaderCacheControl)
	HeaderUpgrade          = []byte(szhttpconst.HeaderUpgrade)
	HeaderXCache           = []byte(szhttpconst.HeaderXCache)
	HeaderAcceptEncoding   = []byte(szhttpconst.HeaderAcceptEncoding)
	HeaderAltSvc           = []byte(szhttpconst.HeaderAltSvc)
	HeaderSetCookie        = []byte(szhttpconst.HeaderSetCookie)
	HeaderETag             = []byte(szhttpconst.HeaderETag)
	HeaderIfNoneMatch      = []byte(szhttpconst.HeaderIfNoneMatch)
	HeaderContentLength    = []byte(szhttpconst.HeaderContentLength)
	HeaderAllow            = []byte(szhttpconst.HeaderAllow)
	HeaderXContentTypeOpts = []byte(szhttpconst.HeaderXContentTypeOpts)

	ValueAppJSON   = []byte(szhttpconst.ValueAppJSON)
	ValueTextPlain = []byte(szhttpconst.ValueTextPlain)
	ValueTextHTML  = []byte(szhttpconst.ValueTextHTML)
	ValueNoSniff   = []byte(szhttpconst.ValueNoSniff)

	ValueAllowGetHead = []byte(szhttpconst.ValueAllowGetHead)
)
