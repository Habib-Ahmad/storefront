package authz_test

import (
	"testing"

	"storefront/backend/internal/authz"
	"storefront/backend/internal/models"
)

func TestPolicies_CoverEveryKnownRole(t *testing.T) {
	policies := authz.Policies()

	for _, role := range []models.UserRole{
		models.UserRoleAdmin,
		models.UserRoleStaff,
	} {
		policy, ok := policies[role]
		if !ok {
			t.Fatalf("missing policy for role %s", role)
		}
		if len(policy.Permissions) == 0 {
			t.Fatalf("expected role %s to have at least one permission", role)
		}
	}

	if len(policies) != 2 {
		t.Fatalf("expected exactly 2 role policies, got %d", len(policies))
	}
}

func TestHasPermission_EnforcesLeastPrivilegeShape(t *testing.T) {
	if !authz.HasPermission(models.UserRoleAdmin, authz.PermissionModulesManage) {
		t.Fatal("expected admin to manage modules")
	}
	if authz.HasPermission(models.UserRoleStaff, authz.PermissionModulesManage) {
		t.Fatal("did not expect staff to manage modules")
	}
	if authz.HasPermission(models.UserRoleStaff, authz.PermissionStorefrontManage) {
		t.Fatal("did not expect staff to manage storefront settings")
	}
	if !authz.HasPermission(models.UserRoleStaff, authz.PermissionProductsManage) {
		t.Fatal("expected staff to manage products")
	}
	if !authz.HasPermission(models.UserRoleStaff, authz.PermissionProductVariantsManage) {
		t.Fatal("expected staff to manage product variants")
	}
	if !authz.HasPermission(models.UserRoleStaff, authz.PermissionProductImagesManage) {
		t.Fatal("expected staff to manage product images")
	}
	if !authz.HasPermission(models.UserRoleStaff, authz.PermissionOrdersDispatch) {
		t.Fatal("expected staff to dispatch orders")
	}
	if !authz.HasPermission(models.UserRoleStaff, authz.PermissionOrdersManage) {
		t.Fatal("expected staff to manage orders")
	}
	if authz.HasPermission(models.UserRole("unknown"), authz.PermissionOrdersRead) {
		t.Fatal("did not expect unknown role to inherit permissions")
	}
}
