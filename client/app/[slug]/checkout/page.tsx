import { notFound } from "next/navigation";
import { PublicStorefrontError, getPublicStorefront } from "@/lib/public-storefront";
import { StorefrontBasketCheckout } from "./storefront-basket-checkout";

interface Props {
  params: Promise<{ slug: string }>;
}

export default async function StorefrontCheckoutPage({ params }: Props) {
  const { slug } = await params;

  try {
    const { storefront } = await getPublicStorefront(slug);
    return <StorefrontBasketCheckout storefront={storefront} />;
  } catch (error) {
    if (error instanceof PublicStorefrontError && error.status === 404) {
      notFound();
    }

    return (
      <main className="flex min-h-screen items-center justify-center px-4 py-12">
        <div className="max-w-xl rounded-[1.75rem] border border-border/60 bg-card p-8 text-center">
          <p className="text-xs font-medium tracking-[0.22em] text-muted-foreground uppercase">
            Checkout unavailable
          </p>
          <h1 className="mt-3 text-3xl font-semibold tracking-tight text-foreground">
            This checkout could not be loaded right now
          </h1>
          <p className="mt-3 text-sm leading-6 text-muted-foreground">
            Please try again in a moment.
          </p>
        </div>
      </main>
    );
  }
}
