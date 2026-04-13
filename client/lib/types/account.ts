import * as z from "zod";
import { TierSchema, TenantSchema, UserSchema } from "./domain";

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
