import Link from "next/link";
import { notFound } from "next/navigation";
import { CheckCircle2, CircleAlert, Clock3, PackageCheck } from "lucide-react";
import { PublicStorefrontActions } from "@/components/public-storefront-actions";
import { buttonVariants } from "@/components/ui/button";
import { ApiError, api } from "@/lib/api";
import { OrderRecoverySync } from "./order-recovery-sync";
import { OrderPaymentStatusSync } from "./order-payment-status-sync";
import { ResumePaymentButton } from "./resume-payment-button";

interface Props {
  params: Promise<{ slug: string }>;
  searchParams: Promise<{ reference?: string; trxref?: string }>;
}

const paymentHeadline: Record<string, string> = {
  pending: "Payment is still processing",
  paid: "Payment confirmed",
  failed: "Payment did not go through",
  refunded: "Payment refunded",
};

const paymentLabel: Record<string, string> = {
  pending: "Payment pending",
  paid: "Payment confirmed",
  failed: "Payment failed",
  refunded: "Payment refunded",
};

const orderProgressLabel: Record<string, string> = {
  processing: "In progress",
  completed: "Completed",
  shipped: "On the way",
  delivered: "Delivered",
  cancelled: "Cancelled",
};

const paymentDetail: Record<string, string> = {
  pending: "We have your order. Payment confirmation can take a short moment to come through.",
  paid: "Payment is complete and the store can continue with your order.",
  failed: "The order exists, but the payment attempt did not complete successfully.",
  refunded: "The payment for this order has been refunded.",
};

const orderProgressDetail: Record<string, string> = {
  processing: "The store has your order and is getting it ready.",
  completed: "The store has marked this order as complete.",
  shipped: "Your order has left the store and is on the way.",
  delivered: "Your order has been delivered.",
  cancelled: "This order was cancelled and will not move forward.",
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

export default async function OrderSummaryPage({ params, searchParams }: Props) {
  const { slug } = await params;
  const query = await searchParams;

  try {
    const tracking = await api.track(slug);
    const isDelivery = tracking.is_delivery === true;
    const returnedFromPayment = Boolean(query.reference || query.trxref);
    const canResumePayment =
      tracking.payment_status === "pending" &&
      tracking.fulfillment_status === "processing" &&
      !returnedFromPayment;
    const storefrontHref = tracking.storefront_slug ? `/${tracking.storefront_slug}` : null;
    const paymentCopy =
      tracking.payment_status === "pending"
        ? returnedFromPayment
          ? "You have returned from your payment attempt. We are checking with Paystack now and will update this page automatically."
          : "Your order has been created. If you have not finished payment yet, you can continue from this page at any time."
        : tracking.payment_status === "paid"
          ? isDelivery
            ? "Payment is complete. The store can now prepare and dispatch your order."
            : "Payment is complete. The store has your pickup order and will take it from here."
          : tracking.payment_status === "failed"
            ? "Your order is on record, but the payment attempt did not complete successfully."
            : "This order has already been refunded.";

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
            <PublicStorefrontActions slug={tracking.storefront_slug} />
          </div>

          <div className="mt-6 rounded-[2rem] border border-border/60 bg-card p-6 sm:p-8">
            <div className="flex flex-col gap-6 lg:flex-row lg:items-start lg:justify-between">
              <div>
                <div className="flex items-center gap-3 text-emerald-600">
                  <SummaryIcon paymentStatus={tracking.payment_status} />
                  <p className="text-sm font-medium tracking-[0.18em] uppercase">Order status</p>
                </div>
                <h1 className="mt-4 text-3xl font-semibold tracking-tight text-foreground sm:text-4xl">
                  {paymentHeadline[tracking.payment_status] ?? "Order received"}
                </h1>
                <p className="mt-3 max-w-2xl text-sm leading-6 text-muted-foreground sm:text-base">
                  {paymentCopy}
                </p>
              </div>

              {isDelivery ? (
                <div className="rounded-[1.5rem] border border-border/60 bg-background p-5 lg:min-w-56">
                  <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                    Tracking code
                  </p>
                  <p className="mt-2 text-lg font-semibold text-foreground">
                    {tracking.tracking_slug}
                  </p>
                </div>
              ) : null}
            </div>

            {storefrontHref || canResumePayment ? (
              <div className="mt-8 flex flex-wrap gap-3">
                {storefrontHref ? (
                  <Link
                    href={storefrontHref}
                    className={buttonVariants({ variant: "outline", size: "lg" })}
                  >
                    Continue shopping
                  </Link>
                ) : null}
                {canResumePayment ? <ResumePaymentButton slug={tracking.tracking_slug} /> : null}
              </div>
            ) : null}

            {isDelivery ? (
              <div className="mt-8 rounded-[1.5rem] border border-border/60 bg-background p-6">
                <div className="flex items-start gap-3">
                  <PackageCheck className="mt-0.5 h-5 w-5 shrink-0 text-foreground" />
                  <div>
                    <p className="text-base font-semibold text-foreground">Delivery updates</p>
                    <p className="mt-1 text-sm leading-6 text-muted-foreground">
                      We&apos;ll keep this page updated as the store prepares and dispatches your
                      order.
                    </p>
                  </div>
                </div>

                <div className="mt-6 grid gap-4 sm:grid-cols-2">
                  <div className="rounded-[1.25rem] border border-border/60 bg-card p-5">
                    <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                      Payment
                    </p>
                    <p className="mt-2 text-xl font-semibold text-foreground">
                      {paymentLabel[tracking.payment_status] ?? tracking.payment_status}
                    </p>
                    <p className="mt-2 text-sm leading-6 text-muted-foreground">
                      {paymentDetail[tracking.payment_status] ??
                        "We&apos;ll keep checking for changes."}
                    </p>
                  </div>

                  <div className="rounded-[1.25rem] border border-foreground/10 bg-foreground/3 p-5">
                    <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                      Order progress
                    </p>
                    <p className="mt-2 text-xl font-semibold text-foreground">
                      {orderProgressLabel[tracking.fulfillment_status] ??
                        tracking.fulfillment_status}
                    </p>
                    <p className="mt-2 text-sm leading-6 text-muted-foreground">
                      {orderProgressDetail[tracking.fulfillment_status] ??
                        "The store will update this page as your order moves forward."}
                    </p>
                  </div>
                </div>
              </div>
            ) : null}

            <OrderPaymentStatusSync
              slug={tracking.tracking_slug}
              paymentStatus={tracking.payment_status}
              reference={query.reference}
              trxref={query.trxref}
            />
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
