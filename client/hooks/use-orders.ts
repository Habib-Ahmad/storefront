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

export const useOrderDispatchOptions = createQueryHook("order-dispatch-options", (id: string) =>
  api.getOrderDispatchOptions(id),
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

export function useResumeOrderPayment() {
  return useMutation({
    mutationFn: (id: string) => api.resumeOrderPayment(id),
  });
}

export function useResumeTrackedOrderPayment() {
  return useMutation({
    mutationFn: (slug: string) => api.resumeTrackedOrderPayment(slug),
  });
}

export function useDispatchOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: string;
      data: { courier_id: string; service_code: string; service_type?: string };
    }) => api.dispatchOrder(id, data),
    onSuccess: (
      _shipment: Shipment,
      {
        id,
      }: { id: string; data: { courier_id: string; service_code: string; service_type?: string } },
    ) => {
      qc.invalidateQueries({ queryKey: ["orders"] });
      qc.invalidateQueries({ queryKey: ["order", id] });
      qc.invalidateQueries({ queryKey: ["order-items", id] });
      qc.invalidateQueries({ queryKey: ["order-dispatch-options", id] });
    },
  });
}
