import { Mail, MapPin, Phone } from "lucide-react";
import { notFound } from "next/navigation";
import { PublicStorefrontError, getPublicStorefront } from "@/lib/public-storefront";

interface Props {
  params: Promise<{ slug: string }>;
}

function formatCurrency(amount: string) {
  return new Intl.NumberFormat("en-NG", {
    style: "currency",
    currency: "NGN",
    maximumFractionDigits: 0,
  }).format(Number(amount));
}

function getInitials(name: string) {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? "")
    .join("");
}

function getProductDescription(description?: string | null) {
  if (!description) {
    return "Ask the store for more details about this item.";
  }

  return description.length > 120 ? `${description.slice(0, 117)}...` : description;
}

export default async function StorefrontPage({ params }: Props) {
  const { slug } = await params;

  try {
    const { storefront, products } = await getPublicStorefront(slug);
    const hasContactDetails = Boolean(
      storefront.contact_phone || storefront.contact_email || storefront.address,
    );

    return (
      <main className="min-h-screen bg-background text-foreground">
        <section className="mx-auto w-full max-w-7xl px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
          <header className="flex items-center justify-between border-b border-border/60 pb-4">
            <div className="flex items-center gap-3">
              <div
                className="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl border border-border/70 bg-card text-sm font-semibold text-foreground"
                style={
                  storefront.logo_url
                    ? {
                        backgroundImage: `linear-gradient(rgba(255,255,255,0.06), rgba(255,255,255,0.06)), url(${storefront.logo_url})`,
                        backgroundPosition: "center",
                        backgroundSize: "cover",
                      }
                    : undefined
                }
                aria-label={`${storefront.name} logo`}
              >
                {storefront.logo_url ? null : getInitials(storefront.name)}
              </div>
              <div>
                <p className="text-sm font-medium text-foreground">{storefront.name}</p>
                <p className="text-xs text-muted-foreground">/{storefront.slug}</p>
              </div>
            </div>

            <div className="hidden text-xs tracking-[0.2em] text-muted-foreground uppercase sm:block">
              Shop online
            </div>
          </header>

          <section className="py-10 lg:py-14">
            <div className="space-y-8">
              <div className="space-y-4">
                <p className="text-xs tracking-[0.22em] text-muted-foreground uppercase">
                  Now available
                </p>
                <h1 className="max-w-3xl text-4xl font-semibold tracking-tight text-foreground sm:text-5xl lg:text-6xl">
                  {storefront.name}
                </h1>
                <p className="max-w-2xl text-base leading-7 text-muted-foreground sm:text-lg">
                  Browse what is available now, then get in touch with the store for delivery,
                  pickup, or payment details.
                </p>
              </div>

              <div className="grid gap-4 border-y border-border/60 py-5 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-end">
                <div>
                  <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                    Collection
                  </p>
                  <p className="mt-2 text-3xl font-semibold text-foreground">
                    {products.length} {products.length === 1 ? "item" : "items"}
                  </p>
                </div>
                <p className="text-sm leading-6 text-muted-foreground">/{storefront.slug}</p>
              </div>

              {hasContactDetails ? (
                <div className="flex flex-col gap-3 border-b border-border/60 pb-5 sm:flex-row sm:flex-wrap sm:gap-4">
                  {storefront.contact_phone ? (
                    <a
                      href={`tel:${storefront.contact_phone}`}
                      className="inline-flex items-center gap-2 text-sm text-foreground transition-opacity hover:opacity-70"
                    >
                      <Phone className="h-4 w-4 text-muted-foreground" />
                      <span>{storefront.contact_phone}</span>
                    </a>
                  ) : null}

                  {storefront.contact_email ? (
                    <a
                      href={`mailto:${storefront.contact_email}`}
                      className="inline-flex items-center gap-2 text-sm text-foreground transition-opacity hover:opacity-70"
                    >
                      <Mail className="h-4 w-4 text-muted-foreground" />
                      <span>{storefront.contact_email}</span>
                    </a>
                  ) : null}

                  {storefront.address ? (
                    <div className="inline-flex items-center gap-2 text-sm text-foreground">
                      <MapPin className="h-4 w-4 text-muted-foreground" />
                      <span>{storefront.address}</span>
                    </div>
                  ) : null}
                </div>
              ) : null}
            </div>
          </section>

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
                  This store is getting its collection ready. Check back soon for available
                  products.
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
        </section>
      </main>
    );
  } catch (error) {
    if (error instanceof PublicStorefrontError && error.status === 404) {
      notFound();
    }

    return (
      <main className="flex min-h-screen items-center justify-center px-4 py-12">
        <div className="max-w-xl rounded-[1.75rem] border border-border/60 bg-card p-8 text-center">
          <p className="text-xs font-medium tracking-[0.22em] text-muted-foreground uppercase">
            Storefront unavailable
          </p>
          <h1 className="mt-3 text-3xl font-semibold tracking-tight text-foreground">
            This storefront could not be loaded right now
          </h1>
          <p className="mt-3 text-sm leading-6 text-muted-foreground">
            Please try again in a moment.
          </p>
        </div>
      </main>
    );
  }
}
