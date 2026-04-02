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

package chains

import (
	corehttp "github.com/sentinez/core/http"
	"github.com/sentinez/shared/zlog"
)

type Handler interface {
	SetNext(mdw Handler) Handler
	Handle(ctx corehttp.Context) error
}

func New() *BaseHandler {
	return &BaseHandler{}
}

type BaseHandler struct {
	next Handler
}

func (b *BaseHandler) SetNext(handler Handler) Handler {
	if b == nil {
		zlog.Warn("chains: uninitialized base chains")
		return nil
	}

	b.next = handler
	return handler
}

func (b *BaseHandler) GetNext() Handler {
	if b == nil {
		zlog.Warn("chains: uninitialized base chains")
		return nil
	}

	return b.next
}

func (b *BaseHandler) HandleNext(ctx corehttp.Context) error {
	if b == nil {
		zlog.Warn("chains: uninitialized base chains")
		return nil
	}

	if b.next != nil {
		return b.next.Handle(ctx)
	}

	return nil
}
