"use client";

import { api } from "@/lib/api";
import { createMutationHook, createQueryHook } from "@/lib/query-factory";
import type { OnboardRequest } from "@/lib/contracts";

export const useTiers = createQueryHook("tiers", () => api.getTiers());

export const useTenant = createQueryHook("tenant", () => api.getTenant());

export const useOnboardTenant = createMutationHook(
  (data: OnboardRequest) => api.onboard(data),
  ["me", "tenant"],
);
