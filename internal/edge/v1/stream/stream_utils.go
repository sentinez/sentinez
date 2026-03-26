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
	"time"

	edgebpf "github.com/sentinez/sentinez/api/bpf/edge"
	"github.com/sentinez/shared/zlog"
)

func RunCounterLoop(objs *edgebpf.EdgeObjects) {
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
