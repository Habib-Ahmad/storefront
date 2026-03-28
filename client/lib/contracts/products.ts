import { z } from "zod";
import { UUIDSchema, TimestampSchema } from "./domain";
import { PaginationParamsSchema, createPaginatedResponseSchema } from "./common";

// ── Product domain schemas ─────────────────────────────

export const ProductVariantSchema = z.object({
  id: UUIDSchema,
  product_id: UUIDSchema,
  sku: z.string(),
  attributes: z.record(z.string(), z.unknown()),
  price: z.string(),
  cost_price: z.string().nullable().optional(),
  stock_qty: z.number().int().nullable().optional(),
  is_default: z.boolean(),
  created_at: TimestampSchema,
  updated_at: TimestampSchema,
});

export const ProductImageSchema = z.object({
  id: UUIDSchema,
  product_id: UUIDSchema,
  url: z.string().url(),
  sort_order: z.number().int(),
  is_primary: z.boolean(),
  created_at: TimestampSchema,
});

export const ProductSchema = z.object({
  id: UUIDSchema,
  tenant_id: UUIDSchema,
  name: z.string(),
  description: z.string().nullable().optional(),
  category: z.string().nullable().optional(),
  is_available: z.boolean(),
  created_at: TimestampSchema,
  updated_at: TimestampSchema,
  variants: z.array(ProductVariantSchema).optional(),
  images: z.array(ProductImageSchema).optional(),
});

export const ProductDetailResponseSchema = z.object({
  product: ProductSchema,
  variants: z.array(ProductVariantSchema),
  images: z.array(ProductImageSchema),
});

// ── Product request schemas ────────────────────────────

export const CreateVariantRequestSchema = z.object({
  sku: z.string(),
  attributes: z.record(z.string(), z.unknown()).optional(),
  price: z.string(),
  cost_price: z.string().nullable().optional(),
  stock_qty: z.number().int().nullable().optional(),
});

export const CreateProductRequestSchema = z.object({
  name: z.string(),
  description: z.string().nullable().optional(),
  category: z.string().nullable().optional(),
  is_available: z.boolean(),
  variants: z.array(CreateVariantRequestSchema).optional(),
});

export const UpdateProductRequestSchema = z.object({
  name: z.string(),
  description: z.string().nullable().optional(),
  category: z.string().nullable().optional(),
  is_available: z.boolean(),
});

export const AddImageRequestSchema = z.object({
  url: z.string().url(),
  sort_order: z.number().int(),
  is_primary: z.boolean(),
});

// ── Shared response schemas ────────────────────────────

export const PaginatedProductsResponseSchema = createPaginatedResponseSchema(ProductSchema);

// ── Inferred types ─────────────────────────────────────

export type ProductVariant = z.infer<typeof ProductVariantSchema>;
export type ProductImage = z.infer<typeof ProductImageSchema>;
export type Product = z.infer<typeof ProductSchema>;
export type ProductDetailResponse = z.infer<typeof ProductDetailResponseSchema>;
export type CreateVariantRequest = z.infer<typeof CreateVariantRequestSchema>;
export type CreateProductRequest = z.infer<typeof CreateProductRequestSchema>;
export type UpdateProductRequest = z.infer<typeof UpdateProductRequestSchema>;
export type AddImageRequest = z.infer<typeof AddImageRequestSchema>;
export type PaginationParams = z.infer<typeof PaginationParamsSchema>;
export type PaginatedProductsResponse = z.infer<typeof PaginatedProductsResponseSchema>;
