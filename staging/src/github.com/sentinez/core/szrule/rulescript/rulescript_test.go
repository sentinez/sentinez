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

package rulescript

import (
	"testing"

	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func cond(src rulepb.FieldSource, key string, op rulepb.Operator,
	v *structpb.Value) *rulepb.Rule {
	return &rulepb.Rule{Condition: &rulepb.Condition{
		Source: src, Key: key, Operator: op, Value: v,
	}}
}

func str(s string) *structpb.Value { return structpb.NewStringValue(s) }

func list(t *testing.T, s ...any) *structpb.Value {
	t.Helper()

	v, err := structpb.NewList(s)
	require.NoError(t, err)

	return structpb.NewListValue(v)
}

const (
	_path   = rulepb.FieldSource_FIELD_SOURCE_PATH
	_method = rulepb.FieldSource_FIELD_SOURCE_METHOD
	_ip     = rulepb.FieldSource_FIELD_SOURCE_IP
	_header = rulepb.FieldSource_FIELD_SOURCE_HEADER
	_query  = rulepb.FieldSource_FIELD_SOURCE_QUERY
	_tls    = rulepb.FieldSource_FIELD_SOURCE_TLS
)
