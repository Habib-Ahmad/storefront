import { describe, expect, it } from "vitest";
import {
  AnalyticsSummarySchema,
  CreatePublicStorefrontDeliveryQuoteRequestSchema,
  MeResponseSchema,
  OrderItemSchema,
  OrderSchema,
  PublicStorefrontDeliveryQuoteResponseSchema,
  ProductDetailResponseSchema,
  ProductSchema,
  TenantSchema,
  TierSchema,
  TrackingResponseSchema,
  TransactionSchema,
  WalletSchema,
} from "../types";
import { CreatePublicStorefrontOrderRequestSchema } from "../types/public-storefront";

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
        storefront_published: true,
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
      storefront_published: false,
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
      storefront_published: false,
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

describe("OrderSchema", () => {
  it("accepts an order response", () => {
    const payload = {
      id: "550e8400-e29b-41d4-a716-446655440100",
      tenant_id: "550e8400-e29b-41d4-a716-446655440101",
      tracking_slug: "abc123def456",
      is_delivery: true,
      customer_name: "Amina Bello",
      customer_phone: "+2348012345678",
      customer_email: "amina@example.com",
      shipping_address: "12 Allen Avenue, Ikeja",
      note: "Handle with care",
      total_amount: "18500",
      shipping_fee: "1500",
      payment_method: "cash",
      payment_status: "paid",
      fulfillment_status: "completed",
      created_at: "2026-03-14T10:00:00Z",
      updated_at: "2026-03-14T10:00:00Z",
    };

    expect(OrderSchema.parse(payload)).toEqual(payload);
  });

  it("rejects invalid order responses", () => {
    const payload = {
      id: "550e8400-e29b-41d4-a716-446655440100",
      tenant_id: "550e8400-e29b-41d4-a716-446655440101",
      tracking_slug: "abc123def456",
      is_delivery: true,
      total_amount: "18500",
      shipping_fee: "1500",
      payment_method: "card",
      payment_status: "pending",
      fulfillment_status: "processing",
      created_at: "2026-03-14T10:00:00Z",
      updated_at: "2026-03-14T10:00:00Z",
    };

    expect(() => OrderSchema.parse(payload)).toThrow();
  });
});

describe("Public storefront delivery quote schemas", () => {
  it("accepts a delivery quote response", () => {
    const payload = {
      storefront: {
        name: "Funke Fabrics",
        slug: "funke-fabrics",
        logo_url: null,
        contact_email: "hello@funkefabrics.com",
        contact_phone: "+2348012345678",
        address: "12 Allen Avenue, Ikeja",
        delivery: {
          enabled: true,
          ready: true,
          unavailable_reason: null,
        },
      },
      options: [
        {
          id: "123:bike:dropoff",
          courier_id: "123",
          courier_name: "Kwik",
          service_code: "bike",
          service_type: "dropoff",
          amount: "3500",
          currency: "NGN",
          delivery_eta: "same day",
          tracking_level: 4,
          is_fastest: true,
          is_cheapest: false,
        },
      ],
      debug: {
        sender_address_code: 1,
        receiver_address_code: 2,
        category_id: 3,
        category_name: "Fashion wears",
        package_box: "small box",
        estimated_weight_kg: "0.35",
      },
    };

    expect(PublicStorefrontDeliveryQuoteResponseSchema.parse(payload)).toEqual(payload);
  });

  it("rejects delivery orders without a selected delivery option", () => {
    expect(() =>
      CreatePublicStorefrontOrderRequestSchema.parse({
        is_delivery: true,
        checkout_id: "550e8400-e29b-41d4-a716-446655440099",
        customer_phone: "08012345678",
        shipping_address: "23 Abuja",
        items: [{ variant_id: "550e8400-e29b-41d4-a716-446655440001", quantity: 1 }],
      }),
    ).toThrow("delivery_option is required for delivery orders");
  });

  it("accepts pickup orders without shipping state", () => {
    const payload = {
      is_delivery: false,
      checkout_id: "550e8400-e29b-41d4-a716-446655440099",
      customer_phone: "08012345678",
      items: [{ variant_id: "550e8400-e29b-41d4-a716-446655440001", quantity: 1 }],
    };

    expect(CreatePublicStorefrontOrderRequestSchema.parse(payload)).toEqual(payload);
  });

  it("accepts delivery quote request payloads", () => {
    const payload = {
      customer_name: "Chidi",
      customer_phone: "08012345678",
      shipping_address: "23 Abuja",
      items: [{ variant_id: "550e8400-e29b-41d4-a716-446655440001", quantity: 1 }],
    };

    expect(CreatePublicStorefrontDeliveryQuoteRequestSchema.parse(payload)).toEqual(payload);
  });
});

