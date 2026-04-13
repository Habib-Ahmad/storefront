"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Clock3 } from "lucide-react";
import { getLatestPendingOrderForStorefront } from "@/lib/public-checkout-recovery";

type PendingOrderState = {
  trackingSlug: string;
  orderPath: string;
} | null;

export function PublicPendingOrderBanner({ storefrontSlug }: { storefrontSlug: string }) {
  const [pendingOrder, setPendingOrder] = useState<PendingOrderState>(null);

  useEffect(() => {
    const recovery = getLatestPendingOrderForStorefront(storefrontSlug);
    if (!recovery?.trackingSlug || !recovery.orderPath) {
      setPendingOrder(null);
      return;
    }

    setPendingOrder({
      trackingSlug: recovery.trackingSlug,
      orderPath: recovery.orderPath,
    });
  }, [storefrontSlug]);

  if (!pendingOrder) {
    return null;
  }

  return (
    <div className="rounded-[1.5rem] border border-amber-300/70 bg-amber-50 px-4 py-4 text-amber-950">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-start gap-3">
          <Clock3 className="mt-0.5 h-5 w-5 shrink-0" />
          <div>
            <p className="text-sm font-medium">You have an unfinished order.</p>
            <p className="mt-1 text-sm text-amber-900/80">
              Open your order to continue payment without starting over.
            </p>
          </div>
        </div>
        <Link
          href={pendingOrder.orderPath}
          className="inline-flex items-center justify-center rounded-full bg-amber-950 px-4 py-2 text-sm font-medium text-amber-50 transition-opacity hover:opacity-90"
        >
          Continue order
        </Link>
      </div>
    </div>
  );
}
