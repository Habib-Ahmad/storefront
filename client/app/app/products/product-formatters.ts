import type { Product } from "@/lib/types";

export function formatPrice(variants?: Product["variants"]): string {
  if (!variants || variants.length === 0) return "—";

  const prices = variants.map((variant) => parseFloat(variant.price));
  const min = Math.min(...prices);
  const max = Math.max(...prices);
  const format = (amount: number) =>
    new Intl.NumberFormat("en-NG", {
      style: "currency",
      currency: "NGN",
      minimumFractionDigits: 0,
    }).format(amount);

  return min === max ? format(min) : `${format(min)} – ${format(max)}`;
}

export function totalStock(variants?: Product["variants"]): number | null {
  if (!variants || variants.length === 0) return null;
  if (variants.some((variant) => variant.stock_qty === null || variant.stock_qty === undefined)) {
    return null;
  }

  return variants.reduce((sum, variant) => sum + (variant.stock_qty ?? 0), 0);
}
