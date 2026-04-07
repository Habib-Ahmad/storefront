import { StorefrontBasketCheckout } from "./storefront-basket-checkout";

interface Props {
  params: Promise<{ slug: string }>;
}

export default async function StorefrontCheckoutPage({ params }: Props) {
  const { slug } = await params;

  return <StorefrontBasketCheckout slug={slug} />;
}
