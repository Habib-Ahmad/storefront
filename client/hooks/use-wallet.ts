"use client";

import { useQuery, type UseQueryOptions } from "@tanstack/react-query";
import { useSession } from "@/components/auth-provider";
import { api } from "@/lib/api";
import type { PaginatedResponse, PaginationParams, Transaction, Wallet } from "@/lib/types";

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

export function useWallet(options?: AuthenticatedQueryOptions<Wallet>) {
  return useAuthenticatedQuery(["wallet"], () => api.getWallet(), options);
}

export function useTransactions(
  params: PaginationParams,
  options?: AuthenticatedQueryOptions<PaginatedResponse<Transaction>>,
) {
  return useAuthenticatedQuery(
    ["wallet-transactions", params],
    () => api.getTransactions(params),
    options,
  );
}
