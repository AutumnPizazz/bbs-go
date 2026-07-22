package services

import (
	"bbs-go/internal/models/constants"
	"bbs-go/internal/permissions"
	"testing"
)

func TestValidateManagedRoleAssignmentsRejectsOwnerRoleForNonOwner(t *testing.T) {
	setupPermissionServiceTestDB(t)
	ownerRole := mustCreateRole(t, constants.RoleOwner, constants.StatusOk)
	operator := mustCreateUser(t, 1)

	if err := validateManagedRoleAssignments(operator, []int64{ownerRole.Id}); err == nil {
		t.Fatal("expected non-owner to be unable to assign the owner role")
	}
}

func TestValidateManagedRoleAssignmentsRejectsPermissionsBeyondOperator(t *testing.T) {
	setupPermissionServiceTestDB(t)
	operatorRole := mustCreateRole(t, "user-manager", constants.StatusOk)
	elevatedRole := mustCreateRole(t, "scope-manager", constants.StatusOk)
	accessScope := mustCreatePermission(t, permissions.PermissionUserAccessScope.Code, constants.StatusOk)
	mustGrantPermission(t, elevatedRole, accessScope)
	operator := mustCreateUser(t, 2)
	mustAssignRole(t, operator, operatorRole)

	if err := validateManagedRoleAssignments(operator, []int64{elevatedRole.Id}); err == nil {
		t.Fatal("expected role assignment beyond operator permissions to be rejected")
	}
}
