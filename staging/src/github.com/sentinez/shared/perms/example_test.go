package perms_test

import (
	"fmt"

	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/shared/perms"
)

func Example_member() {
	// Scenario: A Member trying to access a resource that requires VIEW_OWN permission.

	// 1. Create a claim for a Member who lacks the required permission.
	claim := perms.New(0) // No permissions
	role := typepb.Role_ROLE_MEMBER

	requires := []*typepb.XRequire{
		{Permission: typepb.Permission_PERMISSION_VIEW_OWN},
	}

	err := perms.Allow(requires, role, claim)
	if err == nil {
		fmt.Println("Member (no perms): Access allowed")
	} else {
		fmt.Println("Member (no perms): Access denied")
	}

	// 2. Grant the Member the required VIEW_OWN permission.
	claim = claim.Add(typepb.Permission_PERMISSION_VIEW_OWN)

	err = perms.Allow(requires, role, claim)
	if err == nil {
		fmt.Println("Member (with perms): Access allowed")
	} else {
		fmt.Println("Member (with perms): Access denied")
	}

	// Output:
	// Member (no perms): Access denied
	// Member (with perms): Access allowed
}

func Example_leader() {
	// Scenario: Accessing a resource that requires either Leader role OR VIEW_ANY permission.

	requires := []*typepb.XRequire{
		{Role: typepb.Role_ROLE_LEADER},
		{Permission: typepb.Permission_PERMISSION_VIEW_ANY},
	}

	// 1. A Leader without specific permissions.
	// Since the requirement allows ROLE_LEADER, access should be granted regardless of permissions.
	leaderClaim := perms.New(0)
	leaderRole := typepb.Role_ROLE_LEADER

	err := perms.Allow(requires, leaderRole, leaderClaim)
	if err == nil {
		fmt.Println("Leader: Access allowed")
	} else {
		fmt.Println("Leader: Access denied")
	}

	// 2. A Member with ROLE_MEMBER (not leader) and no permissions.
	// Should be denied.
	memberClaim := perms.New(0)
	memberRole := typepb.Role_ROLE_MEMBER

	err = perms.Allow(requires, memberRole, memberClaim)
	if err == nil {
		fmt.Println("Member: Access allowed")
	} else {
		fmt.Println("Member: Access denied")
	}

	// Output:
	// Leader: Access allowed
	// Member: Access denied
}
