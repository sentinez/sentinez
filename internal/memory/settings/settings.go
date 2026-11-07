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

package settings

import (
	"sync"

	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/sentinez/pkg/szerror"
	"github.com/sentinez/pkg/szsync"
)

var (
	_once sync.Once
	_inst *Setting
)

// New returns the settings of this edge, by namespace, as synced between
// the edges (memory.MemStore.Sync).
func New() *Setting {
	_once.Do(func() {
		_inst = &Setting{
			setting: szsync.NewMap[string, *edgepb.Setting](),
		}
	})

	return _inst
}

type Setting struct {
	setting *szsync.Map[string, *edgepb.Setting]
}

// Store sets the setting of its namespace, replacing the previous one.
func (s *Setting) Store(st *edgepb.Setting) {
	s.setting.Store(st.GetServer().GetName(), st)
}

// Delete removes the setting of namespace.
func (s *Setting) Delete(namespace string) {
	s.setting.Delete(namespace)
}

func (s *Setting) Load(namespace string) (*edgepb.Setting, error) {
	result, ok := s.setting.Load(namespace)
	if ok {
		return result, nil
	}

	return nil, szerror.ErrNotFound
}

func (s *Setting) Visit(fn func(*edgepb.Setting) bool) {
	s.setting.Range(func(_ string, value *edgepb.Setting) bool {
		return fn(value)
	})
}
