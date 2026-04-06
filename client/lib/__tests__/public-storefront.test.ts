import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  createPublicStorefrontOrder,
  PublicStorefrontError,
  getPublicStorefront,
  getPublicStorefrontProduct,
} from "../public-storefront";

const ok = (data: unknown, status = 200) =>
  Promise.resolve(
    new Response(JSON.stringify(data), { status, headers: { "Content-Type": "application/json" } }),
  );

beforeEach(() => {
  vi.stubGlobal("fetch", vi.fn());
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("getPublicStorefront", () => {
  it("parses the published storefront payload", async () => {
    vi.mocked(fetch).mockReturnValue(
      ok({
        storefront: {
          name: "Funke Fabrics",
          slug: "funke-fabrics",
          logo_url: null,
          contact_email: "hello@funkefabrics.com",
          contact_phone: "+2348012345678",
          address: "12 Allen Avenue, Ikeja",
        },
        products: [
          {
            id: "550e8400-e29b-41d4-a716-446655440000",
            name: "Ankara Set",
            description: "A bright two-piece set",
            category: "Fashion",
            image_url: "https://cdn.example.com/ankara.png",
            price: "24500",
            in_stock: true,
          },
        ],
      }),
    );

    await expect(getPublicStorefront("funke-fabrics")).resolves.toMatchObject({
      storefront: { slug: "funke-fabrics" },
      products: [expect.objectContaining({ name: "Ankara Set", in_stock: true })],
    });
  });

  it("throws a typed error when the storefront is missing", async () => {
    vi.mocked(fetch).mockReturnValue(ok({ error: "storefront not found" }, 404));

    await expect(getPublicStorefront("missing-store")).rejects.toEqual(
      expect.objectContaining<Partial<PublicStorefrontError>>({
        name: "PublicStorefrontError",
        status: 404,
        message: "storefront not found",
      }),
    );
  });
});

describe("getPublicStorefrontProduct", () => {
  it("parses the public product detail payload", async () => {
    vi.mocked(fetch).mockReturnValue(
      ok({
        storefront: {
          name: "Funke Fabrics",
          slug: "funke-fabrics",
          logo_url: null,
          contact_email: "hello@funkefabrics.com",
          contact_phone: "+2348012345678",
          address: "12 Allen Avenue, Ikeja",
        },
        product: {
          id: "550e8400-e29b-41d4-a716-446655440000",
          name: "Ankara Set",
          description: "A bright two-piece set",
          category: "Fashion",
          image_url: "https://cdn.example.com/ankara.png",
          price: "24500",
          in_stock: true,
        },
        variants: [
          {
            id: "550e8400-e29b-41d4-a716-446655440001",
            attributes: { size: "M", color: "Blue" },
            price: "24500",
            in_stock: true,
            is_default: true,
          },
        ],
        images: [
          {
            id: "550e8400-e29b-41d4-a716-446655440002",
            url: "https://cdn.example.com/ankara.png",
            sort_order: 0,
            is_primary: true,
          },
        ],
      }),
    );

    await expect(
      getPublicStorefrontProduct("funke-fabrics", "550e8400-e29b-41d4-a716-446655440000"),
    ).resolves.toMatchObject({
      product: { name: "Ankara Set" },
      variants: [expect.objectContaining({ is_default: true })],
    });
  });
});

describe("createPublicStorefrontOrder", () => {
  it("parses the public checkout response", async () => {
    vi.mocked(fetch).mockReturnValue(
      ok(
        {
          storefront: {
            name: "Funke Fabrics",
            slug: "funke-fabrics",
            logo_url: null,
            contact_email: "hello@funkefabrics.com",
            contact_phone: "+2348012345678",
            address: "12 Allen Avenue, Ikeja",
          },
          order: {
            tracking_slug: "abc123def456",
            is_delivery: true,
            customer_name: "Chidi",
            customer_phone: "08012345678",
            customer_email: "chidi@example.com",
            shipping_address: "23 Abuja",
            note: "Please call on arrival",
            total_amount: "49000",
            shipping_fee: "0",
            payment_method: "online",
            payment_status: "pending",
            fulfillment_status: "processing",
          },
        },
        201,
      ),
    );

    await expect(
      createPublicStorefrontOrder("funke-fabrics", {
        is_delivery: true,
        customer_name: "Chidi",
        customer_phone: "08012345678",
        customer_email: "chidi@example.com",
        shipping_address: "23 Abuja",
        note: "Please call on arrival",
        items: [{ variant_id: "550e8400-e29b-41d4-a716-446655440001", quantity: 2 }],
      }),
    ).resolves.toMatchObject({
      storefront: { slug: "funke-fabrics" },
      order: { tracking_slug: "abc123def456", payment_status: "pending" },
    });
  });
});