describe("OrderItemSchema", () => {
  it("accepts an order item response", () => {
    const payload = {
      id: "550e8400-e29b-41d4-a716-446655440110",
      order_id: "550e8400-e29b-41d4-a716-446655440100",
      variant_id: "550e8400-e29b-41d4-a716-446655440111",
      quantity: 2,
      price_at_sale: "8500",
      cost_price_at_sale: "5000",
      product_name: "Ankara Shirt",
      variant_label: "Red / Medium",
    };

    expect(OrderItemSchema.parse(payload)).toEqual(payload);
  });

  it("rejects invalid order item responses", () => {
    const payload = {
      id: "550e8400-e29b-41d4-a716-446655440110",
      order_id: "550e8400-e29b-41d4-a716-446655440100",
      variant_id: "550e8400-e29b-41d4-a716-446655440111",
      quantity: "2",
      price_at_sale: "8500",
    };

    expect(() => OrderItemSchema.parse(payload)).toThrow();
  });
});

describe("WalletSchema", () => {
  it("accepts a wallet response", () => {
    const payload = {
      id: "550e8400-e29b-41d4-a716-446655440120",
      tenant_id: "550e8400-e29b-41d4-a716-446655440121",
      available_balance: "25000",
      pending_balance: "5000",
      last_transaction_id: "550e8400-e29b-41d4-a716-446655440122",
      last_reconciliation_at: "2026-03-14T10:00:00Z",
    };

    expect(WalletSchema.parse(payload)).toEqual(payload);
  });

  it("rejects invalid wallet responses", () => {
    const payload = {
      id: "not-a-uuid",
      tenant_id: "550e8400-e29b-41d4-a716-446655440121",
      available_balance: "25000",
      pending_balance: "5000",
    };

    expect(() => WalletSchema.parse(payload)).toThrow();
  });
});

describe("TransactionSchema", () => {
  it("accepts a transaction response", () => {
    const payload = {
      id: "550e8400-e29b-41d4-a716-446655440130",
      wallet_id: "550e8400-e29b-41d4-a716-446655440120",
      order_id: "550e8400-e29b-41d4-a716-446655440100",
      amount: "15000",
      running_balance: "25000",
      type: "credit",
      signature: "abc123signature",
      created_at: "2026-03-14T10:00:00Z",
    };

    expect(TransactionSchema.parse(payload)).toEqual(payload);
  });

  it("rejects invalid transaction responses", () => {
    const payload = {
      id: "550e8400-e29b-41d4-a716-446655440130",
      wallet_id: "550e8400-e29b-41d4-a716-446655440120",
      amount: "15000",
      running_balance: "25000",
      type: "deposit",
      signature: "abc123signature",
      created_at: "2026-03-14T10:00:00Z",
    };

    expect(() => TransactionSchema.parse(payload)).toThrow();
  });
});

describe("TrackingResponseSchema", () => {
  it("accepts a tracking response", () => {
    const payload = {
      tracking_slug: "abc123def456",
      is_delivery: true,
      storefront_slug: "funke-fabrics",
      customer_name: "Amina Bello",
      payment_status: "paid",
      fulfillment_status: "completed",
    };

    expect(TrackingResponseSchema.parse(payload)).toEqual(payload);
  });

  it("accepts anonymous tracking responses without customer_name", () => {
    const payload = {
      tracking_slug: "abc123def456",
      payment_status: "paid",
      fulfillment_status: "completed",
    };

    expect(TrackingResponseSchema.parse(payload)).toEqual(payload);
  });

  it("rejects invalid tracking responses", () => {
    const payload = {
      tracking_slug: "abc123def456",
      customer_name: "Amina Bello",
      payment_status: "settled",
      fulfillment_status: "shipped",
    };

    expect(() => TrackingResponseSchema.parse(payload)).toThrow();
  });
});

describe("AnalyticsSummarySchema", () => {
  it("accepts an analytics summary response", () => {
    const payload = {
      total_revenue: "125000",
      total_cost: "70000",
      total_profit: "55000",
      order_count: 12,
      avg_order_value: "10416.67",
      by_payment_method: [
        { method: "online", revenue: "85000", count: 8 },
        { method: "cash", revenue: "40000", count: 4 },
      ],
      top_products: [
        { product_name: "Ankara Shirt", quantity_sold: 6, revenue: "45000" },
        { product_name: "Denim Jacket", quantity_sold: 3, revenue: "36000" },
      ],
      period: {
        from: "2026-03-01T00:00:00Z",
        to: "2026-03-31T23:59:59Z",
      },
    };

    expect(AnalyticsSummarySchema.parse(payload)).toEqual(payload);
  });

  it("rejects invalid analytics summary responses", () => {
    const payload = {
      total_revenue: "125000",
      total_cost: "70000",
      total_profit: "55000",
      order_count: "12",
      avg_order_value: "10416.67",
      by_payment_method: [{ method: "online", revenue: "85000", count: 8 }],
      top_products: [{ product_name: "Ankara Shirt", quantity_sold: 6, revenue: "45000" }],
      period: {
        from: "2026-03-01T00:00:00Z",
        to: "2026-03-31T23:59:59Z",
      },
    };

    expect(() => AnalyticsSummarySchema.parse(payload)).toThrow();
  });
});
