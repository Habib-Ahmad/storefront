"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useCallback } from "react";
import { api } from "@/lib/api";
import { PRODUCTS_KNOWN_STORAGE_KEY, removeStorageKeysByPrefix } from "@/lib/storage";
import { getSupabase } from "@/lib/supabase";
import { useSession } from "@/components/auth-provider";
import type { MeResponse } from "@/lib/types";

export function useMe() {
  const { session } = useSession();
  return useQuery<MeResponse>({
    queryKey: ["me"],
    queryFn: () => api.getMe(),
    enabled: !!session,
    staleTime: 5 * 60 * 1000,
  });
}

export function useSignOut() {
  const router = useRouter();
  const queryClient = useQueryClient();

  return useCallback(async () => {
    const supabase = getSupabase();
    if (supabase) await supabase.auth.signOut();
    api.setToken(null);
    removeStorageKeysByPrefix(PRODUCTS_KNOWN_STORAGE_KEY);
    queryClient.clear();
    router.replace("/login");
  }, [router, queryClient]);
}
