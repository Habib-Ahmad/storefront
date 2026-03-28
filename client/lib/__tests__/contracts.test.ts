import { describe, expect, it } from "vitest";
import {
  MeResponseSchema,
  ProductDetailResponseSchema,
  ProductSchema,
  TenantSchema,
  TierSchema,
} from "../contracts";

describe("MeResponseSchema", () => {
  it("accepts the non-onboarded auth response", () => {
    const payload = {
      onboarded: false,
    };

    expect(MeResponseSchema.parse(payload)).toEqual(payload);
  });

  it("accepts the onboarded auth response", () => {
    const payload = {
      onboarded: true,
      tenant: {
        id: "550e8400-e29b-41d4-a716-446655440000",
        tier_id: "550e8400-e29b-41d4-a716-446655440001",
        name: "Funke Fabrics",
        slug: "funke-fabrics",
        contact_email: "hello@funkefabrics.com",
        contact_phone: "+2348012345678",
        address: "12 Allen Avenue, Ikeja",
        logo_url: "https://cdn.example.com/logo.png",
        paystack_subaccount_id: "ACCT_sub_123",
        active_modules: {
          inventory: true,
          payments: true,
          logistics: false,
        },
        status: "active",
        created_at: "2026-03-14T10:00:00Z",
        updated_at: "2026-03-14T10:00:00Z",
      },
      role: "admin",
    };

    expect(MeResponseSchema.parse(payload)).toEqual(payload);
  });

  it("rejects onboarded=true when required fields are missing", () => {
    const payload = {
      onboarded: true,
      role: "admin",
    };

    expect(() => MeResponseSchema.parse(payload)).toThrow();
  });
});

describe("TierSchema", () => {
  it("accepts a tier response", () => {
    const payload = {
      id: "550e8400-e29b-41d4-a716-446655440020",
      name: "Standard",
      debt_ceiling: "50000",
      commission_rate: "0.05",
      created_at: "2026-03-14T10:00:00Z",
      updated_at: "2026-03-14T10:00:00Z",
    };

    expect(TierSchema.parse(payload)).toEqual(payload);
  });

  it("rejects invalid tier responses", () => {
    const payload = {
      id: "not-a-uuid",
      name: "Standard",
      debt_ceiling: "50000",
      commission_rate: "0.05",
      created_at: "2026-03-14T10:00:00Z",
      updated_at: "2026-03-14T10:00:00Z",
    };

    expect(() => TierSchema.parse(payload)).toThrow();
  });
});

describe("TenantSchema", () => {
  it("accepts a tenant response", () => {
    const payload = {
      id: "550e8400-e29b-41d4-a716-446655440030",
      tier_id: "550e8400-e29b-41d4-a716-446655440031",
      name: "Funke Fabrics",
      slug: "funke-fabrics",
      contact_email: "hello@funkefabrics.com",
      contact_phone: "+2348012345678",
      address: "12 Allen Avenue, Ikeja",
      logo_url: "https://cdn.example.com/logo.png",
      paystack_subaccount_id: "ACCT_sub_123",
      active_modules: {
        inventory: true,
        payments: true,
        logistics: false,
      },
      status: "active",
      created_at: "2026-03-14T10:00:00Z",
      updated_at: "2026-03-14T10:00:00Z",
    };

    expect(TenantSchema.parse(payload)).toEqual(payload);
  });

  it("rejects invalid tenant responses", () => {
    const payload = {
      id: "550e8400-e29b-41d4-a716-446655440030",
      tier_id: "550e8400-e29b-41d4-a716-446655440031",
      name: "Funke Fabrics",
      slug: "funke-fabrics",
      active_modules: {
        inventory: true,
        payments: true,
      },
      status: "active",
      created_at: "2026-03-14T10:00:00Z",
      updated_at: "2026-03-14T10:00:00Z",
    };

    expect(() => TenantSchema.parse(payload)).toThrow();
  });
});

describe("ProductSchema", () => {
  it("accepts a product response", () => {
    const payload = {
      id: "550e8400-e29b-41d4-a716-446655440010",
      tenant_id: "550e8400-e29b-41d4-a716-446655440011",
      name: "Ankara Shirt",
      description: "Bright patterned shirt",
      category: "Fashion",
      is_available: true,
      created_at: "2026-03-14T10:00:00Z",
      updated_at: "2026-03-14T10:00:00Z",
    };

    expect(ProductSchema.parse(payload)).toEqual(payload);
  });

  it("rejects invalid product responses", () => {
    const payload = {
      id: "not-a-uuid",
      tenant_id: "550e8400-e29b-41d4-a716-446655440011",
      name: "Ankara Shirt",
      is_available: true,
      created_at: "2026-03-14T10:00:00Z",
      updated_at: "2026-03-14T10:00:00Z",
    };

    expect(() => ProductSchema.parse(payload)).toThrow();
  });
});

describe("ProductDetailResponseSchema", () => {
  it("accepts a product detail response", () => {
    const payload = {
      product: {
        id: "550e8400-e29b-41d4-a716-446655440010",
        tenant_id: "550e8400-e29b-41d4-a716-446655440011",
        name: "Ankara Shirt",
        description: "Bright patterned shirt",
        category: "Fashion",
        is_available: true,
        created_at: "2026-03-14T10:00:00Z",
        updated_at: "2026-03-14T10:00:00Z",
      },
      variants: [
        {
          id: "550e8400-e29b-41d4-a716-446655440012",
          product_id: "550e8400-e29b-41d4-a716-446655440010",
          sku: "RED-M",
          attributes: { color: "Red", size: "M" },
          price: "15000",
          cost_price: "9000",
          stock_qty: 8,
          is_default: true,
          created_at: "2026-03-14T10:00:00Z",
          updated_at: "2026-03-14T10:00:00Z",
        },
      ],
      images: [
        {
          id: "550e8400-e29b-41d4-a716-446655440013",
          product_id: "550e8400-e29b-41d4-a716-446655440010",
          url: "https://cdn.example.com/products/ankara-shirt.jpg",
          sort_order: 0,
          is_primary: true,
          created_at: "2026-03-14T10:00:00Z",
        },
      ],
    };

    expect(ProductDetailResponseSchema.parse(payload)).toEqual(payload);
  });

  it("rejects product detail responses with invalid nested fields", () => {
    const payload = {
      product: {
        id: "550e8400-e29b-41d4-a716-446655440010",
        tenant_id: "550e8400-e29b-41d4-a716-446655440011",
        name: "Ankara Shirt",
        is_available: true,
        created_at: "2026-03-14T10:00:00Z",
        updated_at: "2026-03-14T10:00:00Z",
      },
      variants: [
        {
          id: "550e8400-e29b-41d4-a716-446655440012",
          product_id: "550e8400-e29b-41d4-a716-446655440010",
          sku: "RED-M",
          attributes: {},
          price: "15000",
          cost_price: "9000",
          stock_qty: 8,
          is_default: true,
          created_at: "2026-03-14T10:00:00Z",
          updated_at: "2026-03-14T10:00:00Z",
        },
      ],
      images: [
        {
          id: "550e8400-e29b-41d4-a716-446655440013",
          product_id: "550e8400-e29b-41d4-a716-446655440010",
          url: "not-a-url",
          sort_order: 0,
          is_primary: true,
          created_at: "2026-03-14T10:00:00Z",
        },
      ],
    };

    expect(() => ProductDetailResponseSchema.parse(payload)).toThrow();
  });
});
