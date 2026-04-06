import type { PublicStorefrontProduct } from "@/lib/types/public-storefront";
import { formatCurrency, getProductDescription } from "./storefront-formatters";

interface StorefrontCatalogProps {
  products: PublicStorefrontProduct[];
}

export function StorefrontCatalog({ products }: StorefrontCatalogProps) {
  return (
    <section className="border-t border-border/60 pt-8 lg:pt-10">
      <div className="flex flex-col gap-3 pb-6 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-xs tracking-[0.2em] text-muted-foreground uppercase">Catalog</p>
          <h2 className="mt-2 text-2xl font-semibold tracking-tight text-foreground sm:text-3xl">
            Available now
          </h2>
        </div>
        <p className="max-w-md text-sm leading-6 text-muted-foreground">
          Browse the latest available items and contact the store when you are ready to order.
        </p>
      </div>

      {products.length === 0 ? (
        <div className="rounded-[1.75rem] border border-dashed border-border/80 bg-card px-6 py-12 text-center sm:px-10">
          <p className="text-lg font-medium text-foreground">New items are coming soon</p>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">
            This store is getting its collection ready. Check back soon for available products.
          </p>
        </div>
      ) : (
        <div className="grid gap-x-5 gap-y-8 sm:grid-cols-2 xl:grid-cols-3">
          {products.map((product) => (
            <article
              key={product.id}
              className="group overflow-hidden rounded-[1.5rem] border border-border/60 bg-card transition-colors hover:border-foreground/20"
            >
              <div
                className="relative aspect-4/5 bg-secondary"
                style={
                  product.image_url
                    ? {
                        backgroundImage: `url(${product.image_url})`,
                        backgroundPosition: "center",
                        backgroundSize: "cover",
                      }
                    : undefined
                }
                aria-label={product.image_url ? `${product.name} preview` : undefined}
              >
                {!product.image_url ? (
                  <div className="absolute inset-0 flex items-center justify-center bg-[linear-gradient(180deg,rgba(0,0,0,0.02),rgba(0,0,0,0.08))] text-6xl font-semibold tracking-tight text-foreground/35">
                    {product.name.charAt(0).toUpperCase()}
                  </div>
                ) : null}
              </div>

              <div className="space-y-4 p-5">
                <div className="flex items-start justify-between gap-4">
                  <div className="space-y-2">
                    <p className="text-[11px] tracking-[0.18em] text-muted-foreground uppercase">
                      {product.category || "General"}
                    </p>
                    <h3 className="text-lg font-semibold tracking-tight text-foreground">
                      {product.name}
                    </h3>
                  </div>
                  <span
                    className={`inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-[11px] font-medium ${
                      product.in_stock
                        ? "bg-foreground text-background"
                        : "bg-muted text-muted-foreground"
                    }`}
                  >
                    <span className="h-1.5 w-1.5 rounded-full bg-current" />
                    {product.in_stock ? "In stock" : "Sold out"}
                  </span>
                </div>

                <p className="min-h-12 text-sm leading-6 text-muted-foreground">
                  {getProductDescription(product.description)}
                </p>

                <div className="flex items-end justify-between border-t border-border/60 pt-4">
                  <div>
                    <p className="text-[11px] tracking-[0.18em] text-muted-foreground uppercase">
                      Price
                    </p>
                    <p className="mt-1 text-2xl font-semibold tracking-tight text-foreground">
                      {formatCurrency(product.price)}
                    </p>
                  </div>
                  <p className="max-w-34 text-right text-[11px] leading-5 text-muted-foreground uppercase">
                    Contact to order
                  </p>
                </div>
              </div>
            </article>
          ))}
        </div>
      )}
    </section>
  );
}
