// Copyright 2026 Sentinéz Labs.
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

// Package channel names the channels of the pub/sub of the cluster.
package channel

// Channel is the name of a channel of the pub/sub of the cluster.
type Channel string

// CentralData the channel centraldata publishes the changes of its
// tables on, and the edges subscribe to.
const CentralData Channel = "sentinez.centraldata.changes"

// _syncPrefix prefixes the channel of a sync map.
const _syncPrefix = "sentinez.sync."

// Sync the channel the sync map name replicates its entries on.
func Sync(name string) Channel {
	return Channel(_syncPrefix + name)
}

func (ch Channel) String() string {
	return string(ch)
}
