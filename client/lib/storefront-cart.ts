"use client";

import { useEffect, useMemo, useState } from "react";

const CART_EVENT = "storefront-cart:change";

export interface StorefrontCartItem {
  productId: string;
  productName: string;
  variantId: string;
  variantLabel: string;
  unitPrice: string;
  quantity: number;
  imageUrl?: string | null;
}

function storageKey(slug: string) {
  return `storefront-cart:${slug}`;
}

function isStorefrontCartItem(value: unknown): value is StorefrontCartItem {
  if (!value || typeof value !== "object") {
    return false;
  }

  const candidate = value as Partial<StorefrontCartItem>;
  return (
    typeof candidate.productId === "string" &&
    typeof candidate.productName === "string" &&
    typeof candidate.variantId === "string" &&
    typeof candidate.variantLabel === "string" &&
    typeof candidate.unitPrice === "string" &&
    typeof candidate.quantity === "number" &&
    Number.isFinite(candidate.quantity) &&
    candidate.quantity > 0
  );
}

function normalizeItems(items: StorefrontCartItem[]) {
  const byVariant = new Map<string, StorefrontCartItem>();

  for (const item of items) {
    if (!isStorefrontCartItem(item)) {
      continue;
    }

    const existing = byVariant.get(item.variantId);
    if (existing) {
      byVariant.set(item.variantId, {
        ...existing,
        quantity: existing.quantity + item.quantity,
      });
      continue;
    }

    byVariant.set(item.variantId, {
      ...item,
      quantity: Math.max(1, Math.floor(item.quantity)),
    });
  }

  return Array.from(byVariant.values());
}

function dispatchCartChange(slug: string) {
  if (typeof window === "undefined") {
    return;
  }

  window.dispatchEvent(new CustomEvent(CART_EVENT, { detail: { slug } }));
}

export function getStorefrontCart(slug: string) {
  if (typeof window === "undefined") {
    return [] as StorefrontCartItem[];
  }

  const raw = window.localStorage.getItem(storageKey(slug));
  if (!raw) {
    return [];
  }

  try {
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) {
      return [];
    }

    return normalizeItems(parsed.filter(isStorefrontCartItem));
  } catch {
    return [];
  }
}

function persistStorefrontCart(slug: string, items: StorefrontCartItem[]) {
  if (typeof window === "undefined") {
    return;
  }

  const nextItems = normalizeItems(items);
  window.localStorage.setItem(storageKey(slug), JSON.stringify(nextItems));
  dispatchCartChange(slug);
}

export function addStorefrontCartItem(slug: string, item: StorefrontCartItem) {
  persistStorefrontCart(slug, [...getStorefrontCart(slug), item]);
}

export function replaceStorefrontCart(slug: string, items: StorefrontCartItem[]) {
  persistStorefrontCart(slug, items);
}

export function updateStorefrontCartItemQuantity(
  slug: string,
  variantId: string,
  quantity: number,
) {
  const nextItems = getStorefrontCart(slug)
    .map((item) =>
      item.variantId === variantId
        ? {
            ...item,
            quantity: Math.max(0, Math.floor(quantity)),
          }
        : item,
    )
    .filter((item) => item.quantity > 0);

  persistStorefrontCart(slug, nextItems);
}

export function removeStorefrontCartItem(slug: string, variantId: string) {
  persistStorefrontCart(
    slug,
    getStorefrontCart(slug).filter((item) => item.variantId !== variantId),
  );
}

export function clearStorefrontCart(slug: string) {
  persistStorefrontCart(slug, []);
}

export function useStorefrontCart(slug: string) {
  const [items, setItems] = useState<StorefrontCartItem[]>([]);

  useEffect(() => {
    const sync = () => setItems(getStorefrontCart(slug));
    sync();

    const onStorage = (event: StorageEvent) => {
      if (event.key === storageKey(slug)) {
        sync();
      }
    };

    const onCartChange = (event: Event) => {
      const detail = (event as CustomEvent<{ slug?: string }>).detail;
      if (!detail?.slug || detail.slug === slug) {
        sync();
      }
    };

    window.addEventListener("storage", onStorage);
    window.addEventListener(CART_EVENT, onCartChange as EventListener);

    return () => {
      window.removeEventListener("storage", onStorage);
      window.removeEventListener(CART_EVENT, onCartChange as EventListener);
    };
  }, [slug]);

  const itemCount = useMemo(() => items.reduce((sum, item) => sum + item.quantity, 0), [items]);
  const subtotal = useMemo(
    () => items.reduce((sum, item) => sum + Number(item.unitPrice) * item.quantity, 0),
    [items],
  );

  return {
    items,
    itemCount,
    subtotal,
    isEmpty: items.length === 0,
  };
}
