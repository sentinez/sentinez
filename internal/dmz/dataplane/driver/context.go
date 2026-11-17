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

// Package driver loads the sentinez eBPF objects and drives them: attaching
// the XDP program, and reading and writing its maps.
package driver

import (
	"errors"
	"fmt"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
	sentinezbpf "github.com/sentinez/bpf/sentinez"
)

// Context the loaded eBPF objects, and the XDP link once attached. Close
// releases both.
type Context struct {
	obj  *sentinezbpf.SenzObjects
	link link.Link
}

// NewContext loads the eBPF objects into the kernel.
func NewContext() (*Context, error) {
	if err := rlimit.RemoveMemlock(); err != nil {
		return nil, fmt.Errorf("driver: remove memlock: %w", err)
	}

	var objs sentinezbpf.SenzObjects
	if err := sentinezbpf.LoadSenzObjects(&objs, nil); err != nil {
		return nil, fmt.Errorf("driver: load objects: %w", err)
	}

	return &Context{obj: &objs}, nil
}

// AttachXDP attaches the XDP program to the interface of index ifIndex.
func (ctx *Context) AttachXDP(ifIndex int) error {
	l, err := link.AttachXDP(link.XDPOptions{
		Program:   ctx.obj.SenzMain,
		Interface: ifIndex,
	})
	if err != nil {
		return fmt.Errorf("driver: attach xdp: %w", err)
	}

	ctx.link = l

	return nil
}

// Close detaches the XDP program, then unloads the objects.
func (ctx *Context) Close() error {
	var errs []error
	if ctx.link != nil {
		errs = append(errs, ctx.link.Close())
	}

	errs = append(errs, ctx.obj.Close())

	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("driver: close: %w", err)
	}

	return nil
}
