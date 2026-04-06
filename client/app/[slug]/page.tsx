import { notFound } from "next/navigation";
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
          <StorefrontHero storefront={storefront} productCount={products.length} />
          <StorefrontCatalog products={products} />
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
