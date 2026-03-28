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

export interface Tier {
  id: string;
  name: string;
  debt_ceiling: string;
  commission_rate: string;
  created_at: string;
  updated_at: string;
}

export interface User {
  id: string;
  tenant_id: string;
  email: string;
  first_name?: string | null;
  last_name?: string | null;
  phone?: string | null;
  role: UserRole;
  created_at: string;
  updated_at: string;
}

export interface Wallet {
  id: string;
  tenant_id: string;
  available_balance: string;
  pending_balance: string;
  last_transaction_id?: string | null;
  last_reconciliation_at?: string | null;
}

export interface Transaction {
  id: string;
  wallet_id: string;
  order_id?: string | null;
  amount: string;
  running_balance: string;
  type: TransactionType;
  signature: string;
  created_at: string;
}

export interface Order {
  id: string;
  tenant_id: string;
  tracking_slug: string;
  is_delivery: boolean;
  customer_name?: string | null;
  customer_phone?: string | null;
  customer_email?: string | null;
  shipping_address?: string | null;
  note?: string | null;
  total_amount: string;
  shipping_fee: string;
  payment_method: PaymentMethod;
  payment_status: PaymentStatus;
  fulfillment_status: FulfillmentStatus;
  created_at: string;
  updated_at: string;
}

export interface OrderItem {
  id: string;
  order_id: string;
  variant_id: string;
  quantity: number;
  price_at_sale: string;
  cost_price_at_sale?: string | null;
  product_name?: string | null;
  variant_label?: string | null;
}

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

export interface CreateOrderRequest {
  is_delivery: boolean;
  payment_method?: PaymentMethod;
  customer_name?: string | null;
  customer_phone?: string | null;
  customer_email?: string | null;
  shipping_address?: string | null;
  note?: string | null;
  shipping_fee?: number;
  total_amount?: number;
  items?: { variant_id: string; quantity: number }[];
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

export interface CreateOrderResponse extends Order {
  authorization_url?: string;
}

export interface TrackingResponse {
  tracking_slug: string;
  customer_name: string | null;
  payment_status: PaymentStatus;
  fulfillment_status: FulfillmentStatus;
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
