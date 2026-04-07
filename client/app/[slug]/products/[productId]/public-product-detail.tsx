"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  ArrowLeft,
  CheckCircle2,
  Mail,
  MapPin,
  MessageCircle,
  Phone,
  ShoppingCart,
  Zap,
} from "lucide-react";
import { PublicStorefrontActions } from "@/components/public-storefront-actions";
import { addStorefrontCartItem } from "@/lib/storefront-cart";
import type { PublicStorefrontProductDetailResponse } from "@/lib/types/public-storefront";
import { formatCurrency } from "../../storefront-formatters";

interface PublicProductDetailProps {
  detail: PublicStorefrontProductDetailResponse;
}

function formatVariantLabel(attributes: Record<string, unknown>) {
  const values = Object.values(attributes)
    .map((value) => (typeof value === "string" ? value : String(value)))
    .filter(Boolean);

  return values.length > 0 ? values.join(" / ") : "Default option";
}

function toWhatsAppHref(phone: string) {
  const normalized = phone.replace(/\D/g, "");
  return normalized ? `https://wa.me/${normalized}` : null;
}

export function PublicProductDetail({ detail }: PublicProductDetailProps) {
  const router = useRouter();
  const { storefront, product, variants, images } = detail;
  const defaultVariantIndex = variants.findIndex((variant) => variant.is_default);
  const [selectedVariantIndex, setSelectedVariantIndex] = useState(
    defaultVariantIndex >= 0 ? defaultVariantIndex : 0,
  );
  const primaryImageIndex = images.findIndex((image) => image.is_primary);
  const [selectedImageIndex, setSelectedImageIndex] = useState(
    primaryImageIndex >= 0 ? primaryImageIndex : 0,
  );

  const selectedVariant = variants[selectedVariantIndex] ?? variants[0] ?? null;
  const selectedImage = images[selectedImageIndex] ?? images[0] ?? null;
  const variantOptions = useMemo(
    () =>
      variants.map((variant) => ({
        id: variant.id,
        label: formatVariantLabel(variant.attributes),
        inStock: variant.in_stock,
      })),
    [variants],
  );
  const whatsappHref = storefront.contact_phone ? toWhatsAppHref(storefront.contact_phone) : null;
  const [showCartNotice, setShowCartNotice] = useState(false);
  const canCheckout = selectedVariant?.in_stock ?? product.in_stock;

  function buildCartItem() {
    if (!selectedVariant) {
      return null;
    }

    return {
      productId: product.id,
      productName: product.name,
      variantId: selectedVariant.id,
      variantLabel: formatVariantLabel(selectedVariant.attributes),
      unitPrice: selectedVariant.price,
      quantity: 1,
      imageUrl: selectedImage?.url ?? product.image_url ?? null,
    };
  }

  function handleAddToCart() {
    const item = buildCartItem();
    if (!item) {
      return;
    }

    addStorefrontCartItem(storefront.slug, item);
    setShowCartNotice(true);
  }

  function handleBuyNow() {
    const item = buildCartItem();
    if (!item) {
      return;
    }

    addStorefrontCartItem(storefront.slug, item);
    router.push(`/${storefront.slug}/checkout`);
  }

  return (
    <main className="min-h-screen bg-background text-foreground">
      <section className="mx-auto w-full max-w-7xl px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
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

        <div className="grid gap-10 py-8 lg:grid-cols-[minmax(0,1.1fr)_minmax(20rem,0.9fr)] lg:gap-14 lg:py-12">
          <section className="space-y-4">
            <div className="overflow-hidden rounded-[1.75rem] border border-border/60 bg-card">
              <div
                className="aspect-4/5 bg-secondary"
                style={
                  selectedImage
                    ? {
                        backgroundImage: `url(${selectedImage.url})`,
                        backgroundPosition: "center",
                        backgroundSize: "cover",
                      }
                    : undefined
                }
                aria-label={selectedImage ? `${product.name} image` : undefined}
              >
                {!selectedImage ? (
                  <div className="flex h-full items-center justify-center text-7xl font-semibold text-foreground/20">
                    {product.name.charAt(0).toUpperCase()}
                  </div>
                ) : null}
              </div>
            </div>

            {images.length > 1 ? (
              <div className="grid grid-cols-4 gap-3 sm:grid-cols-5">
                {images.map((image, index) => (
                  <button
                    key={image.id}
                    type="button"
                    onClick={() => setSelectedImageIndex(index)}
                    className={`overflow-hidden rounded-2xl border bg-card transition-colors ${
                      selectedImageIndex === index
                        ? "border-foreground/30"
                        : "border-border/60 hover:border-foreground/20"
                    }`}
                  >
                    <div
                      className="aspect-square bg-secondary"
                      style={{
                        backgroundImage: `url(${image.url})`,
                        backgroundPosition: "center",
                        backgroundSize: "cover",
                      }}
                    />
                  </button>
                ))}
              </div>
            ) : null}
          </section>

          <section className="space-y-8">
            <div className="space-y-4">
              <p className="text-xs tracking-[0.2em] text-muted-foreground uppercase">
                {product.category || "Product details"}
              </p>
              <div className="space-y-3">
                <h1 className="text-4xl font-semibold tracking-tight text-foreground sm:text-5xl">
                  {product.name}
                </h1>
                <div className="flex flex-wrap items-center gap-3">
                  <p className="text-3xl font-semibold tracking-tight text-foreground">
                    {formatCurrency(selectedVariant?.price ?? product.price)}
                  </p>
                  <span
                    className={`inline-flex items-center gap-1 rounded-full px-3 py-1 text-xs font-medium ${
                      (selectedVariant?.in_stock ?? product.in_stock)
                        ? "bg-foreground text-background"
                        : "bg-muted text-muted-foreground"
                    }`}
                  >
                    <span className="h-1.5 w-1.5 rounded-full bg-current" />
                    {(selectedVariant?.in_stock ?? product.in_stock) ? "In stock" : "Sold out"}
                  </span>
                </div>
              </div>
              <p className="max-w-2xl text-base leading-7 text-muted-foreground sm:text-lg">
                {product.description || "Ask the store for more details about this item."}
              </p>
            </div>

            {variantOptions.length > 0 ? (
              <div className="space-y-3 border-y border-border/60 py-6">
                <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                  Choose an option
                </p>
                <div className="flex flex-wrap gap-3">
                  {variantOptions.map((variant, index) => (
                    <button
                      key={variant.id}
                      type="button"
                      onClick={() => setSelectedVariantIndex(index)}
                      className={`rounded-full border px-4 py-2 text-sm transition-colors ${
                        selectedVariantIndex === index
                          ? "border-foreground bg-foreground text-background"
                          : "border-border/70 bg-card text-foreground hover:border-foreground/20"
                      }`}
                    >
                      {variant.label}
                    </button>
                  ))}
                </div>
              </div>
            ) : null}

            <div className="rounded-[1.75rem] border border-border/60 bg-card p-6">
              <p className="text-xs tracking-[0.2em] text-muted-foreground uppercase">
                Cart options
              </p>
              <h2 className="mt-3 text-2xl font-semibold tracking-tight text-foreground">
                Choose what to do next
              </h2>
              <p className="mt-2 text-sm leading-6 text-muted-foreground">
                Add this item to your cart or head straight to checkout.
              </p>

              <div className="mt-6 space-y-3">
                {canCheckout ? (
                  <>
                    <button
                      type="button"
                      onClick={handleAddToCart}
                      className="flex w-full items-center justify-center gap-2 rounded-full bg-foreground px-4 py-3 text-sm font-medium text-background transition-opacity hover:opacity-90"
                    >
                      <ShoppingCart className="h-4 w-4" />
                      Add to cart
                    </button>
                    <button
                      type="button"
                      onClick={handleBuyNow}
                      className="flex w-full items-center justify-center gap-2 rounded-full border border-border/70 px-4 py-3 text-sm font-medium text-foreground transition-colors hover:border-foreground/20"
                    >
                      <Zap className="h-4 w-4" />
                      Buy now
                    </button>
                  </>
                ) : (
                  <div className="flex items-center justify-center gap-2 rounded-full bg-muted px-4 py-3 text-sm font-medium text-muted-foreground">
                    <ShoppingCart className="h-4 w-4" />
                    Currently sold out
                  </div>
                )}

                {showCartNotice ? (
                  <div className="rounded-[1.5rem] border border-border/60 bg-background p-4">
                    <div className="flex items-start gap-2 text-sm text-foreground">
                      <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-emerald-600" />
                      <div>
                        <p className="font-medium">Added to your cart</p>
                        <p className="mt-1 text-muted-foreground">
                          You can keep browsing or head to checkout now.
                        </p>
                      </div>
                    </div>
                    <div className="mt-4 flex flex-col gap-3 sm:flex-row">
                      <Link
                        href={`/${storefront.slug}`}
                        className="inline-flex items-center justify-center rounded-full border border-border/70 px-4 py-3 text-sm font-medium text-foreground transition-colors hover:border-foreground/20"
                      >
                        Continue shopping
                      </Link>
                      <Link
                        href={`/${storefront.slug}/checkout`}
                        className="inline-flex items-center justify-center rounded-full bg-foreground px-4 py-3 text-sm font-medium text-background transition-opacity hover:opacity-90"
                      >
                        Go to checkout
                      </Link>
                    </div>
                  </div>
                ) : null}

                {storefront.contact_phone ? (
                  <a
                    href={`tel:${storefront.contact_phone}`}
                    className="flex items-center justify-center gap-2 rounded-full border border-border/70 px-4 py-3 text-sm font-medium text-foreground transition-colors hover:border-foreground/20"
                  >
                    <Phone className="h-4 w-4" />
                    Call the store
                  </a>
                ) : null}

                {whatsappHref ? (
                  <a
                    href={whatsappHref}
                    target="_blank"
                    rel="noreferrer"
                    className="flex items-center justify-center gap-2 rounded-full border border-border/70 px-4 py-3 text-sm font-medium text-foreground transition-colors hover:border-foreground/20"
                  >
                    <MessageCircle className="h-4 w-4" />
                    Message on WhatsApp
                  </a>
                ) : null}

                {storefront.contact_email ? (
                  <a
                    href={`mailto:${storefront.contact_email}`}
                    className="flex items-center justify-center gap-2 rounded-full border border-border/70 px-4 py-3 text-sm font-medium text-foreground transition-colors hover:border-foreground/20"
                  >
                    <Mail className="h-4 w-4" />
                    Email the store
                  </a>
                ) : null}
              </div>

              {storefront.address ? (
                <div className="mt-5 flex items-start gap-2 border-t border-border/60 pt-5 text-sm text-muted-foreground">
                  <MapPin className="mt-0.5 h-4 w-4 shrink-0" />
                  <span>{storefront.address}</span>
                </div>
              ) : null}
            </div>
          </section>
        </div>
      </section>
    </main>
  );
}
