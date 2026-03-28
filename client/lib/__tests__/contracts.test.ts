import { describe, expect, it } from "vitest";
import { MeResponseSchema } from "../contracts";

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
