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

package corerule

import (
	"sync"

	corehttp "github.com/sentinez/core/http"
	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/secure/ruleengine/v1"
)

var (
	pool = sync.Pool{
		New: func() any {
			return new(node)
		},
	}
)

type (
	nodeFunc func(ctx corehttp.RequestContext, out *rulepb.MatchedRules) bool

	nodeType  int
	logicType int
)

const (
	nodeBase nodeType = iota
	nodeLogic
)

const (
	logicOr  logicType = 0
	logicAnd logicType = 1
	logicNot logicType = 2
)

func newLogic(l *node, op logicType, r *node) *node {
	node := pool.Get().(*node)

	node.left = l
	node.right = r
	node.op = op
	node.types = nodeLogic

	return node
}

func newNode(fn nodeFunc) *node {
	node := pool.Get().(*node)

	node.types = nodeBase
	node.fn = fn

	return node
}

func free(node *node) {
	if node == nil {
		return
	}

	node.fn = nil
	node.types = 0
	node.left = nil
	node.right = nil
	node.op = 0

	pool.Put(node)
}

type node struct {
	types nodeType
	left  *node
	right *node
	op    logicType
	fn    nodeFunc
}

func (n *node) eval(
	ctx corehttp.RequestContext,
	out *rulepb.MatchedRules,
) bool {
	if n == nil {
		return false
	}

	switch n.types {
	case nodeBase:
		return n.fn(ctx, out)
	case nodeLogic:
		switch n.op {
		case logicAnd:
			l := n.left.eval(ctx, out)
			if !l {
				// stop branch AND, left is fasle
				return false
			}
			// zlog.Debug("visit right")
			res := n.right.eval(ctx, out)
			return res
		case logicOr:
			l := n.left.eval(ctx, out)
			if l {
				// stop branch OR, left is true
				return true
			}
			// zlog.Debug("visit right")
			res := n.right.eval(ctx, out)
			return res
		case logicNot:
			l := n.left.eval(ctx, out)
			return !l
		}

		return false
	default:
		return false
	}
}
