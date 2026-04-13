import { z } from "zod";
import { UUIDSchema, TimestampSchema } from "./domain";

// ── Enums ──────────────────────────────────────────────

export const PaymentMethodSchema = z.enum(["online", "cash", "transfer"]);
export const PaymentStatusSchema = z.enum(["pending", "paid", "failed", "refunded"]);
export const FulfillmentStatusSchema = z.enum([
  "processing",
  "completed",
  "shipped",
  "delivered",
  "cancelled",
]);

export type PaymentMethod = z.infer<typeof PaymentMethodSchema>;
export type PaymentStatus = z.infer<typeof PaymentStatusSchema>;
export type FulfillmentStatus = z.infer<typeof FulfillmentStatusSchema>;

// ── Domain schemas ─────────────────────────────────────

export const OrderSchema = z.object({
  id: UUIDSchema,
  tenant_id: UUIDSchema,
  tracking_slug: z.string(),
  is_delivery: z.boolean(),
  customer_name: z.string().nullable().optional(),
  customer_phone: z.string().nullable().optional(),
  customer_email: z.string().nullable().optional(),
  shipping_address: z.string().nullable().optional(),
  note: z.string().nullable().optional(),
  total_amount: z.string(),
  shipping_fee: z.string(),
  payment_method: PaymentMethodSchema,
  payment_status: PaymentStatusSchema,
  fulfillment_status: FulfillmentStatusSchema,
  created_at: TimestampSchema,
  updated_at: TimestampSchema,
});

export const OrderItemSchema = z.object({
  id: UUIDSchema,
  order_id: UUIDSchema,
  variant_id: UUIDSchema,
  quantity: z.number().int(),
  price_at_sale: z.string(),
  cost_price_at_sale: z.string().nullable().optional(),
  product_name: z.string().nullable().optional(),
  variant_label: z.string().nullable().optional(),
});

export const ShipmentSchema = z.object({
  id: UUIDSchema,
  order_id: UUIDSchema,
  tenant_id: UUIDSchema,
  status: z.enum(["queued", "picked_up", "in_transit", "delivered", "failed"]),
  carrier_ref: z.string().nullable().optional(),
  tracking_number: z.string().nullable().optional(),
  carrier_history: z.unknown().optional(),
  created_at: TimestampSchema,
  updated_at: TimestampSchema,
});

export const TrackingResponseSchema = z.object({
  tracking_slug: z.string(),
  customer_name: z.string().nullable(),
  payment_status: PaymentStatusSchema,
  fulfillment_status: FulfillmentStatusSchema,
});

export const ResumePaymentResponseSchema = z.object({
  authorization_url: z.string().url(),
});

export const CreateOrderResponseSchema = OrderSchema.extend({
  authorization_url: z.string().optional(),
});

// ── Request schemas ────────────────────────────────────

export const CreateOrderItemRequestSchema = z.object({
  variant_id: UUIDSchema,
  quantity: z.number().int().positive(),
});

export const CreateOrderRequestSchema = z.object({
  is_delivery: z.boolean(),
  payment_method: PaymentMethodSchema.optional(),
  customer_name: z.string().nullable().optional(),
  customer_phone: z.string().nullable().optional(),
  customer_email: z.string().nullable().optional(),
  shipping_address: z.string().nullable().optional(),
  note: z.string().nullable().optional(),
  shipping_fee: z.number().optional(),
  total_amount: z.number().optional(),
  items: z.array(CreateOrderItemRequestSchema).optional(),
});

// ── Shared response schemas ────────────────────────────

export const PaginatedOrdersResponseSchema = z.object({
  data: z.array(OrderSchema),
  total: z.number().int(),
  page: z.number().int(),
  per_page: z.number().int(),
});

// ── Inferred types ─────────────────────────────────────

export type Order = z.infer<typeof OrderSchema>;
export type OrderItem = z.infer<typeof OrderItemSchema>;
export type Shipment = z.infer<typeof ShipmentSchema>;
export type TrackingResponse = z.infer<typeof TrackingResponseSchema>;
export type ResumePaymentResponse = z.infer<typeof ResumePaymentResponseSchema>;
export type CreateOrderItemRequest = z.infer<typeof CreateOrderItemRequestSchema>;
export type CreateOrderRequest = z.infer<typeof CreateOrderRequestSchema>;
export type CreateOrderResponse = z.infer<typeof CreateOrderResponseSchema>;
export type PaginatedOrdersResponse = z.infer<typeof PaginatedOrdersResponseSchema>;
