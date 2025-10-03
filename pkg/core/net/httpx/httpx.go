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

package httpx

// Server is the interface that provides the basic methods for an HTTP server.
type Server interface {
	ListenAndServe(addr string) error
	Shutdown() error
}

// Context is the interface that wraps the basic methods for an HTTP context.
// It provides methods to handle HTTP requests and responses.
type Context interface {
	Release()
	Path() string
	String(statusCode int, body string) error
	JSON(statusCode int, body []byte) error
}
