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
	"errors"
	"fmt"

	"github.com/sentinez/pkg/szprotobuf/protoaudit"
	"google.golang.org/protobuf/compiler/protogen"
)

// CheckAuditOwners returns an error for each method of file that sets
// x_method.audit, but neither audit_no_owner nor a field x_field.audit
// marks in its request or response: its audit logs would miss their
// tenant and resource.
func CheckAuditOwners(file *protogen.File) error {
	var errs []error
	for _, service := range file.Services {
		for _, method := range service.Methods {
			errs = append(errs, auditOwnerError(method))
		}
	}

	return errors.Join(errs...)
}

// auditOwnerError the error of method, nil unless audited without
// audit_no_owner nor a marked field in its request or response.
func auditOwnerError(method *protogen.Method) error {
	xmethod := getXMethod(method)
	if !xmethod.GetAudit() || xmethod.GetAuditNoOwner() ||
		len(protoaudit.OwnerPaths(method.Desc.Input())) > 0 ||
		len(protoaudit.OwnerPaths(method.Desc.Output())) > 0 {
		return nil
	}

	return fmt.Errorf("%s: audited, but no field of its request or "+
		"response sets x_field.audit, nor audit_no_owner",
		method.Desc.FullName())
}
