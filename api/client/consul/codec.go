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

package consul

import (
	"encoding/json"
	"path"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Format is the encoding of a structured KV value.
type Format int

const (
	// FormatJSON encodes values as JSON.
	FormatJSON Format = iota + 1
	// FormatTOML encodes values as TOML. Only tables (structs and maps)
	// can be encoded at the top level.
	FormatTOML
)

type options struct {
	format Format
}

// Option configures a KV operation.
type Option interface {
	apply(*options)
}

type formatOption Format

func (f formatOption) apply(o *options) {
	o.format = Format(f)
}

// WithFormat forces the encoding of the value. Without it, the format
// is inferred from the key extension (".toml", ".json"), JSON otherwise.
func WithFormat(f Format) Option {
	return formatOption(f)
}

func resolveFormat(key string, opts []Option) Format {
	var o options
	for _, opt := range opts {
		opt.apply(&o)
	}
	if o.format != 0 {
		return o.format
	}
	if strings.EqualFold(path.Ext(key), ".toml") {
		return FormatTOML
	}
	return FormatJSON
}

func encode(value any, f Format) ([]byte, error) {
	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	}
	if f == FormatTOML {
		return toml.Marshal(value)
	}
	return json.Marshal(value)
}

func decode(raw []byte, dst any, f Format) error {
	switch v := dst.(type) {
	case *[]byte:
		*v = append([]byte(nil), raw...)
		return nil
	case *string:
		*v = string(raw)
		return nil
	}
	if f == FormatTOML {
		return toml.Unmarshal(raw, dst)
	}
	return json.Unmarshal(raw, dst)
}
