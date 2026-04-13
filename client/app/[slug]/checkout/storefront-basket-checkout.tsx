"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import {
  AlertCircle,
  ArrowLeft,
  CheckCircle2,
  LoaderCircle,
  Minus,
  Plus,
  ShoppingCart,
  Store,
  Trash2,
  Truck,
} from "lucide-react";
import {
  PublicStorefrontError,
  createPublicStorefrontOrder,
  getPublicStorefrontDeliveryQuotes,
} from "@/lib/public-storefront";
import {
  emptyLogisticsAddress,
  formatLogisticsAddress,
  isLogisticsAddressComplete,
} from "@/lib/logistics-address";
import {
  basketRecoveryKey,
  getOrCreateCheckoutId,
  rememberPendingOrder,
} from "@/lib/public-checkout-recovery";
import {
  clearStorefrontCart,
  removeStorefrontCartItem,
  updateStorefrontCartItemQuantity,
  useStorefrontCart,
} from "@/lib/storefront-cart";
import type {
  PublicStorefrontCheckoutResponse,
  PublicStorefrontDeliveryQuoteOption,
  PublicStorefrontDeliveryQuoteResponse,
} from "@/lib/types/public-storefront";
import { PublicStorefrontActions } from "@/components/public-storefront-actions";
import { PublicPendingOrderBanner } from "@/components/public-pending-order-banner";
import { formatCurrency } from "../storefront-formatters";

interface StorefrontBasketCheckoutProps {
  slug: string;
}

type FulfillmentMode = "pickup" | "delivery";

function quoteBadge(option: PublicStorefrontDeliveryQuoteOption) {
  if (option.is_cheapest) {
    return "Lowest cost";
  }
  if (option.is_fastest) {
    return "Fastest";
  }
  return null;
}

