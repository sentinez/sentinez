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

import "github.com/sentinez/shared/zlog"

const networkInterface = "veth0"

func Init() error {
	if err := setupRlimit(); err != nil {
		zlog.Errorf("remove mem lock err=%v", err)
		return err
	}

	objs, err := loadBPFObjects()
	if err != nil {
		zlog.Errorf("loading eBPF objects: %v", err)
		return err
	}
	defer func() { _ = objs.Close() }()

	iface, err := getInterface(networkInterface)
	if err != nil {
		zlog.Errorf("Getting interface: %v", err)
		return err
	}

	link, err := attachXDP(objs.EdgeMain, iface.Index)
	if err != nil {
		zlog.Errorf("Attaching XDP: %v", err)
		return err
	}
	defer func() { _ = link.Close() }()

	zlog.Infof("Counting incoming packets on %s..", iface.Name)

	runCounterLoop(objs)

	return nil
}
