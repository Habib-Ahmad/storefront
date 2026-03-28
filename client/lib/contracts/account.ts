import { z } from "zod";
import { TierSchema, TenantSchema, UserRoleSchema, UUIDSchema, TimestampSchema } from "./auth";

// ── Account / settings domain schemas ──────────────────

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

// ── Account / settings request schemas ─────────────────

export const UpdateUserRequestSchema = z.object({
  first_name: z.string().nullable().optional(),
  last_name: z.string().nullable().optional(),
  phone: z.string().nullable().optional(),
});

// ── Response schemas ───────────────────────────────────

export const TiersResponseSchema = z.array(TierSchema);

// ── Inferred types ─────────────────────────────────────

export type Tier = z.infer<typeof TierSchema>;
export type Tenant = z.infer<typeof TenantSchema>;
export type User = z.infer<typeof UserSchema>;
export type UpdateUserRequest = z.infer<typeof UpdateUserRequestSchema>;
export type TiersResponse = z.infer<typeof TiersResponseSchema>;
