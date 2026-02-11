package perms

import (
	"testing"

	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
)

func TestAllow(t *testing.T) {
	tests := []struct {
		name      string
		method    *typepb.XMethod
		console   typepb.Console
		wantError bool
	}{
		{
			name:      "Nil method",
			method:    nil,
			console:   typepb.Console_CONSOLE_PORTAL,
			wantError: false,
		},
		{
			name: "Ignored method",
			method: &typepb.XMethod{
				Ignore: true,
			},
			console:   typepb.Console_CONSOLE_PORTAL,
			wantError: false,
		},
		{
			name:      "Empty consoles list",
			method:    &typepb.XMethod{},
			console:   typepb.Console_CONSOLE_PORTAL,
			wantError: false,
		},
		{
			name: "Console match",
			method: &typepb.XMethod{
				Consoles: []typepb.Console{typepb.Console_CONSOLE_PORTAL},
			},
			console:   typepb.Console_CONSOLE_PORTAL,
			wantError: false,
		},
		{
			name: "Console mismatch",
			method: &typepb.XMethod{
				Consoles: []typepb.Console{typepb.Console_CONSOLE_ADMIN},
			},
			console:   typepb.Console_CONSOLE_PORTAL,
			wantError: true,
		},
		{
			name: "Multiple consoles - match",
			method: &typepb.XMethod{
				Consoles: []typepb.Console{
					typepb.Console_CONSOLE_PORTAL,
					typepb.Console_CONSOLE_ADMIN,
				},
			},
			console:   typepb.Console_CONSOLE_PORTAL,
			wantError: false,
		},
		{
			name: "Multiple consoles - no match",
			method: &typepb.XMethod{
				Consoles: []typepb.Console{
					typepb.Console_CONSOLE_ADMIN,
				},
			},
			console:   typepb.Console_CONSOLE_PORTAL,
			wantError: true,
		},
		{
			name: "Ignored method with consoles still allows access",
			method: &typepb.XMethod{
				Ignore: true,
				Consoles: []typepb.Console{
					typepb.Console_CONSOLE_ADMIN,
				},
			},
			console:   typepb.Console_CONSOLE_PORTAL,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Allow(tt.method, tt.console)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
