import { Mail, MapPin, Phone } from "lucide-react";
import type { PublicStorefront } from "@/lib/types/public-storefront";
import { getInitials } from "./storefront-formatters";

interface StorefrontHeroProps {
  storefront: PublicStorefront;
  productCount: number;
}

export function StorefrontHero({ storefront, productCount }: StorefrontHeroProps) {
  const hasContactDetails = Boolean(
    storefront.contact_phone || storefront.contact_email || storefront.address,
  );

  return (
    <>
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
              Browse what is available now, then get in touch with the store for delivery, pickup,
              or payment details.
            </p>
          </div>

          <div className="grid gap-4 border-y border-border/60 py-5 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-end">
            <div>
              <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                Collection
              </p>
              <p className="mt-2 text-3xl font-semibold text-foreground">
                {productCount} {productCount === 1 ? "item" : "items"}
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
    </>
  );
}
