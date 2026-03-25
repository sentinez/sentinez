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
	"net"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
	edgebpf "github.com/sentinez/sentinez/api/bpf/edge"
	"github.com/sentinez/shared/zlog"
)

func setupRlimit() error {
	return rlimit.RemoveMemlock()
}

func loadBPFObjects() (*edgebpf.EdgeObjects, error) {
	var objs edgebpf.EdgeObjects
	err := edgebpf.LoadEdgeObjects(&objs, nil)
	return &objs, err
}

func getInterface(name string) (*net.Interface, error) {
	return net.InterfaceByName(name)
}

func attachXDP(prog *ebpf.Program, ifIndex int) (link.Link, error) {
	return link.AttachXDP(link.XDPOptions{
		Program:   prog,
		Interface: ifIndex,
	})
}

func runCounterLoop(objs *edgebpf.EdgeObjects) {
	tick := time.Tick(time.Second)
	for range tick {
		printCounter(objs)
	}
}

func printCounter(objs *edgebpf.EdgeObjects) {
	var idx uint64
	err := objs.PktCount.Lookup(uint32(0), &idx)
	if err != nil {
		zlog.Fatal("Map lookup:", err)
	}
	zlog.Infof("Received %d", idx)
}
