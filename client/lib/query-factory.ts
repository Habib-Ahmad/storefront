"use client";

import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryOptions,
  type UseMutationOptions,
} from "@tanstack/react-query";

/**
 * Factory for creating a typed useQuery hook bound to a query key + fetcher.
 * Route changes only require updating api.ts — hooks stay untouched.
 */
export function createQueryHook<TData, TParams = void>(
  key: string,
  fetcher: (params: TParams) => Promise<TData>,
) {
  return (
    params: TParams,
    options?: Omit<UseQueryOptions<TData>, "queryKey" | "queryFn">,
  ) =>
    useQuery<TData>({
      queryKey: [key, params],
      queryFn: () => fetcher(params),
      ...options,
    });
}

/**
 * Factory for creating a typed useMutation hook that auto-invalidates
 * related query keys on success.
 */
export function createMutationHook<TData, TInput>(
  mutationFn: (data: TInput) => Promise<TData>,
  invalidateKeys: string[],
  options?: Omit<UseMutationOptions<TData, Error, TInput>, "mutationFn">,
) {
  return () => {
    const qc = useQueryClient();
    return useMutation<TData, Error, TInput>({
      mutationFn,
      onSuccess: (...args) => {
        invalidateKeys.forEach((k) =>
          qc.invalidateQueries({ queryKey: [k] }),
        );
        options?.onSuccess?.(...args);
      },
      ...options,
    });
  };
}
