"use client";

import { useEffect } from "react";
import {
  clearPendingOrderByTrackingSlug,
  keepPendingOrderByTrackingSlug,
} from "@/lib/public-checkout-recovery";

export function OrderRecoverySync({
  trackingSlug,
  paymentStatus,
  fulfillmentStatus,
}: {
  trackingSlug: string;
  paymentStatus: string;
  fulfillmentStatus: string;
}) {
  useEffect(() => {
    if (paymentStatus === "pending" && fulfillmentStatus === "processing") {
      keepPendingOrderByTrackingSlug(trackingSlug);
      return;
    }

    clearPendingOrderByTrackingSlug(trackingSlug);
  }, [fulfillmentStatus, paymentStatus, trackingSlug]);

  return null;
}
