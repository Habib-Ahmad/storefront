// Types mirroring Go backend models.
// Decimals are strings (shopspring/decimal serializes as string).
// UUIDs and timestamps are strings.

// ── Enums ──────────────────────────────────────────────

export type TenantStatus = "active" | "suspended";
export type UserRole = "admin" | "staff" | "manager";
export type PaymentMethod = "online" | "cash" | "transfer";
export type PaymentStatus = "pending" | "paid" | "failed" | "refunded";
export type FulfillmentStatus = "processing" | "shipped" | "delivered" | "cancelled";
export type TransactionType = "credit" | "debit" | "commission" | "payout" | "release" | "refund";
export type ShipmentStatus = "queued" | "picked_up" | "in_transit" | "delivered" | "failed";

// ── Domain Models ──────────────────────────────────────

export interface Shipment {
  id: string;
  order_id: string;
  tenant_id: string;
  status: ShipmentStatus;
  carrier_ref?: string | null;
  tracking_number?: string | null;
  carrier_history?: unknown;
  created_at: string;
  updated_at: string;
}

// ── Request Types ──────────────────────────────────────

export interface UpdateUserRequest {
  first_name?: string | null;
  last_name?: string | null;
  phone?: string | null;
}

// ── Response Types ─────────────────────────────────────

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
}

export interface PaginationParams {
  page?: number;
  per_page?: number;
}

export interface TrackingResponse {
  tracking_slug: string;
  customer_name: string | null;
  payment_status: string;
  fulfillment_status: string;
}

export interface AnalyticsSummary {
  total_revenue: string;
  total_cost: string;
  total_profit: string;
  order_count: number;
  avg_order_value: string;
  by_payment_method: { method: PaymentMethod; revenue: string; count: number }[];
  top_products: { product_name: string; quantity_sold: number; revenue: string }[];
  period: { from: string; to: string };
}

export interface ApiErrorResponse {
  error?: string;
  errors?: Record<string, string>;
}
