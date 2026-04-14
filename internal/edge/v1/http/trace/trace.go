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

package trace

import (
	corechains "github.com/sentinez/core/chains"
	corehttp "github.com/sentinez/core/http"
	"github.com/sentinez/sentinez"
	"github.com/sentinez/shared/ids"
	"github.com/sentinez/shared/zlog"
)

func NewTracer(_ zlog.Level) corechains.ChainNode {
	return &Trace{
		Node: corechains.NewNode(),
	}
}

type Trace struct {
	*corechains.Node
}

func (t *Trace) Handle(ctx corehttp.Context) error {
	requestId := ids.NewXID(sentinez.PrefixXRequestIdBytes)
	ctx.SetRequestId(requestId)

	return t.HandleNext(ctx)
}
