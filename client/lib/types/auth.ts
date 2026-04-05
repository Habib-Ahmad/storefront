import { z } from "zod";
import { ActiveModulesSchema, TenantSchema, UserRoleSchema } from "./domain";

export {
  UUIDSchema,
  TimestampSchema,
  TenantStatusSchema,
  UserRoleSchema,
  ActiveModulesSchema,
  TierSchema,
  TenantSchema,
} from "./domain";

export type { ActiveModules, Tenant, TenantStatus, Tier, UserRole } from "./domain";

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
  admin_email: z.string().email(),
});

export const UpdateTenantRequestSchema = z.object({
  name: z.string(),
  contact_email: z.string().email().nullable().optional(),
  contact_phone: z.string().nullable().optional(),
  address: z.string().nullable().optional(),
  logo_url: z.string().url().nullable().optional(),
});

export const UpdateStorefrontRequestSchema = z.object({
  slug: z.string(),
  storefront_published: z.boolean(),
});

export const SetModulesRequestSchema = ActiveModulesSchema;

// ── Inferred types ─────────────────────────────────────

export type MeResponse = z.infer<typeof MeResponseSchema>;
export type OnboardRequest = z.infer<typeof OnboardRequestSchema>;
export type UpdateTenantRequest = z.infer<typeof UpdateTenantRequestSchema>;
export type UpdateStorefrontRequest = z.infer<typeof UpdateStorefrontRequestSchema>;
export type SetModulesRequest = z.infer<typeof SetModulesRequestSchema>;
