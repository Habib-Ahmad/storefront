import { notFound } from "next/navigation";
import { PublicStorefrontActions } from "@/components/public-storefront-actions";
import { PublicPendingOrderBanner } from "@/components/public-pending-order-banner";
import { PublicStorefrontError, getPublicStorefront } from "@/lib/public-storefront";
import { StorefrontCatalog } from "./storefront-catalog";
import { StorefrontHero } from "./storefront-hero";
import { StorefrontUnavailable } from "./storefront-unavailable";

interface Props {
  params: Promise<{ slug: string }>;
}

export default async function StorefrontPage({ params }: Props) {
  const { slug } = await params;

  try {
    const { storefront, products } = await getPublicStorefront(slug);

    return (
      <main className="min-h-screen bg-background text-foreground">
        <section className="mx-auto w-full max-w-7xl px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
          <div className="mb-4 flex justify-end">
            <PublicStorefrontActions slug={storefront.slug} />
          </div>
          <div className="mb-4">
            <PublicPendingOrderBanner storefrontSlug={storefront.slug} />
          </div>
          <StorefrontHero storefront={storefront} productCount={products.length} />
          <StorefrontCatalog slug={storefront.slug} products={products} />
        </section>
      </main>
    );
  } catch (error) {
    if (error instanceof PublicStorefrontError && error.status === 404) {
      notFound();
    }

    return <StorefrontUnavailable />;
  }
}
