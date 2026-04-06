import {
  PublicStorefrontProductDetailResponseSchema,
  PublicStorefrontResponseSchema,
  type PublicStorefrontProductDetailResponse,
  type PublicStorefrontResponse,
} from "./types/public-storefront";

export class PublicStorefrontError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
    this.name = "PublicStorefrontError";
  }
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "";

export async function getPublicStorefront(slug: string): Promise<PublicStorefrontResponse> {
  const response = await fetch(`${API_BASE}/storefronts/${encodeURIComponent(slug)}`, {
    cache: "no-store",
    headers: {
      "Content-Type": "application/json",
    },
  });

  if (!response.ok) {
    const payload = await response.json().catch(() => ({}));
    throw new PublicStorefrontError(response.status, payload.error ?? "Unable to load storefront");
  }

  return PublicStorefrontResponseSchema.parse(await response.json());
}

export async function getPublicStorefrontProduct(
  slug: string,
  productId: string,
): Promise<PublicStorefrontProductDetailResponse> {
  const response = await fetch(
    `${API_BASE}/storefronts/${encodeURIComponent(slug)}/products/${encodeURIComponent(productId)}`,
    {
      cache: "no-store",
      headers: {
        "Content-Type": "application/json",
      },
    },
  );

  if (!response.ok) {
    const payload = await response.json().catch(() => ({}));
    throw new PublicStorefrontError(response.status, payload.error ?? "Unable to load product");
  }

  return PublicStorefrontProductDetailResponseSchema.parse(await response.json());
}
