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

package corechains

import (
	corehttp "github.com/sentinez/core/http"
	"github.com/sentinez/shared/zlog"
)

type ChainNode interface {
	SetNext(mdw ChainNode) ChainNode
	Handle(ctx corehttp.Context) error
}

func NewNode() *Node {
	return &Node{}
}

type Node struct {
	next ChainNode
}

func (n *Node) SetNext(node ChainNode) ChainNode {
	if n == nil {
		zlog.Warn("node: uninitialized base node")
		return nil
	}

	n.next = node
	return node
}

func (n *Node) GetNext() ChainNode {
	if n == nil {
		zlog.Warn("node: uninitialized base node")
		return nil
	}

	return n.next
}

func (n *Node) HandleNext(ctx corehttp.Context) error {
	if n == nil {
		zlog.Warn("node: uninitialized base node")
		return nil
	}

	if n.next != nil {
		return n.next.Handle(ctx)
	}

	return nil
}
