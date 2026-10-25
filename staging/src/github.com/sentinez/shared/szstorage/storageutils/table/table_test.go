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

package table_test

import (
	"strings"
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	_ "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	_ "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	_ "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	_ "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	"github.com/sentinez/shared/szstorage/storageutils/table"
	"github.com/sentinez/shared/sztable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// TestProtoIdPrefixesMatchTables keeps every buf.validate string prefix in
// the sentinez protos equal to an ID prefix the repos generate.
func TestProtoIdPrefixesMatchTables(t *testing.T) {
	keys := make(map[string]struct{})
	for _, name := range []string{
		sztable.IAMUsers, sztable.IAMAccounts,
		sztable.Tenants, sztable.TenantResources, sztable.TenantMembers,
		sztable.AnalyticActivities,
		sztable.SecuritySecRules, sztable.SecurityRateLimits,
		sztable.SecurityCoreRulesets,
	} {
		keys[table.NewPrimaryKey(name)] = struct{}{}
	}

	prefixes := protoPrefixes()
	require.NotEmpty(t, prefixes)

	for field, prefix := range prefixes {
		assert.Contains(t, keys, prefix, field)
	}
}

// protoPrefixes maps each sentinez field carrying a string prefix rule,
// directly or on its repeated items, to that prefix.
func protoPrefixes() map[string]string {
	prefixes := make(map[string]string)
	protoregistry.GlobalFiles.RangeFiles(
		func(fd protoreflect.FileDescriptor) bool {
			if strings.HasPrefix(string(fd.Package()), "sentinez.") {
				collectPrefixes(fd.Messages(), prefixes)
			}

			return true
		})

	return prefixes
}

func collectPrefixes(
	msgs protoreflect.MessageDescriptors, prefixes map[string]string) {
	for i := range msgs.Len() {
		md := msgs.Get(i)
		collectPrefixes(md.Messages(), prefixes)

		for j := range md.Fields().Len() {
			fd := md.Fields().Get(j)
			rules, ok := proto.GetExtension(
				fd.Options(), validate.E_Field).(*validate.FieldRules)
			if !ok {
				continue
			}

			prefix := rules.GetString().GetPrefix()
			if prefix == "" {
				prefix = rules.GetRepeated().GetItems().GetString().GetPrefix()
			}

			if prefix != "" {
				prefixes[string(fd.FullName())] = prefix
			}
		}
	}
}
