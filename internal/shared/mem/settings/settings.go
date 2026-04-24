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

	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	"github.com/sentinez/sentinez/pkg/common/errorx"
	ssync "github.com/sentinez/shared/sync"
)

var (
	once sync.Once
	inst *Setting
	mu   sync.Mutex
)

func New() *Setting {
	once.Do(func() {
		inst = &Setting{
			setting: ssync.NewMap[string, *edgepb.Setting](),
		}
	})

	return inst
}

func Get() *Setting {
	return New()
}

type Setting struct {
	setting *ssync.Map[string, *edgepb.Setting]
}

func (s *Setting) Store(st *edgepb.Setting) error {
	ns := st.GetServer().GetName()
	if _, ok := s.setting.Load(ns); ok {
		return errorx.F("[mem][setting][store]:namespace %s already exists", ns)
	}

	s.setting.Store(ns, st)
	return nil
}

func (s *Setting) Load(namespace string) (*edgepb.Setting, error) {
	result, ok := s.setting.Load(namespace)
	if ok {
		return result, nil
	}

	return nil, errorx.ErrNotFound
}

func (s *Setting) Visit(fn func(*edgepb.Setting) bool) {
	s.setting.Range(func(_ string, value *edgepb.Setting) bool {
		return fn(value)
	})
}

func Visit(fn func(*edgepb.Setting) bool) {
	if inst == nil {
		inst = New()
	}

	inst.Visit(fn)
}

func Store(st *edgepb.Setting) error {
	mu.Lock()
	defer mu.Unlock()

	if inst == nil {
		inst = New()
	}

	return inst.Store(st)
}
