import Link from "next/link";
import { notFound } from "next/navigation";
import { ArrowRight, CheckCircle2, CircleAlert, Clock3, PackageCheck } from "lucide-react";
import { PublicStorefrontActions } from "@/components/public-storefront-actions";
import { ApiError, api } from "@/lib/api";
import { OrderRecoverySync } from "./order-recovery-sync";
import { ResumePaymentButton } from "./resume-payment-button";

interface Props {
  params: Promise<{ slug: string }>;
}

const paymentHeadline: Record<string, string> = {
  pending: "Payment is still processing",
  paid: "Payment confirmed",
  failed: "Payment did not go through",
  refunded: "Payment refunded",
};

const paymentCopy: Record<string, string> = {
  pending:
    "Your order was created successfully. Payment confirmation can take a moment to reflect after the redirect.",
  paid: "Your order has been confirmed and the store can continue with fulfillment.",
  failed: "Your order is on record, but the payment attempt did not complete successfully.",
  refunded: "This order has already been refunded.",
};

const fulfillmentLabel: Record<string, string> = {
  processing: "Processing",
  completed: "Completed",
  shipped: "Shipped",
  delivered: "Delivered",
  cancelled: "Cancelled",
};

function SummaryIcon({ paymentStatus }: { paymentStatus: string }) {
  if (paymentStatus === "paid") {
    return <CheckCircle2 className="h-6 w-6" />;
  }
  if (paymentStatus === "failed") {
    return <CircleAlert className="h-6 w-6" />;
  }
  return <Clock3 className="h-6 w-6" />;
}

export default async function OrderSummaryPage({ params }: Props) {
  const { slug } = await params;

  try {
    const tracking = await api.track(slug);
    const canResumePayment =
      tracking.payment_status === "pending" && tracking.fulfillment_status === "processing";

    return (
      <main className="min-h-screen bg-background text-foreground">
        <section className="mx-auto w-full max-w-4xl px-4 py-6 sm:px-6 lg:px-8 lg:py-10">
          <OrderRecoverySync
            trackingSlug={tracking.tracking_slug}
            paymentStatus={tracking.payment_status}
            fulfillmentStatus={tracking.fulfillment_status}
          />
          <div className="flex items-center justify-between gap-4 border-b border-border/60 pb-4">
            <div className="text-sm text-muted-foreground">Order confirmation</div>
            <PublicStorefrontActions />
          </div>

          <div className="mt-6 rounded-[2rem] border border-border/60 bg-card p-6 sm:p-8">
            <div className="flex items-center gap-3 text-emerald-600">
              <SummaryIcon paymentStatus={tracking.payment_status} />
              <p className="text-sm font-medium tracking-[0.18em] uppercase">Order update</p>
            </div>
            <h1 className="mt-4 text-3xl font-semibold tracking-tight text-foreground sm:text-4xl">
              {paymentHeadline[tracking.payment_status] ?? "Order received"}
            </h1>
            <p className="mt-3 max-w-2xl text-sm leading-6 text-muted-foreground sm:text-base">
              {paymentCopy[tracking.payment_status] ?? "Your order is on record."}
            </p>

            <div className="mt-8 grid gap-4 sm:grid-cols-3">
              <div className="rounded-[1.5rem] border border-border/60 bg-background p-5">
                <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                  Tracking code
                </p>
                <p className="mt-2 text-lg font-semibold text-foreground">
                  {tracking.tracking_slug}
                </p>
              </div>
              <div className="rounded-[1.5rem] border border-border/60 bg-background p-5">
                <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                  Payment status
                </p>
                <p className="mt-2 text-lg font-semibold text-foreground">
                  {tracking.payment_status}
                </p>
              </div>
              <div className="rounded-[1.5rem] border border-border/60 bg-background p-5">
                <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                  Fulfillment
                </p>
                <p className="mt-2 text-lg font-semibold text-foreground">
                  {fulfillmentLabel[tracking.fulfillment_status] ?? tracking.fulfillment_status}
                </p>
              </div>
            </div>

            <div className="mt-8 rounded-[1.5rem] border border-border/60 bg-background p-5">
              <div className="flex items-start gap-3">
                <PackageCheck className="mt-0.5 h-5 w-5 shrink-0 text-foreground" />
                <div>
                  <p className="text-sm font-medium text-foreground">What happens next</p>
                  <p className="mt-2 text-sm leading-6 text-muted-foreground">
                    This page is the immediate post-checkout summary. Delivery-by-delivery tracking
                    comes next, once the logistics integration is wired in.
                  </p>
                </div>
              </div>
            </div>

            <div className="mt-8 flex flex-col gap-3 sm:flex-row">
              {canResumePayment ? <ResumePaymentButton slug={tracking.tracking_slug} /> : null}
              <Link
                href={`/track/${tracking.tracking_slug}`}
                className="inline-flex items-center justify-center gap-2 rounded-full bg-foreground px-5 py-3 text-sm font-medium text-background transition-opacity hover:opacity-90"
              >
                Track order
                <ArrowRight className="h-4 w-4" />
              </Link>
            </div>
          </div>
        </section>
      </main>
    );
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      notFound();
    }

    throw error;
  }
}
