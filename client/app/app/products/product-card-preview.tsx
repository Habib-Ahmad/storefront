import { CaretDownIcon, ImageIcon } from "@phosphor-icons/react";
import { getProductDescription } from "@/app/[slug]/storefront-formatters";

type PreviewVariant = {
  price: string;
  stock_qty?: string | number | null;
};

interface ProductCardPreviewProps {
  name: string;
  description?: string | null;
  category?: string | null;
  imageURL?: string | null;
  variants: PreviewVariant[];
  available?: boolean;
  summary?: string;
}

function formatCurrency(amount: string) {
  const numericAmount = Number(amount);
  if (!Number.isFinite(numericAmount) || numericAmount <= 0) {
    return "Add a price";
  }

  return new Intl.NumberFormat("en-NG", {
    style: "currency",
    currency: "NGN",
    maximumFractionDigits: 0,
  }).format(numericAmount);
}

function getPriceLabel(variants: PreviewVariant[]) {
  const numericPrices = variants
    .map((variant) => Number(variant.price))
    .filter((value) => Number.isFinite(value) && value > 0);

  if (numericPrices.length === 0) {
    return "Add a price";
  }

  const lowest = Math.min(...numericPrices);
  const highest = Math.max(...numericPrices);

  if (lowest === highest) {
    return formatCurrency(String(lowest));
  }

  return `${formatCurrency(String(lowest))} - ${formatCurrency(String(highest))}`;
}

function isInStock(variants: PreviewVariant[]) {
  if (variants.length === 0) {
    return false;
  }

  return variants.some((variant) => {
    if (variant.stock_qty === "") {
      return true;
    }
    if (variant.stock_qty === null || variant.stock_qty === undefined) {
      return true;
    }

    return Number(variant.stock_qty) > 0;
  });
}

function getDisplayName(name: string) {
  const trimmed = name.trim();
  return trimmed || "Untitled product";
}

export function ProductCardPreview({
  name,
  description,
  category,
  imageURL,
  variants,
  available = true,
  summary = "See how this product card will read before customers do.",
}: ProductCardPreviewProps) {
  const displayName = getDisplayName(name);
  const stockState = available && isInStock(variants);

  return (
    <details className="card-3d overflow-hidden rounded-2xl" open>
      <summary className="flex cursor-pointer list-none items-center justify-between gap-3 p-5">
        <div>
          <h2 className="text-base font-semibold">Live storefront preview</h2>
          <p className="mt-1 text-sm text-muted-foreground">{summary}</p>
        </div>
        <CaretDownIcon className="details-open:rotate-180 size-4 text-muted-foreground transition-transform" />
      </summary>

      <div className="px-5 pt-1 pb-5">
        <div className="mx-auto max-w-sm overflow-hidden rounded-[1.5rem] border border-border/60 bg-card">
          <article>
            <div
              className="relative aspect-4/5 bg-secondary"
              aria-label={imageURL ? `${displayName} preview` : undefined}
            >
              {imageURL ? (
                <img
                  src={imageURL}
                  alt={`${displayName} preview`}
                  className="absolute inset-0 h-full w-full object-contain"
                />
              ) : null}
              {!imageURL ? (
                <div className="absolute inset-0 flex flex-col items-center justify-center gap-2 bg-[linear-gradient(180deg,rgba(0,0,0,0.02),rgba(0,0,0,0.08))] px-6 text-center text-muted-foreground">
                  <ImageIcon className="size-10 opacity-50" />
                  <p className="text-sm">Primary product image will appear here</p>
                </div>
              ) : null}
            </div>

            <div className="space-y-4 p-5">
              <div className="flex items-start justify-between gap-4">
                <div className="space-y-2">
                  <p className="text-[11px] tracking-[0.18em] text-muted-foreground uppercase">
                    {category?.trim() || "General"}
                  </p>
                  <h3 className="text-lg font-semibold tracking-tight text-foreground">
                    {displayName}
                  </h3>
                </div>
                <span
                  className={`inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-[11px] font-medium ${
                    stockState ? "bg-foreground text-background" : "bg-muted text-muted-foreground"
                  }`}
                >
                  <span className="h-1.5 w-1.5 rounded-full bg-current" />
                  {stockState ? "In stock" : "Sold out"}
                </span>
              </div>

              <p className="min-h-12 text-sm leading-6 text-muted-foreground">
                {getProductDescription(description?.trim())}
              </p>

              <div className="flex items-end justify-between border-t border-border/60 pt-4">
                <div>
                  <p className="text-[11px] tracking-[0.18em] text-muted-foreground uppercase">
                    Price
                  </p>
                  <p className="mt-1 text-2xl font-semibold tracking-tight text-foreground">
                    {getPriceLabel(variants)}
                  </p>
                </div>
                <p className="max-w-34 text-right text-[11px] leading-5 text-muted-foreground uppercase">
                  View details
                </p>
              </div>
            </div>
          </article>
        </div>
      </div>
    </details>
  );
}
