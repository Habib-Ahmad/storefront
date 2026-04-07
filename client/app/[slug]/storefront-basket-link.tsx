"use client";

import Link from "next/link";
import { ShoppingCart } from "lucide-react";
import { useStorefrontCart } from "@/lib/storefront-cart";

interface StorefrontBasketLinkProps {
  slug: string;
}

export function StorefrontBasketLink({ slug }: StorefrontBasketLinkProps) {
  const { itemCount } = useStorefrontCart(slug);

  return (
    <Link
      href={`/${slug}/checkout`}
      className="inline-flex items-center gap-2 rounded-full border border-border/70 bg-card px-4 py-2 text-sm font-medium text-foreground transition-colors hover:border-foreground/20"
    >
      <ShoppingCart className="h-4 w-4" />
      Cart
      <span className="inline-flex min-w-5 items-center justify-center rounded-full bg-foreground px-1.5 py-0.5 text-xs font-semibold text-background">
        {itemCount}
      </span>
    </Link>
  );
}
