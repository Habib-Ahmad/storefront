import * as z from "zod";

// ── Reusable primitives ────────────────────────────────

export const UUIDSchema = z.string().uuid();
export const TimestampSchema = z.string();

// ── Enums ──────────────────────────────────────────────

export const TenantStatusSchema = z.enum(["active", "suspended"]);
export const UserRoleSchema = z.enum(["admin", "staff"]);

export type TenantStatus = z.infer<typeof TenantStatusSchema>;
export type UserRole = z.infer<typeof UserRoleSchema>;

// ── Shared account/domain schemas ──────────────────────

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
  commission_cap: z.string().optional(),
  created_at: TimestampSchema,
  updated_at: TimestampSchema,
});

export const TenantSchema = z.object({
  id: UUIDSchema,
  tier_id: UUIDSchema,
  name: z.string(),
  slug: z.string(),
  storefront_published: z.boolean(),
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

export const UserSchema = z.object({
  id: UUIDSchema,
  tenant_id: UUIDSchema,
  email: z.string().email(),
  first_name: z.string().nullable().optional(),
  last_name: z.string().nullable().optional(),
  phone: z.string().nullable().optional(),
  role: UserRoleSchema,
  created_at: TimestampSchema,
  updated_at: TimestampSchema,
});

// ── Inferred types ─────────────────────────────────────

export type ActiveModules = z.infer<typeof ActiveModulesSchema>;
export type Tier = z.infer<typeof TierSchema>;
export type Tenant = z.infer<typeof TenantSchema>;
export type User = z.infer<typeof UserSchema>;
