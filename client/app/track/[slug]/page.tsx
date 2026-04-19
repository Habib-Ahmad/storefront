import Link from "next/link";
import { notFound } from "next/navigation";
import { ArrowLeft, CircleDot, PackageCheck, Truck } from "lucide-react";
import { PublicStorefrontActions } from "@/components/public-storefront-actions";
import { ApiError, api } from "@/lib/api";

interface Props {
  params: Promise<{ slug: string }>;
}

const paymentStatusLabel: Record<string, string> = {
  pending: "Payment pending",
  paid: "Payment confirmed",
  failed: "Payment failed",
  refunded: "Payment refunded",
};

const fulfillmentStatusLabel: Record<string, string> = {
  processing: "Preparing order",
  completed: "Order completed",
  shipped: "Shipment in progress",
  delivered: "Delivered",
  cancelled: "Order cancelled",
};

export default async function TrackPage({ params }: Props) {
  const { slug } = await params;

  try {
    const tracking = await api.track(slug);

    return (
      <main className="min-h-screen bg-background text-foreground">
        <section className="mx-auto w-full max-w-4xl px-4 py-6 sm:px-6 lg:px-8 lg:py-10">
          <div className="flex items-center justify-between gap-4 border-b border-border/60 pb-4">
            <Link
              href={`/order/${slug}`}
              className="inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground"
            >
              <ArrowLeft className="h-4 w-4" />
              Back to order
            </Link>
            <PublicStorefrontActions slug={tracking.storefront_slug} />
          </div>

          <div className="mt-6 rounded-[2rem] border border-border/60 bg-card p-6 sm:p-8">
            <div className="flex items-center gap-3 text-amber-600">
              <Truck className="h-6 w-6" />
              <p className="text-sm font-medium tracking-[0.18em] uppercase">Tracking soon</p>
            </div>
            <h1 className="mt-4 text-3xl font-semibold tracking-tight text-foreground sm:text-4xl">
              Live delivery tracking is not enabled yet
            </h1>
            <p className="mt-3 max-w-2xl text-sm leading-6 text-muted-foreground sm:text-base">
              This page will show courier movement once logistics integration is live. For now, you
              can still confirm your payment and order progress here.
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
                <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">Payment</p>
                <p className="mt-2 text-lg font-semibold text-foreground">
                  {paymentStatusLabel[tracking.payment_status] ?? tracking.payment_status}
                </p>
              </div>
              <div className="rounded-[1.5rem] border border-border/60 bg-background p-5">
                <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                  Fulfillment
                </p>
                <p className="mt-2 text-lg font-semibold text-foreground">
                  {fulfillmentStatusLabel[tracking.fulfillment_status] ??
                    tracking.fulfillment_status}
                </p>
              </div>
            </div>

            <div className="mt-8 rounded-[1.5rem] border border-border/60 bg-background p-5 text-sm text-muted-foreground">
              <div className="flex items-start gap-3">
                <CircleDot className="mt-0.5 h-4 w-4 shrink-0 text-foreground" />
                <p>
                  Keep this tracking code handy. As soon as shipment events are available, this page
                  will become the delivery timeline.
                </p>
              </div>
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
