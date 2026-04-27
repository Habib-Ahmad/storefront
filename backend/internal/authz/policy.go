package authz

import (
	"sort"

	"storefront/backend/internal/models"
)

type Permission string

const (
	PermissionSelfProfileRead       Permission = "self:profile:read"
	PermissionSelfProfileManage     Permission = "self:profile:manage"
	PermissionTenantProfileManage   Permission = "tenant:profile:manage"
	PermissionStorefrontManage      Permission = "tenant:storefront:manage"
	PermissionModulesManage         Permission = "tenant:modules:manage"
	PermissionProductsRead          Permission = "products:read"
	PermissionProductsManage        Permission = "products:manage"
	PermissionProductVariantsRead   Permission = "product_variants:read"
	PermissionProductVariantsManage Permission = "product_variants:manage"
	PermissionProductImagesRead     Permission = "product_images:read"
	PermissionProductImagesManage   Permission = "product_images:manage"
	PermissionOrdersCreate          Permission = "orders:create"
	PermissionOrdersRead            Permission = "orders:read"
	PermissionOrdersManage          Permission = "orders:manage"
	PermissionOrdersDispatch        Permission = "orders:dispatch"
	PermissionWalletRead            Permission = "wallet:read"
	PermissionAnalyticsRead         Permission = "analytics:read"
	PermissionMediaUpload           Permission = "media:upload"
)

type RolePolicy struct {
	Permissions []Permission
}

var rolePolicies = map[models.UserRole]RolePolicy{
	models.UserRoleAdmin: {
		Permissions: []Permission{
			PermissionSelfProfileRead,
			PermissionSelfProfileManage,
			PermissionTenantProfileManage,
			PermissionStorefrontManage,
			PermissionModulesManage,
			PermissionProductsRead,
			PermissionProductsManage,
			PermissionProductVariantsRead,
			PermissionProductVariantsManage,
			PermissionProductImagesRead,
			PermissionProductImagesManage,
			PermissionOrdersCreate,
			PermissionOrdersRead,
			PermissionOrdersManage,
			PermissionOrdersDispatch,
			PermissionWalletRead,
			PermissionAnalyticsRead,
			PermissionMediaUpload,
		},
	},
	models.UserRoleStaff: {
		Permissions: []Permission{
			PermissionSelfProfileRead,
			PermissionSelfProfileManage,
			PermissionProductsRead,
			PermissionProductsManage,
			PermissionProductVariantsRead,
			PermissionProductVariantsManage,
			PermissionProductImagesRead,
			PermissionProductImagesManage,
			PermissionOrdersCreate,
			PermissionOrdersRead,
			PermissionOrdersManage,
			PermissionOrdersDispatch,
			PermissionWalletRead,
			PermissionAnalyticsRead,
			PermissionMediaUpload,
		},
	},
}

func Policies() map[models.UserRole]RolePolicy {
	out := make(map[models.UserRole]RolePolicy, len(rolePolicies))
	for role, policy := range rolePolicies {
		permissions := make([]Permission, len(policy.Permissions))
		copy(permissions, policy.Permissions)
		out[role] = RolePolicy{Permissions: permissions}
	}
	return out
}

func PermissionsForRole(role models.UserRole) []Permission {
	policy, ok := rolePolicies[role]
	if !ok {
		return nil
	}

	permissions := make([]Permission, len(policy.Permissions))
	copy(permissions, policy.Permissions)
	sort.Slice(permissions, func(i, j int) bool {
		return permissions[i] < permissions[j]
	})
	return permissions
}

func HasPermission(role models.UserRole, permission Permission) bool {
	policy, ok := rolePolicies[role]
	if !ok {
		return false
	}

	for _, granted := range policy.Permissions {
		if granted == permission {
			return true
		}
	}

	return false
}
