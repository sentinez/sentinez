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

package core

import "context"

// RequestContext rule engine context
type RequestContext interface {
	GetContext() context.Context

	// GetHeader request or response header
	GetHeader() map[string]string

	// GetQueries request queries
	GetQueries() []string

	// GetPath request path
	GetPath() string

	// GetBody request or response body
	GetBody() []byte

	// GetIP request ip address
	GetIP() string

	// GetJA4 JA4 fingerprint
	GetJA4() string

	// GetTLS is the connection secure
	GetTLS() bool

	// GetMethod request method
	GetMethod() string

	// GetHost hostname
	GetHost() string
}
