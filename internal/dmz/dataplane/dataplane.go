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

package dataplane

import (
	"github.com/sentinez/sentinez/internal/dmz/dataplane/ebpf"
	"github.com/sentinez/sentinez/pkg/utils/network"
	"github.com/sentinez/shared/zlog"
)

type DataPlane struct {
}

const VETH0 = "veth0"

func (d *DataPlane) Run(networkInterface string) error {
	ctx := ebpf.NewContext()

	iface, err := network.GetInterface(networkInterface)
	if err != nil {
		zlog.Errorf("stream: getting interface: %v", err)
		return err
	}

	if err := ctx.AttachXDP(iface.Index); err != nil {
		zlog.Errorf("stream: attaching xdp: %v", err)
		return err
	}

	return nil
}

func (d *DataPlane) Close() error {
	return ebpf.CloseContext()
}
