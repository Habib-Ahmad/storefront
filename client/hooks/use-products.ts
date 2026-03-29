"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { createQueryHook, createMutationHook } from "@/lib/query-factory";
import type {
  PaginationParams,
  CreateProductRequest,
  UpdateProductRequest,
  CreateVariantRequest,
  AddImageRequest,
} from "@/lib/types";

// ── Queries (via factory) ──────────────────────────────

export const useProducts = createQueryHook("products", (params: PaginationParams) =>
  api.getProducts(params),
);

export const useProduct = createQueryHook("product", (id: string) => api.getProduct(id));

export const useVariants = createQueryHook("variants", (productId: string) =>
  api.getVariants(productId),
);

export const useImages = createQueryHook("images", (productId: string) => api.getImages(productId));

// ── Mutations (simple — via factory) ───────────────────

export const useCreateProduct = createMutationHook(
  (data: CreateProductRequest) => api.createProduct(data),
  ["products"],
);

export const useDeleteProduct = createMutationHook(
  (id: string) => api.deleteProduct(id),
  ["products"],
);

// ── Mutations (multi-arg — thin wrappers) ──────────────

export function useUpdateProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateProductRequest }) =>
      api.updateProduct(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["products"] });
      qc.invalidateQueries({ queryKey: ["product"] });
    },
  });
}

export function useCreateVariant() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ productId, data }: { productId: string; data: CreateVariantRequest }) =>
      api.createVariant(productId, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["product"] });
      qc.invalidateQueries({ queryKey: ["variants"] });
    },
  });
}

export function useUpdateVariant() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      productId,
      variantId,
      data,
    }: {
      productId: string;
      variantId: string;
      data: CreateVariantRequest;
    }) => api.updateVariant(productId, variantId, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["product"] });
      qc.invalidateQueries({ queryKey: ["variants"] });
    },
  });
}

export function useDeleteVariant() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ productId, variantId }: { productId: string; variantId: string }) =>
      api.deleteVariant(productId, variantId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["product"] });
      qc.invalidateQueries({ queryKey: ["variants"] });
    },
  });
}

export function useAddImage() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ productId, data }: { productId: string; data: AddImageRequest }) =>
      api.addImage(productId, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["product"] });
      qc.invalidateQueries({ queryKey: ["images"] });
    },
  });
}

export function useDeleteImage() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ productId, imageId }: { productId: string; imageId: string }) =>
      api.deleteImage(productId, imageId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["product"] });
      qc.invalidateQueries({ queryKey: ["images"] });
    },
  });
}
