// Copyright 2025 Duc-Hung Ho.
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

// Package types provides the types for the service.
package types

import (
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
	fieldmask "google.golang.org/protobuf/types/known/fieldmaskpb"
)

// Empty is an alias of emptypb.Empty.
type Empty = emptypb.Empty

// Any is an alias of anypb.Any.
type Any = anypb.Any

// FieldMask is an alias of fieldmask.FieldMask.
type FieldMask = fieldmask.FieldMask
