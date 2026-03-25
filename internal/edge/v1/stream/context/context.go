// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package streamctx

import (
	"fmt"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	edgebpf "github.com/sentinez/sentinez/api/bpf/edge"
	"github.com/sentinez/shared/zlog"
)

var (
	once sync.Once
	inst *Context
)

func Get() *Context {
	return New()
}

func New() *Context {
	once.Do(func() {
		var objs edgebpf.EdgeObjects
		if err := edgebpf.LoadEdgeObjects(&objs, nil); err != nil {
			zlog.Errorf("stream: load bpf object err: %v", err)
		}

		inst = &Context{
			obj: &objs,
		}
	})

	return inst
}

type Context struct {
	obj  *edgebpf.EdgeObjects
	link link.Link
}

func (ctx *Context) Program() *ebpf.Program {
	if ctx == nil {
		return nil
	}

	return ctx.obj.EdgeMain
}

func (ctx *Context) AttachXDP(ifIndex int) error {
	if ctx == nil || ctx.obj == nil {
		return fmt.Errorf("stream context unavailable: context is nil")
	}

	l, err := link.AttachXDP(link.XDPOptions{
		Program:   ctx.Program(),
		Interface: ifIndex,
	})
	if err != nil {
		return err
	}

	ctx.link = l

	return nil
}

func (ctx *Context) Close() error {
	if ctx == nil {
		return nil
	}

	if err := ctx.obj.Close(); err != nil {
		return err
	}

	if err := ctx.link.Close(); err != nil {
		return err
	}

	return nil
}
