import type { Product } from "./types";

/**
 * Formats the price range from a product's variants as a localized NGN string.
 * Returns "—" when there are no variants.
 * Returns a single price when min === max, or "min – max" otherwise.
 */
export function formatPrice(variants?: Product["variants"]): string {
  if (!variants || variants.length === 0) return "—";
  const prices = variants.map((v) => parseFloat(v.price));
  const min = Math.min(...prices);
  const max = Math.max(...prices);
  const fmt = (n: number) =>
    new Intl.NumberFormat("en-NG", {
      style: "currency",
      currency: "NGN",
      minimumFractionDigits: 0,
    }).format(n);
  return min === max ? fmt(min) : `${fmt(min)} – ${fmt(max)}`;
}

/**
 * Returns the total stock across all variants, or null if any variant
 * has an unknown (null / undefined) stock quantity.
 */
export function totalStock(variants?: Product["variants"]): number | null {
  if (!variants || variants.length === 0) return null;
  if (variants.some((v) => v.stock_qty === null || v.stock_qty === undefined)) return null;
  return variants.reduce((sum, v) => sum + (v.stock_qty ?? 0), 0);
}
