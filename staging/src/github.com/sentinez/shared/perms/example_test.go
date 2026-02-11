package perms_test

import (
	"fmt"

	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/shared/perms"
)

func ExampleAllow_consoleMatch() {
	viewMethod := &typepb.XMethod{
		Consoles: []typepb.Console{
			typepb.Console_CONSOLE_PORTAL,
			typepb.Console_CONSOLE_ADMIN,
		},
	}

	// User has PORTAL console
	userConsole := typepb.Console_CONSOLE_PORTAL

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
		Consoles: []typepb.Console{
			typepb.Console_CONSOLE_ADMIN,
		},
	}

	// User has PORTAL console
	userConsole := typepb.Console_CONSOLE_PORTAL

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
	userConsole := typepb.Console_CONSOLE_UNSPECIFIED

	err := perms.Allow(ignoredMethod, userConsole)
	if err == nil {
		fmt.Println("Access granted")
	}
	//Output: Access granted
}

func ExampleAllow_noRestrictions() {
	// Method with no consoles specified - allows all
	publicMethod := &typepb.XMethod{}

	userConsole := typepb.Console_CONSOLE_PORTAL
	err := perms.Allow(publicMethod, userConsole)
	if err == nil {
		fmt.Println("Access granted")
	}
	//Output: Access granted
}
