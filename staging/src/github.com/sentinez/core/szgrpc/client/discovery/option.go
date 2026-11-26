// Copyright 2026 Duc-Hung Ho.
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

package discovery

type options struct {
	consulURL string
}

// Option configures a Discovery.
type Option interface {
	apply(*options)
}

type consulURLOption string

func (u consulURLOption) apply(o *options) {
	o.consulURL = string(u)
}

// WithConsulURL sets the address of the consul agent the services are
// registered to and discovered from, ex: SZ_CONSUL_URI.
func WithConsulURL(uri string) Option {
	return consulURLOption(uri)
}

func newOptions(opts []Option) options {
	var o options
	for _, opt := range opts {
		opt.apply(&o)
	}

	return o
}
