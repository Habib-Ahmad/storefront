"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { createQueryHook } from "@/lib/query-factory";
import type { PaginationParams } from "@/lib/types";

export function useWallet() {
  return useQuery({
    queryKey: ["wallet"],
    queryFn: () => api.getWallet(),
  });
}

export const useTransactions = createQueryHook("wallet-transactions", (params: PaginationParams) =>
  api.getTransactions(params),
);
