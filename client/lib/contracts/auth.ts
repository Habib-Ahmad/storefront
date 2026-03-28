import { z } from "zod";

// ── Reusable primitives ────────────────────────────────

export const UUIDSchema = z.string().uuid();
export const TimestampSchema = z.string();

// ── Enums ──────────────────────────────────────────────

export const TenantStatusSchema = z.enum(["active", "suspended"]);
export const UserRoleSchema = z.enum(["admin", "staff", "manager"]);

export type TenantStatus = z.infer<typeof TenantStatusSchema>;
export type UserRole = z.infer<typeof UserRoleSchema>;

// ── Shared domain schemas for auth/onboarding ──────────

export const ActiveModulesSchema = z.object({
  inventory: z.boolean(),
  payments: z.boolean(),
  logistics: z.boolean(),
});

export const TierSchema = z.object({
  id: UUIDSchema,
  name: z.string(),
  debt_ceiling: z.string(),
  commission_rate: z.string(),
  created_at: TimestampSchema,
  updated_at: TimestampSchema,
});

export const TenantSchema = z.object({
  id: UUIDSchema,
  tier_id: UUIDSchema,
  name: z.string(),
  slug: z.string(),
  contact_email: z.string().nullable().optional(),
  contact_phone: z.string().nullable().optional(),
  address: z.string().nullable().optional(),
  logo_url: z.string().nullable().optional(),
  paystack_subaccount_id: z.string().nullable().optional(),
  active_modules: ActiveModulesSchema,
  status: TenantStatusSchema,
  created_at: TimestampSchema,
  updated_at: TimestampSchema,
});

// ── Auth response schemas ──────────────────────────────

/**
 * Backend shape from GET /auth/me
 *
 * Current backend behavior:
 * - when not onboarded: { onboarded: false }
 * - when onboarded: { onboarded: true, tenant: Tenant, role: UserRole }
 */
export const MeResponseSchema = z.union([
  z.object({
    onboarded: z.literal(false),
  }),
  z.object({
    onboarded: z.literal(true),
    tenant: TenantSchema,
    role: UserRoleSchema,
  }),
]);

// ── Onboarding / tenant request schemas ────────────────

export const OnboardRequestSchema = z.object({
  name: z.string(),
  slug: z.string(),
  admin_email: z.email(),
});

export const UpdateTenantRequestSchema = z.object({
  name: z.string(),
  contact_email: z.email().nullable().optional(),
  contact_phone: z.string().nullable().optional(),
  address: z.string().nullable().optional(),
  logo_url: z.url().nullable().optional(),
});

export const SetModulesRequestSchema = ActiveModulesSchema;

// ── Inferred types ─────────────────────────────────────

export type ActiveModules = z.infer<typeof ActiveModulesSchema>;
export type Tier = z.infer<typeof TierSchema>;
export type Tenant = z.infer<typeof TenantSchema>;
export type MeResponse = z.infer<typeof MeResponseSchema>;
export type OnboardRequest = z.infer<typeof OnboardRequestSchema>;
export type UpdateTenantRequest = z.infer<typeof UpdateTenantRequestSchema>;
export type SetModulesRequest = z.infer<typeof SetModulesRequestSchema>;
