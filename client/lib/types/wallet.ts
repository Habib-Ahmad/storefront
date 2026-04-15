import * as z from "zod";
import { UUIDSchema, TimestampSchema } from "./domain";
import { createPaginatedResponseSchema } from "./common";

// ── Enums ──────────────────────────────────────────────

export const TransactionTypeSchema = z.enum([
  "credit",
  "debit",
  "commission",
  "payout",
  "release",
  "refund",
]);

export type TransactionType = z.infer<typeof TransactionTypeSchema>;

// ── Domain schemas ─────────────────────────────────────

export const WalletSchema = z.object({
  id: UUIDSchema,
  tenant_id: UUIDSchema,
  available_balance: z.string(),
  pending_balance: z.string(),
  last_transaction_id: UUIDSchema.nullable().optional(),
  last_reconciliation_at: TimestampSchema.nullable().optional(),
});

export const TransactionSchema = z.object({
  id: UUIDSchema,
  wallet_id: UUIDSchema,
  order_id: UUIDSchema.nullable().optional(),
  amount: z.string(),
  running_balance: z.string(),
  platform_fee_base: z.string().optional(),
  platform_fee_rate: z.string().optional(),
  platform_fee_cap: z.string().optional(),
  platform_fee_amount: z.string().optional(),
  type: TransactionTypeSchema,
  signature: z.string(),
  created_at: TimestampSchema,
});

// ── Shared response schemas ────────────────────────────

export const PaginatedTransactionsResponseSchema = createPaginatedResponseSchema(TransactionSchema);

// ── Inferred types ─────────────────────────────────────

export type Wallet = z.infer<typeof WalletSchema>;
export type Transaction = z.infer<typeof TransactionSchema>;
export type PaginatedTransactionsResponse = z.infer<typeof PaginatedTransactionsResponseSchema>;
