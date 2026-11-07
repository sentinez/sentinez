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

package ingress

import (
	"io"

	"github.com/sentinez/pkg/szsync"
	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	"github.com/sentinez/sentinez/pkg/geoip"
)

var (
	_ io.Closer           = (*IngressEvent)(nil)
	_ geoip.CountrySetter = (*IngressEvent)(nil)
	_ geoip.AsnSetter     = (*IngressEvent)(nil)

	_pool = szsync.NewPoolCtr(func() *IngressEvent {
		return &IngressEvent{IngressEvent: &httppb.IngressEvent{}}
	})
)

// IngressEvent a pooled httppb.IngressEvent, released on Close.
type IngressEvent struct {
	*httppb.IngressEvent
}

// Close releases ie to the pool.
func (ie *IngressEvent) Close() error {
	Release(ie)

	return nil
}

// SetCountry asks geoip.DB.Fill for the client country.
func (ie *IngressEvent) SetCountry(isoCode string) {
	ie.Country = isoCode
}

// SetAsn asks geoip.DB.Fill for the client autonomous system.
func (ie *IngressEvent) SetAsn(number int64) {
	ie.Asn = number
}

// Acquire returns a reset IngressEvent of the pool.
func Acquire() *IngressEvent {
	return _pool.Get()
}

// Release resets obj and puts it back in the pool.
func Release(obj *IngressEvent) {
	obj.ResetVT()
	_pool.Put(obj)
}
