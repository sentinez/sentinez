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

package sz

import (
	"testing"

	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestAuditOwnerValid(t *testing.T) {
	tests := []struct {
		name       string
		giveMethod *typepb.XMethod
		giveMarked bool
	}{
		{"not audited", &typepb.XMethod{}, false},
		{"marked field", &typepb.XMethod{Audit: true}, true},
		{"no owner", &typepb.XMethod{Audit: true, AuditNoOwner: true}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := methodWith(t, tt.giveMethod, tt.giveMarked)

			assert.NoError(t, auditOwnerError(md))
		})
	}
}

func TestAuditOwnerInvalid(t *testing.T) {
	md := methodWith(t, &typepb.XMethod{Audit: true}, false)

	assert.ErrorContains(t, auditOwnerError(md), "test.S.M: audited")
}

// methodWith returns the method test.S.M of a message of a string id, its
// x_method option xmethod, the id marked a tenant when marked.
func methodWith(t *testing.T, xmethod *typepb.XMethod,
	marked bool) *protogen.Method {
	t.Helper()

	methodOpts := &descriptorpb.MethodOptions{}
	proto.SetExtension(methodOpts, typepb.E_XMethod, xmethod)
	fieldOpts := &descriptorpb.FieldOptions{}
	if marked {
		proto.SetExtension(fieldOpts, typepb.E_XField, &typepb.XField{
			Audit: &typepb.XAudit{TenantId: true},
		})
	}

	file, err := protodesc.NewFile(fileWith(methodOpts, fieldOpts), nil)
	require.NoError(t, err)

	return &protogen.Method{Desc: file.Services().Get(0).Methods().Get(0)}
}

// fileWith returns the file of the method test.S.M of a message test.Msg
// of a string id, their options methodOpts and fieldOpts.
func fileWith(methodOpts *descriptorpb.MethodOptions,
	fieldOpts *descriptorpb.FieldOptions) *descriptorpb.FileDescriptorProto {
	optional := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL

	return &descriptorpb.FileDescriptorProto{
		Name:    proto.String("test.proto"),
		Package: proto.String("test"),
		Syntax:  proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("Msg"),
			Field: []*descriptorpb.FieldDescriptorProto{{
				Name:    proto.String("id"),
				Number:  proto.Int32(1),
				Type:    descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Label:   optional.Enum(),
				Options: fieldOpts,
			}},
		}},
		Service: []*descriptorpb.ServiceDescriptorProto{{
			Name: proto.String("S"),
			Method: []*descriptorpb.MethodDescriptorProto{{
				Name:       proto.String("M"),
				InputType:  proto.String(".test.Msg"),
				OutputType: proto.String(".test.Msg"),
				Options:    methodOpts,
			}},
		}},
	}
}
