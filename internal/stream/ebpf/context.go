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

package streamebpf

import (
	"fmt"
	"sync"

	"github.com/cilium/ebpf/link"
	sentinezbpf "github.com/sentinez/sentinez/pkg/dmz/bpf/sentinez"
	"github.com/sentinez/shared/zlog"
)

var (
	once    sync.Once
	context *Context
)

func NewContext() *Context {
	once.Do(func() {
		if err := setupRlimit(); err != nil {
			zlog.Errorf("stream: remove mem lock err=%v", err)
			return
		}

		var objs sentinezbpf.SenzObjects
		if err := sentinezbpf.LoadSenzObjects(&objs, nil); err != nil {
			zlog.Errorf("stream: load bpf object err: %v", err)
			return
		}

		context = &Context{
			obj: &objs,
		}
	})

	return context
}

func CloseContext() error {
	if context == nil {
		return nil
	}

	if err := context.Close(); err != nil {
		return fmt.Errorf("stream: close context err: %v", err)
	}

	return nil
}

type Context struct {
	obj  *sentinezbpf.SenzObjects
	link link.Link
}

func (ctx *Context) AttachXDP(ifIndex int) error {
	if ctx == nil || ctx.obj == nil {
		return fmt.Errorf("stream context unavailable: context is nil")
	}

	l, err := link.AttachXDP(link.XDPOptions{
		Program:   ctx.obj.SenzMain,
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

	if ctx.obj != nil {
		if err := ctx.obj.Close(); err != nil {
			return err
		}
	}

	if ctx.link != nil {
		if err := ctx.link.Close(); err != nil {
			return err
		}
	}

	return nil
}
