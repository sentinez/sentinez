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

package logic

import (
	"sync"

	corehttp "github.com/sentinez/sentinez/core/http"
)

var (
	pool = sync.Pool{
		New: func() any {
			return new(Node)
		},
	}
)

type (
	NodeFunc func(ctx corehttp.RequestContext) bool

	NodeType  int
	LogicType int
)

const (
	NodeBase NodeType = iota
	NodeLogic
)

const (
	LogicOr  LogicType = 0
	LogicAnd LogicType = 1
)

func NewLogic(l *Node, op LogicType, r *Node) *Node {
	node := pool.Get().(*Node)

	node.left = l
	node.right = r
	node.op = op
	node.types = NodeLogic

	return node
}

func NewNode(fn NodeFunc) *Node {
	node := pool.Get().(*Node)

	node.types = NodeBase
	node.fn = fn

	return node
}

func Free(node *Node) {
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

type Node struct {
	types NodeType
	left  *Node
	right *Node
	op    LogicType
	fn    NodeFunc
}

func (n *Node) Eval(ctx corehttp.RequestContext) bool {
	if n == nil {
		return false
	}

	switch n.types {
	case NodeBase:
		return n.fn(ctx)
	case NodeLogic:
		switch n.op {
		case LogicAnd:
			l := n.left.Eval(ctx)
			if !l {
				// stop branch AND, left is fasle
				return false
			}
			// zlog.Debug("visit right")
			res := n.right.Eval(ctx)
			return res
		case LogicOr:
			l := n.left.Eval(ctx)
			if l {
				// stop branch OR, left is true
				return true
			}
			// zlog.Debug("visit right")
			res := n.right.Eval(ctx)
			return res
		}

		return false
	default:
		return false
	}
}

func TraversePostfix(node *Node, traveler func(*Node) bool) {
	if node == nil {
		// zlog.Debugf("node is nil")
		return
	}

	if node.left != nil {
		// zlog.Debugf("visit left")
		TraversePostfix(node.left, traveler)
	}

	if node.right != nil {
		// zlog.Debugf("visit right")
		TraversePostfix(node.right, traveler)
	}

	if !traveler(node) {
		return
	}
}

func TraversePrefix(node *Node, traveler func(*Node) bool) {
	if node == nil {
		return
	}

	if !traveler(node) {
		return
	}

	if node.left != nil {
		TraversePrefix(node.left, traveler)
	}

	if node.right != nil {
		TraversePrefix(node.right, traveler)
	}
}
