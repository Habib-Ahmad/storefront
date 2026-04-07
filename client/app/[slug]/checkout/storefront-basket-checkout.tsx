"use client";

import { useState } from "react";
import Link from "next/link";
import {
  AlertCircle,
  ArrowLeft,
  CheckCircle2,
  LoaderCircle,
  Minus,
  Plus,
  ShoppingCart,
  Trash2,
} from "lucide-react";
import { PublicStorefrontError, createPublicStorefrontOrder } from "@/lib/public-storefront";
import {
  clearStorefrontCart,
  removeStorefrontCartItem,
  updateStorefrontCartItemQuantity,
  useStorefrontCart,
} from "@/lib/storefront-cart";
import type { PublicStorefrontCheckoutResponse } from "@/lib/types/public-storefront";
import { PublicStorefrontActions } from "@/components/public-storefront-actions";
import { formatCurrency } from "../storefront-formatters";

interface StorefrontBasketCheckoutProps {
  slug: string;
}

export function StorefrontBasketCheckout({ slug }: StorefrontBasketCheckoutProps) {
  const { items, isEmpty, subtotal } = useStorefrontCart(slug);
  const [customerPhone, setCustomerPhone] = useState("");
  const [shippingAddress, setShippingAddress] = useState("");
  const [note, setNote] = useState("");
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [result, setResult] = useState<PublicStorefrontCheckoutResponse | null>(null);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (items.length === 0) {
      setSubmitError("Your cart is empty.");
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
      const response = await createPublicStorefrontOrder(slug, {
        is_delivery: true,
        customer_phone: customerPhone.trim(),
        shipping_address: shippingAddress.trim(),
        note: note.trim() || null,
        items: items.map((item) => ({
          variant_id: item.variantId,
          quantity: item.quantity,
        })),
      });
      clearStorefrontCart(slug);
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
              href={`/${slug}`}
              className="inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground"
            >
              <ArrowLeft className="h-4 w-4" />
              Back to store
            </Link>
            <PublicStorefrontActions slug={slug} />
          </div>

          <div className="mt-6 rounded-[2rem] border border-border/60 bg-card p-6 sm:p-8">
            <div className="flex items-center gap-3 text-emerald-600">
              <CheckCircle2 className="h-6 w-6" />
              <p className="text-sm font-medium tracking-[0.18em] uppercase">Order received</p>
            </div>
            <h1 className="mt-4 text-3xl font-semibold tracking-tight text-foreground sm:text-4xl">
              Your order is in
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
                href={`/${slug}`}
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
        <div className="flex items-center justify-between gap-4 border-b border-border/60 pb-4">
          <Link
            href={`/${slug}`}
            className="inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground"
          >
            <ArrowLeft className="h-4 w-4" />
            Back to store
          </Link>
          <PublicStorefrontActions slug={slug} />
        </div>

        {isEmpty ? (
          <div className="mx-auto mt-12 max-w-2xl rounded-[2rem] border border-border/60 bg-card p-8 text-center">
            <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-secondary text-foreground">
              <ShoppingCart className="h-6 w-6" />
            </div>
            <h1 className="mt-5 text-3xl font-semibold tracking-tight text-foreground">
              Your cart is empty
            </h1>
            <p className="mt-3 text-sm leading-6 text-muted-foreground sm:text-base">
              Add a product first, then come back here to enter your delivery details.
            </p>
            <Link
              href={`/${slug}`}
              className="mt-6 inline-flex items-center justify-center rounded-full bg-foreground px-5 py-3 text-sm font-medium text-background transition-opacity hover:opacity-90"
            >
              Browse products
            </Link>
          </div>
        ) : (
          <div className="grid gap-8 py-8 lg:grid-cols-[minmax(0,0.95fr)_minmax(22rem,1.05fr)] lg:gap-12 lg:py-12">
            <aside className="space-y-4 rounded-[1.75rem] border border-border/60 bg-card p-5 sm:p-6">
              <div>
                <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                  Your cart
                </p>
                <h1 className="mt-2 text-3xl font-semibold tracking-tight text-foreground">
                  Review your items
                </h1>
                <p className="mt-2 text-sm leading-6 text-muted-foreground">
                  Check your items, add one phone number, and tell us where to deliver.
                </p>
              </div>

              <div className="space-y-3">
                {items.map((item) => (
                  <div
                    key={item.variantId}
                    className="rounded-[1.5rem] border border-border/60 bg-background p-4"
                  >
                    <div className="flex items-start justify-between gap-4">
                      <div className="min-w-0">
                        <p className="truncate text-sm font-medium text-foreground">
                          {item.productName}
                        </p>
                        <p className="mt-1 text-sm text-muted-foreground">{item.variantLabel}</p>
                      </div>
                      <button
                        type="button"
                        onClick={() => removeStorefrontCartItem(slug, item.variantId)}
                        className="rounded-full p-2 text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                        aria-label={`Remove ${item.productName}`}
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </div>

                    <div className="mt-4 flex items-center justify-between gap-4">
                      <div className="inline-flex items-center rounded-full border border-border/70">
                        <button
                          type="button"
                          onClick={() =>
                            updateStorefrontCartItemQuantity(
                              slug,
                              item.variantId,
                              item.quantity - 1,
                            )
                          }
                          className="p-2 text-muted-foreground transition-colors hover:text-foreground"
                          aria-label={`Reduce ${item.productName}`}
                        >
                          <Minus className="h-4 w-4" />
                        </button>
                        <span className="min-w-10 text-center text-sm font-medium text-foreground">
                          {item.quantity}
                        </span>
                        <button
                          type="button"
                          onClick={() =>
                            updateStorefrontCartItemQuantity(
                              slug,
                              item.variantId,
                              item.quantity + 1,
                            )
                          }
                          className="p-2 text-muted-foreground transition-colors hover:text-foreground"
                          aria-label={`Increase ${item.productName}`}
                        >
                          <Plus className="h-4 w-4" />
                        </button>
                      </div>
                      <p className="text-sm font-medium text-foreground">
                        {formatCurrency(String(Number(item.unitPrice) * item.quantity))}
                      </p>
                    </div>
                  </div>
                ))}
              </div>

              <div className="rounded-[1.5rem] border border-border/60 bg-background p-4">
                <div className="flex items-center justify-between text-sm text-muted-foreground">
                  <span>Subtotal</span>
                  <span className="font-medium text-foreground">
                    {formatCurrency(String(subtotal))}
                  </span>
                </div>
                <div className="mt-3 flex items-start justify-between gap-4 text-sm text-muted-foreground">
                  <span>Delivery fee</span>
                  <span className="text-right">Confirmed before payment</span>
                </div>
              </div>
            </aside>

            <section className="space-y-6 rounded-[1.75rem] border border-border/60 bg-card p-5 sm:p-6 lg:p-8">
              <div>
                <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                  Delivery details
                </p>
                <h2 className="mt-2 text-3xl font-semibold tracking-tight text-foreground">
                  Where should we deliver?
                </h2>
                <p className="mt-3 text-sm leading-6 text-muted-foreground">
                  Enter one phone number and one delivery address. We will confirm the order before
                  payment.
                </p>
              </div>

              <form className="space-y-6" onSubmit={handleSubmit}>
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
                    className="min-h-24 w-full rounded-[1.5rem] border border-border/70 bg-background px-4 py-3 text-sm transition-colors outline-none focus:border-foreground/30"
                    placeholder="Optional gate code, landmark, or short note"
                  />
                </label>

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
                      {formatCurrency(String(subtotal))}
                    </p>
                  </div>

                  <button
                    type="submit"
                    disabled={isSubmitting}
                    className="inline-flex items-center justify-center gap-2 rounded-full bg-foreground px-5 py-3 text-sm font-medium text-background transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    {isSubmitting ? <LoaderCircle className="h-4 w-4 animate-spin" /> : null}
                    Place order
                  </button>
                </div>
              </form>
            </section>
          </div>
        )}
      </section>
    </main>
  );
}
