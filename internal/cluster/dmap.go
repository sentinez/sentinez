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

package cluster

import (
	"context"
	"fmt"

	"github.com/olric-data/olric"
	"github.com/sentinez/shared/zlog"
	"google.golang.org/protobuf/proto"
)

const (
	defaultChanBufferSize = 1024
)

type objKV[T any] struct {
	value *T
	key   string
}

func NewDMap[T any](name string) *DMap[T] {
	m := &DMap[T]{
		name: name,
		temp: make(chan objKV[T], defaultChanBufferSize),
	}
	go m.wait()
	return m
}

type DMap[T any] struct {
	name string
	temp chan objKV[T]
	dmap olric.DMap
}

func (m *DMap[T]) wait() {
	zlog.Infof("dmap: %s waitting to connection", m.name)
	<-dictReady

	var err error
	m.dmap, err = dict.client.NewDMap(m.name)
	if err != nil {
		zlog.Errorf("dmap: create %s got err: %v", m.name, err)
	}

	zlog.Infof("dmap: create %s successfully", m.name)
	for obj := range m.temp {
		err = m.Put(context.Background(), obj.key, obj.value)
		if err != nil {
			zlog.Errorf("dmap: put %s got err: %v", obj.key, err)
		}
	}
}

func (m *DMap[T]) Get(ctx context.Context, key string) (*T, error) {
	if m == nil || m.dmap == nil {
		return nil, fmt.Errorf("dmap is nil")
	}

	resp, err := m.dmap.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	buf, err := resp.Byte()
	if err != nil {
		return nil, err
	}

	t := new(T)
	msg, ok := any(t).(proto.Message)
	if !ok {
		return nil, fmt.Errorf("message is not proto.Message")
	}

	if err := proto.Unmarshal(buf, msg); err != nil {
		return nil, err
	}

	return t, nil
}

func (m *DMap[T]) Put(ctx context.Context, key string, value *T) error {
	if m == nil {
		return fmt.Errorf("mapp is nil")
	}

	if m.dmap == nil {
		m.temp <- objKV[T]{key: key, value: value}
		return nil
	}

	msg, ok := any(value).(proto.Message)
	if !ok {
		return fmt.Errorf("message is not proto.Message")
	}

	buf, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return m.dmap.Put(ctx, key, buf)
}
