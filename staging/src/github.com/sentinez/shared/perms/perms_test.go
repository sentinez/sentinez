package perms

import (
	"testing"

	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
)

func TestClaim_Operations(t *testing.T) {
	t.Run("New Claim", func(t *testing.T) {
		c := New(int32(typepb.Permission_PERMISSION_VIEW_OWN))
		assert.Equal(t, int32(typepb.Permission_PERMISSION_VIEW_OWN), c.Int32())
	})

	t.Run("Add Permission", func(t *testing.T) {
		c := New(0)
		c = c.Add(typepb.Permission_PERMISSION_VIEW_OWN)
		assert.True(t, c.Has(typepb.Permission_PERMISSION_VIEW_OWN))
	})

	t.Run("Remove Permission", func(t *testing.T) {
		c := New(int32(typepb.Permission_PERMISSION_VIEW_OWN | typepb.Permission_PERMISSION_UPDATE_OWN))
		c = c.Remove(typepb.Permission_PERMISSION_UPDATE_OWN)
		assert.True(t, c.Has(typepb.Permission_PERMISSION_VIEW_OWN))
		assert.False(t, c.Has(typepb.Permission_PERMISSION_UPDATE_OWN))
	})

	t.Run("Has Permission", func(t *testing.T) {
		c := New(int32(typepb.Permission_PERMISSION_VIEW_OWN))
		assert.True(t, c.Has(typepb.Permission_PERMISSION_VIEW_OWN))
		assert.False(t, c.Has(typepb.Permission_PERMISSION_UPDATE_OWN))
	})

	t.Run("HasAny", func(t *testing.T) {
		c := New(int32(typepb.Permission_PERMISSION_VIEW_OWN))
		assert.True(t, c.HasAny(New(int32(typepb.Permission_PERMISSION_VIEW_OWN|typepb.Permission_PERMISSION_UPDATE_OWN))))
		assert.False(t, c.HasAny(New(int32(typepb.Permission_PERMISSION_UPDATE_OWN))))
	})

	t.Run("HasAll", func(t *testing.T) {
		c := New(int32(typepb.Permission_PERMISSION_VIEW_OWN | typepb.Permission_PERMISSION_UPDATE_OWN))
		assert.True(t, c.HasAll(New(int32(typepb.Permission_PERMISSION_VIEW_OWN|typepb.Permission_PERMISSION_UPDATE_OWN))))
		assert.False(t, c.HasAll(New(int32(typepb.Permission_PERMISSION_VIEW_OWN|typepb.Permission_PERMISSION_UPDATE_OWN|typepb.Permission_PERMISSION_DELETE_OWN))))
	})
}

func TestAllow(t *testing.T) {
	tests := []struct {
		name      string
		requires  []*typepb.XRequire
		role      typepb.Role
		claim     Claim
		wantError bool
	}{
		{
			name:      "Empty requirements",
			requires:  []*typepb.XRequire{},
			role:      typepb.Role_ROLE_MEMBER,
			claim:     New(0),
			wantError: false,
		},
		{
			name: "Role match",
			requires: []*typepb.XRequire{
				{Role: typepb.Role_ROLE_LEADER},
			},
			role:      typepb.Role_ROLE_LEADER,
			claim:     New(0),
			wantError: false,
		},
		{
			name: "Role mismatch",
			requires: []*typepb.XRequire{
				{Role: typepb.Role_ROLE_LEADER},
			},
			role:      typepb.Role_ROLE_MEMBER,
			claim:     New(0),
			wantError: true,
		},
		{
			name: "Permission match",
			requires: []*typepb.XRequire{
				{Permission: typepb.Permission_PERMISSION_VIEW_OWN},
			},
			role:      typepb.Role_ROLE_MEMBER,
			claim:     New(int32(typepb.Permission_PERMISSION_VIEW_OWN)),
			wantError: false,
		},
		{
			name: "Permission mismatch",
			requires: []*typepb.XRequire{
				{Permission: typepb.Permission_PERMISSION_VIEW_OWN},
			},
			role:      typepb.Role_ROLE_MEMBER,
			claim:     New(0),
			wantError: true,
		},
		{
			name: "Role AND Permission match",
			requires: []*typepb.XRequire{
				{
					Role:       typepb.Role_ROLE_LEADER,
					Permission: typepb.Permission_PERMISSION_VIEW_OWN,
				},
			},
			role:      typepb.Role_ROLE_LEADER,
			claim:     New(int32(typepb.Permission_PERMISSION_VIEW_OWN)),
			wantError: false,
		},
		{
			name: "Role match but Permission mismatch",
			requires: []*typepb.XRequire{
				{
					Role:       typepb.Role_ROLE_LEADER,
					Permission: typepb.Permission_PERMISSION_VIEW_OWN,
				},
			},
			role:      typepb.Role_ROLE_LEADER,
			claim:     New(0),
			wantError: true,
		},
		{
			name: "Permission match but Role mismatch",
			requires: []*typepb.XRequire{
				{
					Role:       typepb.Role_ROLE_LEADER,
					Permission: typepb.Permission_PERMISSION_VIEW_OWN,
				},
			},
			role:      typepb.Role_ROLE_MEMBER,
			claim:     New(int32(typepb.Permission_PERMISSION_VIEW_OWN)),
			wantError: true,
		},
		{
			name: "Multiple requirements (OR logic) - First match",
			requires: []*typepb.XRequire{
				{Role: typepb.Role_ROLE_LEADER},
				{Permission: typepb.Permission_PERMISSION_VIEW_OWN},
			},
			role:      typepb.Role_ROLE_LEADER,
			claim:     New(0),
			wantError: false,
		},
		{
			name: "Multiple requirements (OR logic) - Second match",
			requires: []*typepb.XRequire{
				{Role: typepb.Role_ROLE_LEADER},
				{Permission: typepb.Permission_PERMISSION_VIEW_OWN},
			},
			role:      typepb.Role_ROLE_MEMBER,
			claim:     New(int32(typepb.Permission_PERMISSION_VIEW_OWN)),
			wantError: false,
		},
		{
			name: "Multiple requirements (OR logic) - No match",
			requires: []*typepb.XRequire{
				{Role: typepb.Role_ROLE_LEADER},
				{Permission: typepb.Permission_PERMISSION_VIEW_OWN},
			},
			role:      typepb.Role_ROLE_MEMBER,
			claim:     New(0),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Allow(tt.requires, tt.role, tt.claim)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
