"use client";

import { useMutation, useQuery, useQueryClient, type UseQueryOptions } from "@tanstack/react-query";
import { useSession } from "@/components/auth-provider";
import { api } from "@/lib/api";
import { createMutationHook } from "@/lib/query-factory";
import type {
  CreateOrderRequest,
  DispatchShipmentOption,
  Order,
  OrderItem,
  PaginatedResponse,
  PaginationParams,
  Shipment,
} from "@/lib/types";

type AuthenticatedQueryOptions<TData> = Omit<UseQueryOptions<TData>, "queryKey" | "queryFn">;

function useAuthenticatedQuery<TData>(
  queryKey: ReadonlyArray<unknown>,
  queryFn: () => Promise<TData>,
  options?: AuthenticatedQueryOptions<TData>,
) {
  const { session, loading } = useSession();

  return useQuery<TData>({
    queryKey,
    queryFn,
    enabled: !loading && !!session && (options?.enabled ?? true),
    ...options,
  });
}

export function useOrders(
  params: PaginationParams,
  options?: AuthenticatedQueryOptions<PaginatedResponse<Order>>,
) {
  return useAuthenticatedQuery(["orders", params], () => api.getOrders(params), options);
}

export function useOrder(id: string, options?: AuthenticatedQueryOptions<Order>) {
  return useAuthenticatedQuery(["order", id], () => api.getOrder(id), options);
}

export function useOrderItems(orderId: string, options?: AuthenticatedQueryOptions<OrderItem[]>) {
  return useAuthenticatedQuery(["order-items", orderId], () => api.getOrderItems(orderId), options);
}

export function useOrderDispatchOptions(
  id: string,
  options?: AuthenticatedQueryOptions<DispatchShipmentOption[]>,
) {
  return useAuthenticatedQuery(
    ["order-dispatch-options", id],
    () => api.getOrderDispatchOptions(id),
    options,
  );
}

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
