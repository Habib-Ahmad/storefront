import { notFound } from "next/navigation";
import { PublicStorefrontError, getPublicStorefrontProduct } from "@/lib/public-storefront";
import { PublicCheckout } from "./public-checkout";

interface Props {
  params: Promise<{ slug: string; productId: string }>;
  searchParams: Promise<{ variant?: string | string[] }>;
}

export default async function PublicCheckoutPage({ params, searchParams }: Props) {
  const { slug, productId } = await params;
  const query = await searchParams;
  const initialVariantId = Array.isArray(query.variant) ? query.variant[0] : query.variant;

  try {
    const detail = await getPublicStorefrontProduct(slug, productId);
    return <PublicCheckout detail={detail} initialVariantId={initialVariantId ?? null} />;
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
