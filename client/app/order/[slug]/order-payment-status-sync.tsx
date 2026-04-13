"use client";

import { useEffect, useMemo, useState } from "react";
import { LoaderCircle } from "lucide-react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { ApiError, api } from "@/lib/api";

const maxAttempts = 5;
const attemptIntervalMs = 2500;

export function OrderPaymentStatusSync({
  slug,
  paymentStatus,
  reference,
  trxref,
}: {
  slug: string;
  paymentStatus: string;
  reference?: string;
  trxref?: string;
}) {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const [message, setMessage] = useState(
    "We're confirming your payment with Paystack now. This usually takes a few seconds.",
  );

  const cleanedHref = useMemo(() => {
    const params = new URLSearchParams(searchParams.toString());
    params.delete("reference");
    params.delete("trxref");
    const nextQuery = params.toString();
    return nextQuery ? `${pathname}?${nextQuery}` : pathname;
  }, [pathname, searchParams]);

  useEffect(() => {
    if (paymentStatus !== "pending") {
      return;
    }

    const paymentReference = reference?.trim() || trxref?.trim() || "";
    if (!paymentReference) {
      return;
    }

    let cancelled = false;
    let timer: number | undefined;

    const finish = () => {
      router.replace(cleanedHref, { scroll: false });
      router.refresh();
    };

    const confirm = async (attempt: number) => {
      try {
        const tracking = await api.confirmTrackedOrderPayment(slug, {
          reference: paymentReference,
          trxref,
        });
        if (cancelled) {
          return;
        }

        if (tracking.payment_status !== "pending") {
          finish();
          return;
        }
      } catch (error) {
        if (cancelled) {
          return;
        }
        if (error instanceof ApiError) {
          setMessage(
            error.message === "payment verification failed"
              ? "We're still waiting for Paystack to confirm your payment. This can take a few seconds."
              : error.message,
          );
        } else {
          setMessage("We're still checking for payment confirmation.");
        }
      }

      if (attempt + 1 >= maxAttempts) {
        finish();
        return;
      }

      timer = window.setTimeout(() => {
        void confirm(attempt + 1);
      }, attemptIntervalMs);
    };

    void confirm(0);

    return () => {
      cancelled = true;
      if (timer !== undefined) {
        window.clearTimeout(timer);
      }
    };
  }, [cleanedHref, paymentStatus, reference, router, slug, trxref]);

  if (paymentStatus !== "pending" || (!reference && !trxref)) {
    return null;
  }

  return (
    <div className="mt-8 rounded-[1.5rem] border border-border/60 bg-background p-5">
      <div className="flex items-start gap-3">
        <LoaderCircle className="mt-0.5 h-5 w-5 shrink-0 animate-spin text-foreground" />
        <div>
          <p className="text-sm font-medium text-foreground">Confirming payment</p>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">{message}</p>
        </div>
      </div>
    </div>
  );
}
