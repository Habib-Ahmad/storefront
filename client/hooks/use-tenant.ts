"use client";

import { api } from "@/lib/api";
import { createMutationHook } from "@/lib/query-factory";
import type { OnboardRequest, UpdateStorefrontRequest } from "@/lib/types";

export const useOnboardTenant = createMutationHook(
  (data: OnboardRequest) => api.onboard(data),
  ["me"],
);

export const useUpdateStorefront = createMutationHook(
  (data: UpdateStorefrontRequest) => api.updateStorefront(data),
  ["me"],
);