export function StorefrontBasketCheckout({ slug }: StorefrontBasketCheckoutProps) {
  const recoveryKey = basketRecoveryKey(slug);
  const { items, isEmpty, subtotal } = useStorefrontCart(slug);
  const [fulfillmentMode, setFulfillmentMode] = useState<FulfillmentMode>("pickup");
  const [customerPhone, setCustomerPhone] = useState("");
  const [deliveryAddress, setDeliveryAddress] = useState(() => emptyLogisticsAddress());
  const [note, setNote] = useState("");
  const [quoteError, setQuoteError] = useState<string | null>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [isLoadingQuotes, setIsLoadingQuotes] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [quotes, setQuotes] = useState<PublicStorefrontDeliveryQuoteResponse | null>(null);
  const [selectedQuoteId, setSelectedQuoteId] = useState("");
  const [result, setResult] = useState<PublicStorefrontCheckoutResponse | null>(null);
  const [checkoutId] = useState(() => getOrCreateCheckoutId(recoveryKey, slug));

  const itemsSignature = useMemo(
    () => items.map((item) => `${item.variantId}:${item.quantity}`).join("|"),
    [items],
  );
  const formattedShippingAddress = useMemo(
    () => formatLogisticsAddress(deliveryAddress),
    [deliveryAddress],
  );
  const deliveryAddressReady = useMemo(
    () => isLogisticsAddressComplete(deliveryAddress),
    [deliveryAddress],
  );
  const selectedQuote = quotes?.options.find((option) => option.id === selectedQuoteId) ?? null;
  const shippingAmount = Number(selectedQuote?.amount ?? 0);
  const orderTotal = subtotal + shippingAmount;

  useEffect(() => {
    setQuotes(null);
    setSelectedQuoteId("");
    setQuoteError(null);
  }, [fulfillmentMode, customerPhone, formattedShippingAddress, note, itemsSignature]);

  useEffect(() => {
    if (fulfillmentMode !== "delivery") {
      setIsLoadingQuotes(false);
      return;
    }
    if (items.length === 0 || !customerPhone.trim() || !deliveryAddressReady) {
      setIsLoadingQuotes(false);
      return;
    }

    let cancelled = false;
    const timer = window.setTimeout(async () => {
      setIsLoadingQuotes(true);
      setQuoteError(null);

      try {
        const response = await getPublicStorefrontDeliveryQuotes(slug, {
          customer_name: "Guest customer",
          customer_phone: customerPhone.trim(),
          shipping_address: formattedShippingAddress,
          delivery_instructions: note.trim() || null,
          items: items.map((item) => ({
            variant_id: item.variantId,
            quantity: item.quantity,
          })),
        });

        if (!cancelled) {
          setQuotes(response);
        }
      } catch (error) {
        if (cancelled) {
          return;
        }
        if (error instanceof PublicStorefrontError) {
          setQuoteError(error.message);
        } else {
          setQuoteError("Could not load delivery options right now.");
        }
      } finally {
        if (!cancelled) {
          setIsLoadingQuotes(false);
        }
      }
    }, 350);

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [
    customerPhone,
    deliveryAddressReady,
    formattedShippingAddress,
    fulfillmentMode,
    items,
    note,
    slug,
  ]);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (items.length === 0) {
      setSubmitError("Your cart is empty.");
      return;
    }
    if (!customerPhone.trim()) {
      setSubmitError("Enter a phone number so the store can reach you.");
      return;
    }
    if (fulfillmentMode === "delivery") {
      if (!deliveryAddressReady) {
        setSubmitError("Enter a complete delivery address with street, city, state, and country.");
        return;
      }
      if (!selectedQuote) {
        setSubmitError("Choose one delivery option before continuing.");
        return;
      }
    }

    setIsSubmitting(true);
    setSubmitError(null);

    try {
      const response = await createPublicStorefrontOrder(slug, {
        is_delivery: fulfillmentMode === "delivery",
        checkout_id: checkoutId,
        customer_phone: customerPhone.trim(),
        shipping_address: fulfillmentMode === "delivery" ? formattedShippingAddress : null,
        delivery_option:
          fulfillmentMode === "delivery" && selectedQuote
            ? {
                courier_id: selectedQuote.courier_id,
                service_code: selectedQuote.service_code,
                service_type: selectedQuote.service_type,
              }
            : null,
        note: note.trim() || null,
        items: items.map((item) => ({
          variant_id: item.variantId,
          quantity: item.quantity,
        })),
      });
      clearStorefrontCart(slug);
      rememberPendingOrder(recoveryKey, slug, response.order.tracking_slug);
      if (response.authorization_url) {
        window.location.href = response.authorization_url;
        return;
      }
      if (response.order.payment_status === "pending") {
        window.location.href = `/order/${response.order.tracking_slug}`;
        return;
      }
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
              We have your order. If payment did not start automatically, use your order page for
              the latest payment and fulfillment status.
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
                  Order total
                </p>
                <p className="mt-2 text-lg font-semibold text-foreground">
                  {formatCurrency(result.order.total_amount)}
                </p>
              </div>
            </div>

            <div className="mt-8 flex flex-col gap-3 sm:flex-row">
              <Link
                href={`/order/${result.order.tracking_slug}`}
                className="inline-flex items-center justify-center rounded-full bg-foreground px-5 py-3 text-sm font-medium text-background transition-opacity hover:opacity-90"
              >
                View order
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

        <div className="mt-6">
          <PublicPendingOrderBanner storefrontSlug={slug} />
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
              Add a product first, then come back here to choose pickup or delivery.
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
                  Keep checkout short: pick delivery only if you need it, then continue to payment.
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

              <div className="space-y-3 rounded-[1.5rem] border border-border/60 bg-background p-4">
                <div className="flex items-center justify-between text-sm text-muted-foreground">
                  <span>Subtotal</span>
                  <span className="font-medium text-foreground">
                    {formatCurrency(String(subtotal))}
                  </span>
                </div>
                <div className="flex items-center justify-between text-sm text-muted-foreground">
                  <span>Delivery</span>
                  <span className="font-medium text-foreground">
                    {fulfillmentMode === "delivery" && selectedQuote
                      ? formatCurrency(selectedQuote.amount)
                      : fulfillmentMode === "delivery"
                        ? "Choose an option"
                        : "Pickup"}
                  </span>
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">Total before payment</span>
                  <span className="font-semibold text-foreground">
                    {formatCurrency(String(orderTotal))}
                  </span>
                </div>
              </div>
            </aside>

            <section className="space-y-6 rounded-[1.75rem] border border-border/60 bg-card p-5 sm:p-6 lg:p-8">
              <div>
                <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                  Checkout
                </p>
                <h2 className="mt-2 text-3xl font-semibold tracking-tight text-foreground">
                  Choose pickup or delivery
                </h2>
                <p className="mt-3 text-sm leading-6 text-muted-foreground">
                  Pickup stays frictionless. If you want delivery, complete the address step and
                  live courier options will appear automatically.
                </p>
              </div>

              <form className="space-y-6" onSubmit={handleSubmit}>
                <div className="grid gap-3 sm:grid-cols-2">
                  <button
                    type="button"
                    onClick={() => setFulfillmentMode("pickup")}
                    className={`rounded-[1.5rem] border p-4 text-left transition-colors ${
                      fulfillmentMode === "pickup"
                        ? "border-foreground bg-foreground text-background"
                        : "border-border/60 bg-background text-foreground"
                    }`}
                  >
                    <Store className="h-5 w-5" />
                    <p className="mt-3 text-sm font-semibold">Store pickup</p>
                    <p className="mt-1 text-sm opacity-80">No delivery fee, no address step.</p>
                  </button>
                  <button
                    type="button"
                    onClick={() => setFulfillmentMode("delivery")}
                    className={`rounded-[1.5rem] border p-4 text-left transition-colors ${
                      fulfillmentMode === "delivery"
                        ? "border-foreground bg-foreground text-background"
                        : "border-border/60 bg-background text-foreground"
                    }`}
                  >
                    <Truck className="h-5 w-5" />
                    <p className="mt-3 text-sm font-semibold">Delivery</p>
                    <p className="mt-1 text-sm opacity-80">
                      Compare courier options before you pay.
                    </p>
                  </button>
                </div>

                <label className="space-y-2">
                  <span className="text-sm font-medium text-foreground">Phone number</span>
                  <input
                    value={customerPhone}
                    onChange={(event) => setCustomerPhone(event.target.value)}
                    className="w-full rounded-2xl border border-border/70 bg-background px-4 py-3 text-sm transition-colors outline-none focus:border-foreground/30"
                    placeholder="08012345678"
                  />
                </label>

                {fulfillmentMode === "delivery" ? (
                  <>
                    <div className="space-y-3 rounded-[1.5rem] border border-border/60 bg-background p-4">
                      <div>
                        <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                          Step 2
                        </p>
                        <p className="mt-1 text-sm font-medium text-foreground">Delivery address</p>
                        <p className="mt-1 text-sm text-muted-foreground">
                          Use a clear street, city, state, and country so courier validation can
                          succeed.
                        </p>
                      </div>

                      <div className="grid gap-4 sm:grid-cols-2">
                        <label className="space-y-2 sm:col-span-2">
                          <span className="text-sm font-medium text-foreground">
                            Street address
                          </span>
                          <input
                            value={deliveryAddress.streetAddress}
                            onChange={(event) =>
                              setDeliveryAddress((current) => ({
                                ...current,
                                streetAddress: event.target.value,
                              }))
                            }
                            className="w-full rounded-2xl border border-border/70 bg-background px-4 py-3 text-sm transition-colors outline-none focus:border-foreground/30"
                            placeholder="16 Owerri Street, War College, Gwarinpa"
                          />
                        </label>

                        <label className="space-y-2">
                          <span className="text-sm font-medium text-foreground">City</span>
                          <input
                            value={deliveryAddress.city}
                            onChange={(event) =>
                              setDeliveryAddress((current) => ({
                                ...current,
                                city: event.target.value,
                              }))
                            }
                            className="w-full rounded-2xl border border-border/70 bg-background px-4 py-3 text-sm transition-colors outline-none focus:border-foreground/30"
                            placeholder="Abuja"
                          />
                        </label>

                        <label className="space-y-2">
                          <span className="text-sm font-medium text-foreground">State</span>
                          <input
                            value={deliveryAddress.state}
                            onChange={(event) =>
                              setDeliveryAddress((current) => ({
                                ...current,
                                state: event.target.value,
                              }))
                            }
                            className="w-full rounded-2xl border border-border/70 bg-background px-4 py-3 text-sm transition-colors outline-none focus:border-foreground/30"
                            placeholder="FCT"
                          />
                        </label>

                        <label className="space-y-2 sm:col-span-2">
                          <span className="text-sm font-medium text-foreground">Country</span>
                          <input
                            value={deliveryAddress.country}
                            onChange={(event) =>
                              setDeliveryAddress((current) => ({
                                ...current,
                                country: event.target.value,
                              }))
                            }
                            className="w-full rounded-2xl border border-border/70 bg-background px-4 py-3 text-sm transition-colors outline-none focus:border-foreground/30"
                            placeholder="Nigeria"
                          />
                        </label>
                      </div>
                    </div>

                    <label className="space-y-2">
                      <span className="text-sm font-medium text-foreground">Delivery note</span>
                      <textarea
                        value={note}
                        onChange={(event) => setNote(event.target.value)}
                        className="min-h-24 w-full rounded-[1.5rem] border border-border/70 bg-background px-4 py-3 text-sm transition-colors outline-none focus:border-foreground/30"
                        placeholder="Optional gate code, landmark, or short note"
                      />
                    </label>

                    <div className="space-y-4 rounded-[1.5rem] border border-border/60 bg-background p-4">
                      <div>
                        <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                          Step 3
                        </p>
                        <div>
                          <p className="text-sm font-medium text-foreground">Delivery options</p>
                          <p className="mt-1 text-sm text-muted-foreground">
                            Rates load automatically as soon as your delivery details are complete.
                          </p>
                        </div>
                      </div>

                      {isLoadingQuotes ? (
                        <div className="flex items-center gap-2 rounded-[1.25rem] border border-border/60 bg-card p-4 text-sm text-muted-foreground">
                          <LoaderCircle className="h-4 w-4 animate-spin" />
                          Loading live delivery options...
                        </div>
                      ) : null}

                      {!customerPhone.trim() || !deliveryAddressReady ? (
                        <div className="rounded-[1.25rem] border border-border/60 bg-card p-4 text-sm text-muted-foreground">
                          Enter your phone number and full delivery address to see available courier
                          options.
                        </div>
                      ) : null}

                      {quotes ? (
                        <div className="space-y-3">
                          {quotes.options.map((option) => {
                            const active = option.id === selectedQuoteId;
                            const badge = quoteBadge(option);
                            return (
                              <label
                                key={option.id}
                                className={`block cursor-pointer rounded-[1.25rem] border p-4 transition-colors ${
                                  active
                                    ? "border-foreground bg-secondary"
                                    : "border-border/60 bg-card"
                                }`}
                              >
                                <input
                                  type="radio"
                                  name="delivery-option"
                                  value={option.id}
                                  checked={active}
                                  onChange={() => setSelectedQuoteId(option.id)}
                                  className="sr-only"
                                />
                                <div className="flex items-start justify-between gap-4">
                                  <div>
                                    <p className="text-sm font-semibold text-foreground">
                                      {option.courier_name}
                                    </p>
                                    <p className="mt-1 text-sm text-muted-foreground">
                                      {option.service_type || option.service_code}
                                    </p>
                                    <p className="mt-1 text-sm text-muted-foreground">
                                      Delivery ETA: {option.delivery_eta || "Not provided"}
                                    </p>
                                  </div>
                                  <div className="text-right">
                                    {badge ? (
                                      <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                                        {badge}
                                      </p>
                                    ) : null}
                                    <p className="mt-1 text-lg font-semibold text-foreground">
                                      {formatCurrency(option.amount)}
                                    </p>
                                  </div>
                                </div>
                              </label>
                            );
                          })}
                        </div>
                      ) : null}

                      {quoteError ? (
                        <div className="flex items-start gap-2 rounded-[1.25rem] border border-red-200 bg-red-50 p-4 text-sm text-red-700">
                          <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
                          <span>{quoteError}</span>
                        </div>
                      ) : null}

                      {process.env.NODE_ENV !== "production" && quotes?.debug ? (
                        <details className="rounded-[1.25rem] border border-border/60 bg-card p-4 text-sm text-muted-foreground">
                          <summary className="cursor-pointer font-medium text-foreground">
                            Provider debug
                          </summary>
                          <pre className="mt-3 overflow-x-auto text-xs whitespace-pre-wrap">
                            {JSON.stringify(quotes.debug, null, 2)}
                          </pre>
                        </details>
                      ) : null}
                    </div>
                  </>
                ) : (
                  <div className="rounded-[1.5rem] border border-border/60 bg-background p-4 text-sm text-muted-foreground">
                    Pickup keeps checkout short. You can head straight to payment after this step.
                  </div>
                )}

                {submitError ? (
                  <div className="flex items-start gap-2 rounded-[1.5rem] border border-red-200 bg-red-50 p-4 text-sm text-red-700">
                    <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
                    <span>{submitError}</span>
                  </div>
                ) : null}

                <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                  <div>
                    <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                      Total before payment
                    </p>
                    <p className="mt-1 text-2xl font-semibold tracking-tight text-foreground">
                      {formatCurrency(String(orderTotal))}
                    </p>
                  </div>

                  <button
                    type="submit"
                    disabled={isSubmitting}
                    className="inline-flex items-center justify-center gap-2 rounded-full bg-foreground px-5 py-3 text-sm font-medium text-background transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    {isSubmitting ? (
                      <LoaderCircle className="h-4 w-4 animate-spin" />
                    ) : fulfillmentMode === "delivery" ? (
                      <Truck className="h-4 w-4" />
                    ) : (
                      <Store className="h-4 w-4" />
                    )}
                    Continue to payment
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
