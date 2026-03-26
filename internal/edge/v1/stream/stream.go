// Copyright 2025 Duc-Hung Ho.
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

package stream

import (
	streamctx "github.com/sentinez/sentinez/internal/edge/v1/stream/context"
	"github.com/sentinez/sentinez/pkg/network"
	"github.com/sentinez/shared/zlog"
)

const networkInterface = "veth0"

func Init() error {
	ctx := streamctx.New()

	iface, err := network.GetInterface(networkInterface)
	if err != nil {
		zlog.Errorf("getting interface: %v", err)
		return err
	}

	if err := ctx.AttachXDP(iface.Index); err != nil {
		zlog.Errorf("attaching XDP: %v", err)
		return err
	}

	zlog.Infof("counting incoming packets on %s..", iface.Name)

	return nil
}

func Close() error {
	return streamctx.Get().Close()
}
