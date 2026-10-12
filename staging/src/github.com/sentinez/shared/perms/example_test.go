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

package perms_test

import (
	"fmt"

	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/perms"
)

func ExampleAllow_consoleMatch() {
	viewMethod := &typepb.XMethod{
		ControlPlanes: []typepb.ControlPlane{
			typepb.ControlPlane_CONTROL_PLANE_PORTAL,
			typepb.ControlPlane_CONTROL_PLANE_ADMIN,
		},
	}

	// User has PORTAL console
	userConsole := typepb.ControlPlane_CONTROL_PLANE_PORTAL

	err := perms.Allow(viewMethod, userConsole)
	if err == nil {
		fmt.Println("Access granted")
	} else {
		fmt.Println("Access denied")
	}
	//Output: Access granted
}

func ExampleAllow_consoleMismatch() {
	adminMethod := &typepb.XMethod{
		ControlPlanes: []typepb.ControlPlane{
			typepb.ControlPlane_CONTROL_PLANE_ADMIN,
		},
	}

	// User has PORTAL console
	userConsole := typepb.ControlPlane_CONTROL_PLANE_PORTAL

	err := perms.Allow(adminMethod, userConsole)
	if err != nil {
		fmt.Println("Access denied")
	}
	//Output: Access denied
}

func ExampleAllow_ignoredMethod() {
	ignoredMethod := &typepb.XMethod{
		Ignore: true,
	}

	// Even if user has no matching console, ignored methods allow access
	userConsole := typepb.ControlPlane_CONTROL_PLANE_UNSPECIFIED

	err := perms.Allow(ignoredMethod, userConsole)
	if err == nil {
		fmt.Println("Access granted")
	}
	//Output: Access granted
}

func ExampleAllow_noRestrictions() {
	// Method with no consoles specified - allows all
	publicMethod := &typepb.XMethod{}

	userConsole := typepb.ControlPlane_CONTROL_PLANE_PORTAL
	err := perms.Allow(publicMethod, userConsole)
	if err == nil {
		fmt.Println("Access granted")
	}
	//Output: Access granted
}
