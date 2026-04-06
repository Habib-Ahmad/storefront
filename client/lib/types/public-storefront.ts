import { z } from "zod";
import { UUIDSchema } from "./domain";

export const PublicStorefrontSchema = z.object({
  name: z.string(),
  slug: z.string(),
  logo_url: z.string().nullable().optional(),
  contact_email: z.string().nullable().optional(),
  contact_phone: z.string().nullable().optional(),
  address: z.string().nullable().optional(),
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

export type PublicStorefront = z.infer<typeof PublicStorefrontSchema>;
export type PublicStorefrontProduct = z.infer<typeof PublicStorefrontProductSchema>;
export type PublicStorefrontResponse = z.infer<typeof PublicStorefrontResponseSchema>;
export type PublicStorefrontVariant = z.infer<typeof PublicStorefrontVariantSchema>;
export type PublicStorefrontImage = z.infer<typeof PublicStorefrontImageSchema>;
export type PublicStorefrontProductDetailResponse = z.infer<
  typeof PublicStorefrontProductDetailResponseSchema
>;
