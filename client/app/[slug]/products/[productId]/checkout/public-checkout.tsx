"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { AlertCircle, ArrowLeft, CheckCircle2, LoaderCircle, MapPin, Phone } from "lucide-react";
import { PublicStorefrontActions } from "@/components/public-storefront-actions";
import { PublicStorefrontError, createPublicStorefrontOrder } from "@/lib/public-storefront";
import type {
  PublicStorefrontCheckoutResponse,
  PublicStorefrontProductDetailResponse,
} from "@/lib/types/public-storefront";
import { formatCurrency } from "../../../storefront-formatters";

interface PublicCheckoutProps {
  detail: PublicStorefrontProductDetailResponse;
  initialVariantId?: string | null;
}

function formatVariantLabel(attributes: Record<string, unknown>) {
  const values = Object.values(attributes)
    .map((value) => (typeof value === "string" ? value : String(value)))
    .filter(Boolean);

  return values.length > 0 ? values.join(" / ") : "Default option";
}

function formatExtendedCurrency(amount: number) {
  return formatCurrency(String(amount));
}

export function PublicCheckout({ detail, initialVariantId }: PublicCheckoutProps) {
  const { storefront, product, variants, images } = detail;
  const defaultVariant =
    variants.find((variant) => variant.id === initialVariantId) ??
    variants.find((variant) => variant.is_default) ??
    variants[0] ??
    null;

  const [selectedVariantId, setSelectedVariantId] = useState(defaultVariant?.id ?? "");
  const [quantity, setQuantity] = useState(1);
  const [customerPhone, setCustomerPhone] = useState("");
  const [shippingAddress, setShippingAddress] = useState("");
  const [note, setNote] = useState("");
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [result, setResult] = useState<PublicStorefrontCheckoutResponse | null>(null);

  const selectedVariant = useMemo(
    () => variants.find((variant) => variant.id === selectedVariantId) ?? variants[0] ?? null,
    [selectedVariantId, variants],
  );
  const primaryImage = images.find((image) => image.is_primary) ?? images[0] ?? null;
  const quantityValue = Number.isFinite(quantity) && quantity > 0 ? quantity : 1;
  const unitPrice = Number(selectedVariant?.price ?? product.price ?? 0);
  const subtotal = unitPrice * quantityValue;
  const canSubmit = !!selectedVariant && selectedVariant.in_stock;

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!selectedVariant) {
      setSubmitError("Choose an option before continuing.");
      return;
    }
    if (!selectedVariant.in_stock) {
      setSubmitError("This option is currently sold out.");
      return;
    }
    if (!customerPhone.trim()) {
      setSubmitError("Enter a phone number for delivery updates.");
      return;
    }
    if (!shippingAddress.trim()) {
      setSubmitError("Enter a delivery address.");
      return;
    }

    setIsSubmitting(true);
    setSubmitError(null);

    try {
      const response = await createPublicStorefrontOrder(storefront.slug, {
        is_delivery: true,
        customer_phone: customerPhone.trim(),
        shipping_address: shippingAddress.trim(),
        note: note.trim() || null,
        items: [{ variant_id: selectedVariant.id, quantity: quantityValue }],
      });
      setResult(response);
    } catch (error) {
      if (error instanceof PublicStorefrontError) {
        setSubmitError(error.message);
      } else {
        setSubmitError("Could not place the order right now.");
      }
    } finally {
      setIsSubmitting(false);
    }
  }

  if (result) {
    return (
      <main className="min-h-screen bg-background text-foreground">
        <section className="mx-auto w-full max-w-3xl px-4 py-6 sm:px-6 lg:px-8 lg:py-10">
          <div className="flex items-center justify-between gap-4 border-b border-border/60 pb-4">
            <Link
              href={`/${storefront.slug}`}
              className="inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground"
            >
              <ArrowLeft className="h-4 w-4" />
              Back to {storefront.name}
            </Link>
            <PublicStorefrontActions slug={storefront.slug} />
          </div>

          <div className="mt-6 rounded-[2rem] border border-border/60 bg-card p-6 sm:p-8">
            <div className="flex items-center gap-3 text-emerald-600">
              <CheckCircle2 className="h-6 w-6" />
              <p className="text-sm font-medium tracking-[0.18em] uppercase">Order submitted</p>
            </div>
            <h1 className="mt-4 text-3xl font-semibold tracking-tight text-foreground sm:text-4xl">
              {storefront.name} has your checkout request
            </h1>
            <p className="mt-3 max-w-2xl text-sm leading-6 text-muted-foreground sm:text-base">
              We have your delivery request. Payment comes next once the store confirms the order.
            </p>

            <div className="mt-8 grid gap-4 rounded-[1.5rem] border border-border/60 bg-background p-5 sm:grid-cols-2">
              <div>
                <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                  Tracking code
                </p>
                <p className="mt-2 text-lg font-semibold text-foreground">
                  {result.order.tracking_slug}
                </p>
              </div>
              <div>
                <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                  Order subtotal
                </p>
                <p className="mt-2 text-lg font-semibold text-foreground">
                  {formatCurrency(result.order.total_amount)}
                </p>
              </div>
            </div>

            <div className="mt-8 flex flex-col gap-3 sm:flex-row">
              <Link
                href={`/track/${result.order.tracking_slug}`}
                className="inline-flex items-center justify-center rounded-full bg-foreground px-5 py-3 text-sm font-medium text-background transition-opacity hover:opacity-90"
              >
                Track order
              </Link>
              <Link
                href={`/${storefront.slug}`}
                className="inline-flex items-center justify-center rounded-full border border-border/70 px-5 py-3 text-sm font-medium text-foreground transition-colors hover:border-foreground/20"
              >
                Continue shopping
              </Link>
            </div>
          </div>
        </section>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-background text-foreground">
      <section className="mx-auto w-full max-w-6xl px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
        <div className="border-b border-border/60 pb-4">
          <div className="flex items-center justify-between gap-4">
            <Link
              href={`/${storefront.slug}/products/${product.id}`}
              className="inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground"
            >
              <ArrowLeft className="h-4 w-4" />
              Back to product
            </Link>
            <PublicStorefrontActions slug={storefront.slug} />
          </div>
        </div>

        <div className="grid gap-8 py-8 lg:grid-cols-[minmax(0,0.9fr)_minmax(22rem,1.1fr)] lg:gap-12 lg:py-12">
          <aside className="space-y-4 rounded-[1.75rem] border border-border/60 bg-card p-5 sm:p-6">
            <div className="overflow-hidden rounded-[1.5rem] border border-border/60 bg-secondary">
              <div
                className="aspect-4/5"
                style={
                  primaryImage
                    ? {
                        backgroundImage: `url(${primaryImage.url})`,
                        backgroundPosition: "center",
                        backgroundSize: "cover",
                      }
                    : undefined
                }
              >
                {!primaryImage ? (
                  <div className="flex h-full items-center justify-center text-7xl font-semibold text-foreground/20">
                    {product.name.charAt(0).toUpperCase()}
                  </div>
                ) : null}
              </div>
            </div>

            <div>
              <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">Your item</p>
              <h1 className="mt-2 text-2xl font-semibold tracking-tight text-foreground">
                {product.name}
              </h1>
              <p className="mt-2 text-sm leading-6 text-muted-foreground">
                {product.description || "Order details will be confirmed directly with the store."}
              </p>
            </div>

            <div className="space-y-3 rounded-[1.5rem] border border-border/60 bg-background p-4">
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">Selected option</span>
                <span className="font-medium text-foreground">
                  {selectedVariant ? formatVariantLabel(selectedVariant.attributes) : "Unavailable"}
                </span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">Quantity</span>
                <span className="font-medium text-foreground">{quantityValue}</span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">Subtotal</span>
                <span className="font-medium text-foreground">
                  {formatExtendedCurrency(subtotal)}
                </span>
              </div>
              <div className="flex items-start justify-between gap-4 text-sm">
                <span className="text-muted-foreground">Delivery fee</span>
                <span className="text-right font-medium text-foreground">
                  Confirmed before payment
                </span>
              </div>
            </div>

            {storefront.address ? (
              <div className="flex items-start gap-2 text-sm text-muted-foreground">
                <MapPin className="mt-0.5 h-4 w-4 shrink-0" />
                <span>{storefront.address}</span>
              </div>
            ) : null}
          </aside>

          <section className="space-y-6 rounded-[1.75rem] border border-border/60 bg-card p-5 sm:p-6 lg:p-8">
            <div>
              <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                Direct checkout
              </p>
              <h2 className="mt-2 text-3xl font-semibold tracking-tight text-foreground">
                Delivery details
              </h2>
              <p className="mt-3 text-sm leading-6 text-muted-foreground">
                Keep this short: choose the quantity, add one phone number, and tell us where to
                deliver.
              </p>
            </div>

            <form className="space-y-6" onSubmit={handleSubmit}>
              <div className="space-y-3">
                <label className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                  Option
                </label>
                <div className="flex flex-wrap gap-3">
                  {variants.map((variant) => {
                    const active = variant.id === selectedVariantId;
                    return (
                      <button
                        key={variant.id}
                        type="button"
                        onClick={() => setSelectedVariantId(variant.id)}
                        className={`rounded-full border px-4 py-2 text-sm transition-colors ${
                          active
                            ? "border-foreground bg-foreground text-background"
                            : "border-border/70 bg-background text-foreground hover:border-foreground/20"
                        }`}
                      >
                        {formatVariantLabel(variant.attributes)}
                      </button>
                    );
                  })}
                </div>
              </div>

              <div className="grid gap-5 sm:grid-cols-[minmax(0,1fr)_10rem]">
                <label className="space-y-2">
                  <span className="text-sm font-medium text-foreground">Phone number</span>
                  <input
                    value={customerPhone}
                    onChange={(event) => setCustomerPhone(event.target.value)}
                    className="w-full rounded-2xl border border-border/70 bg-background px-4 py-3 text-sm transition-colors outline-none focus:border-foreground/30"
                    placeholder="08012345678"
                  />
                </label>

                <label className="space-y-2">
                  <span className="text-sm font-medium text-foreground">Quantity</span>
                  <input
                    value={quantityValue}
                    onChange={(event) => setQuantity(Math.max(1, Number(event.target.value) || 1))}
                    className="w-full rounded-2xl border border-border/70 bg-background px-4 py-3 text-sm transition-colors outline-none focus:border-foreground/30"
                    min={1}
                    step={1}
                    type="number"
                  />
                </label>
              </div>

              <label className="space-y-2">
                <span className="text-sm font-medium text-foreground">Delivery address</span>
                <textarea
                  value={shippingAddress}
                  onChange={(event) => setShippingAddress(event.target.value)}
                  className="min-h-28 w-full rounded-[1.5rem] border border-border/70 bg-background px-4 py-3 text-sm transition-colors outline-none focus:border-foreground/30"
                  placeholder="Street, area, city"
                />
              </label>

              <label className="space-y-2">
                <span className="text-sm font-medium text-foreground">Delivery note</span>
                <textarea
                  value={note}
                  onChange={(event) => setNote(event.target.value)}
                  className="min-h-28 w-full rounded-[1.5rem] border border-border/70 bg-background px-4 py-3 text-sm transition-colors outline-none focus:border-foreground/30"
                  placeholder="Optional gate code, landmark, or short note"
                />
              </label>

              <div className="rounded-[1.5rem] border border-border/60 bg-background p-4 text-sm text-muted-foreground">
                <div className="flex items-start gap-2">
                  <MapPin className="mt-0.5 h-4 w-4 shrink-0" />
                  <p>
                    We will confirm the delivery fee and payment before the order moves forward.
                  </p>
                </div>
              </div>

              {submitError ? (
                <div className="flex items-start gap-2 rounded-[1.5rem] border border-red-200 bg-red-50 p-4 text-sm text-red-700">
                  <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
                  <span>{submitError}</span>
                </div>
              ) : null}

              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                    Order subtotal
                  </p>
                  <p className="mt-1 text-2xl font-semibold tracking-tight text-foreground">
                    {formatExtendedCurrency(subtotal)}
                  </p>
                </div>

                <button
                  type="submit"
                  disabled={!canSubmit || isSubmitting}
                  className="inline-flex items-center justify-center gap-2 rounded-full bg-foreground px-5 py-3 text-sm font-medium text-background transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {isSubmitting ? (
                    <LoaderCircle className="h-4 w-4 animate-spin" />
                  ) : (
                    <Phone className="h-4 w-4" />
                  )}
                  Place order
                </button>
              </div>
            </form>
          </section>
        </div>
      </section>
    </main>
  );
}
