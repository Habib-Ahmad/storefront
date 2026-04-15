import * as z from "zod";
import { UUIDSchema } from "./domain";
import { FulfillmentStatusSchema, PaymentMethodSchema, PaymentStatusSchema } from "./orders";

export const PublicStorefrontSchema = z.object({
  name: z.string(),
  slug: z.string(),
  logo_url: z.string().nullable().optional(),
  contact_email: z.string().nullable().optional(),
  contact_phone: z.string().nullable().optional(),
  address: z.string().nullable().optional(),
  delivery: z.object({
    enabled: z.boolean(),
    ready: z.boolean(),
    unavailable_reason: z.string().nullable().optional(),
  }),
});

export const PublicStorefrontProductSchema = z.object({
  id: UUIDSchema,
  name: z.string(),
  description: z.string().nullable().optional(),
  category: z.string().nullable().optional(),
  image_url: z.string().nullable().optional(),
  price: z.string(),
  in_stock: z.boolean(),
});

export const PublicStorefrontVariantSchema = z.object({
  id: UUIDSchema,
  attributes: z.record(z.string(), z.unknown()),
  price: z.string(),
  in_stock: z.boolean(),
  is_default: z.boolean(),
});

export const PublicStorefrontImageSchema = z.object({
  id: UUIDSchema,
  url: z.string().url(),
  sort_order: z.number().int(),
  is_primary: z.boolean(),
});

export const PublicStorefrontResponseSchema = z.object({
  storefront: PublicStorefrontSchema,
  products: z.array(PublicStorefrontProductSchema),
});

export const PublicStorefrontProductDetailResponseSchema = z.object({
  storefront: PublicStorefrontSchema,
  product: PublicStorefrontProductSchema,
  variants: z.array(PublicStorefrontVariantSchema),
  images: z.array(PublicStorefrontImageSchema),
});

export const PublicStorefrontCheckoutOrderSchema = z.object({
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
});

export const CreatePublicStorefrontOrderItemSchema = z.object({
  variant_id: UUIDSchema,
  quantity: z.number().int().positive(),
});

export const PublicStorefrontDeliveryQuoteSelectionSchema = z.object({
  courier_id: z.string().min(1),
  service_code: z.string().min(1),
  service_type: z.string().nullable().optional(),
});

export const CreatePublicStorefrontDeliveryQuoteRequestSchema = z.object({
  customer_name: z.string().min(1),
  customer_phone: z.string().min(1),
  customer_email: z.string().email().nullable().optional(),
  shipping_address: z.string().min(1),
  delivery_instructions: z.string().nullable().optional(),
  items: z.array(CreatePublicStorefrontOrderItemSchema).min(1),
});

export const PublicStorefrontDeliveryQuoteOptionSchema = z.object({
  id: z.string(),
  courier_id: z.string(),
  courier_name: z.string(),
  service_code: z.string(),
  service_type: z.string(),
  amount: z.string(),
  currency: z.string(),
  pickup_eta: z.string().optional(),
  delivery_eta: z.string().optional(),
  tracking_label: z.string().optional(),
  tracking_level: z.number().int(),
  is_fastest: z.boolean(),
  is_cheapest: z.boolean(),
  provider_fields: z.unknown().optional(),
});

export const PublicStorefrontDeliveryQuoteDebugSchema = z.object({
  sender_address_code: z.number().int(),
  receiver_address_code: z.number().int(),
  category_id: z.number().int(),
  category_name: z.string(),
  package_box: z.string(),
  estimated_weight_kg: z.string(),
  assumptions: z.array(z.string()).optional(),
  raw_response: z.unknown().optional(),
});

export const PublicStorefrontDeliveryQuoteResponseSchema = z.object({
  storefront: PublicStorefrontSchema,
  options: z.array(PublicStorefrontDeliveryQuoteOptionSchema),
  debug: PublicStorefrontDeliveryQuoteDebugSchema.optional(),
});

export const CreatePublicStorefrontOrderRequestSchema = z
  .object({
    is_delivery: z.boolean(),
    checkout_id: UUIDSchema,
    customer_name: z.string().min(1).nullable().optional(),
    customer_phone: z.string().min(1),
    customer_email: z.string().email().nullable().optional(),
    shipping_address: z.string().nullable().optional(),
    delivery_option: PublicStorefrontDeliveryQuoteSelectionSchema.nullable().optional(),
    note: z.string().nullable().optional(),
    items: z.array(CreatePublicStorefrontOrderItemSchema).min(1),
  })
  .superRefine((value, ctx) => {
    if (!value.is_delivery) {
      return;
    }

    if (!value.shipping_address?.trim()) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "shipping_address is required for delivery orders",
        path: ["shipping_address"],
      });
    }

    if (!value.delivery_option) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "delivery_option is required for delivery orders",
        path: ["delivery_option"],
      });
    }
  });

export const PublicStorefrontCheckoutResponseSchema = z.object({
  storefront: PublicStorefrontSchema,
  order: PublicStorefrontCheckoutOrderSchema,
  authorization_url: z.string().url().optional(),
});

export type PublicStorefront = z.infer<typeof PublicStorefrontSchema>;
export type PublicStorefrontProduct = z.infer<typeof PublicStorefrontProductSchema>;
export type PublicStorefrontResponse = z.infer<typeof PublicStorefrontResponseSchema>;
export type PublicStorefrontVariant = z.infer<typeof PublicStorefrontVariantSchema>;
export type PublicStorefrontImage = z.infer<typeof PublicStorefrontImageSchema>;
export type PublicStorefrontProductDetailResponse = z.infer<
  typeof PublicStorefrontProductDetailResponseSchema
>;
export type PublicStorefrontCheckoutOrder = z.infer<typeof PublicStorefrontCheckoutOrderSchema>;
export type CreatePublicStorefrontOrderItem = z.infer<typeof CreatePublicStorefrontOrderItemSchema>;
export type CreatePublicStorefrontDeliveryQuoteRequest = z.infer<
  typeof CreatePublicStorefrontDeliveryQuoteRequestSchema
>;
export type PublicStorefrontDeliveryQuoteSelection = z.infer<
  typeof PublicStorefrontDeliveryQuoteSelectionSchema
>;
export type PublicStorefrontDeliveryQuoteOption = z.infer<
  typeof PublicStorefrontDeliveryQuoteOptionSchema
>;
export type PublicStorefrontDeliveryQuoteResponse = z.infer<
  typeof PublicStorefrontDeliveryQuoteResponseSchema
>;
export type CreatePublicStorefrontOrderRequest = z.infer<
  typeof CreatePublicStorefrontOrderRequestSchema
>;
export type PublicStorefrontCheckoutResponse = z.infer<
  typeof PublicStorefrontCheckoutResponseSchema
>;
