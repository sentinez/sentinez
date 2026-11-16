// Copyright 2026 Duc-Hung Ho.
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

package szprotobuf

import (
	"slices"

	"github.com/sentinez/pkg/szerror"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// MaskAll is the path of an update mask that names every field.
const MaskAll = "*"

// ApplyMask copies onto dst the fields of src that mask names, read as
// AIP-134 reads the update_mask of an update, and returns their names:
//   - no path: those of updatable set in src (an implied mask)
//   - "*": all of updatable
//   - otherwise its paths, each one of updatable
//
// A named field unset in src is cleared on dst. updatable are top-level
// fields, named as in the .proto, of dst and src, messages of the same
// type; the values copied are shared with src, not cloned. A path not
// in updatable is an InvalidArgument error and leaves dst unchanged.
func ApplyMask(dst, src proto.Message, mask *fieldmaskpb.FieldMask,
	updatable ...string) ([]string, error) {

	paths, err := maskPaths(src, mask, updatable)
	if err != nil {
		return nil, err
	}

	from, to := src.ProtoReflect(), dst.ProtoReflect()
	fields := to.Descriptor().Fields()
	for _, path := range paths {
		fd := fields.ByName(protoreflect.Name(path))
		switch {
		case fd == nil:
			// updatable names no field of dst: nothing to copy
			continue
		case !from.Has(fd):
			to.Clear(fd)
		default:
			to.Set(fd, from.Get(fd))
		}
	}

	return paths, nil
}

// maskPaths the fields of src mask names, checked against updatable.
func maskPaths(src proto.Message, mask *fieldmaskpb.FieldMask,
	updatable []string) ([]string, error) {

	paths := mask.GetPaths()
	switch {
	case len(paths) == 0:
		return setFields(src, updatable), nil
	case slices.Contains(paths, MaskAll):
		return slices.Clone(updatable), nil
	}

	for _, path := range paths {
		if !slices.Contains(updatable, path) {
			return nil, szerror.StatusInvalidArgumentF(
				"update_mask: %q is not an updatable field", path)
		}
	}

	// sorted, once each; the request keeps its own order
	named := &fieldmaskpb.FieldMask{Paths: slices.Clone(paths)}
	named.Normalize()

	return named.GetPaths(), nil
}

// setFields the fields of updatable that are set in m.
func setFields(m proto.Message, updatable []string) []string {
	msg := m.ProtoReflect()
	fields := msg.Descriptor().Fields()

	var set []string
	for _, name := range updatable {
		fd := fields.ByName(protoreflect.Name(name))
		if fd != nil && msg.Has(fd) {
			set = append(set, name)
		}
	}

	return set
}
