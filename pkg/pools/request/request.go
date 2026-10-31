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

package request

import (
	"io"

	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	"github.com/sentinez/sentinez/pkg/geoip"
	"github.com/sentinez/shared/szsync"
)

var (
	_ io.Closer           = (*RequestEvent)(nil)
	_ geoip.CountrySetter = (*RequestEvent)(nil)
	_ geoip.AsnSetter     = (*RequestEvent)(nil)

	_pool = szsync.NewPoolCtr(func() *RequestEvent {
		return &RequestEvent{Event: &httppb.Event{}}
	})
)

type RequestEvent struct {
	*httppb.Event
}

func (re *RequestEvent) Close() error {
	Release(re)

	return nil
}

// SetCountry asks geoip.DB.Fill for the client country.
func (re *RequestEvent) SetCountry(isoCode string) {
	re.Country = isoCode
}

// SetAsn asks geoip.DB.Fill for the client autonomous system.
func (re *RequestEvent) SetAsn(number int64) {
	re.Asn = number
}

func Acquire() *RequestEvent {
	return _pool.Get()
}

func Release(obj *RequestEvent) {
	obj.Reset()
	_pool.Put(obj)
}
