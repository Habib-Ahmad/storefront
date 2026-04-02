"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { createMutationHook, createQueryHook } from "@/lib/query-factory";
import type { CreateOrderRequest, PaginationParams, Shipment } from "@/lib/types";

export const useOrders = createQueryHook("orders", (params: PaginationParams) =>
  api.getOrders(params),
);

export const useOrder = createQueryHook("order", (id: string) => api.getOrder(id));

export const useOrderItems = createQueryHook("order-items", (orderId: string) =>
  api.getOrderItems(orderId),
);

export const useCreateOrder = createMutationHook(
  (data: CreateOrderRequest) => api.createOrder(data),
  ["orders"],
);

export function useCancelOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.cancelOrder(id),
    onSuccess: (_data, id) => {
      qc.invalidateQueries({ queryKey: ["orders"] });
      qc.invalidateQueries({ queryKey: ["order", id] });
      qc.invalidateQueries({ queryKey: ["order-items", id] });
    },
  });
}

export function useDispatchOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: unknown }) => api.dispatchOrder(id, data),
    onSuccess: (_shipment: Shipment, { id }: { id: string; data: unknown }) => {
      qc.invalidateQueries({ queryKey: ["orders"] });
      qc.invalidateQueries({ queryKey: ["order", id] });
      qc.invalidateQueries({ queryKey: ["order-items", id] });
    },
  });
}
