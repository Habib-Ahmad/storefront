import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { api } from "../api";

const ok = (data: unknown, status = 200) =>
  Promise.resolve(
    new Response(JSON.stringify(data), { status, headers: { "Content-Type": "application/json" } }),
  );
const no_content = () => Promise.resolve(new Response(null, { status: 204 }));

beforeEach(() => {
  vi.stubGlobal("fetch", vi.fn());
  api.setToken(null);
});
afterEach(() => {
  vi.unstubAllGlobals();
});

describe("pagination offset", () => {
  it("calculates offset as (page - 1) * per_page", async () => {
    vi.mocked(fetch).mockReturnValue(ok({ data: [], total: 0, page: 3, per_page: 10 }));
    await api.getProducts({ page: 3, per_page: 10 });
    expect(fetch).toHaveBeenCalledWith(
      expect.stringMatching(/limit=10.*offset=20/),
      expect.any(Object),
    );
  });
});

describe("204 responses", () => {
  it("returns undefined instead of trying to parse an empty body", async () => {
    vi.mocked(fetch).mockReturnValue(no_content());
    await expect(api.updateTenant({ name: "x" })).resolves.toBeUndefined();
  });
});

describe("error parsing", () => {
  it("throws ApiError with the server message and status", async () => {
    vi.mocked(fetch).mockReturnValue(ok({ error: "not found" }, 404));
    await expect(api.getMe()).rejects.toMatchObject({
      name: "ApiError",
      status: 404,
      message: "not found",
    });
  });

  it("includes field-level errors from the response body", async () => {
    vi.mocked(fetch).mockReturnValue(ok({ error: "invalid", errors: { name: "required" } }, 422));
    await expect(api.onboard({ name: "", admin_email: "" })).rejects.toMatchObject({
      fields: { name: "required" },
    });
  });

  it("falls back to 'Unknown error' when the body has no error field", async () => {
    vi.mocked(fetch).mockReturnValue(Promise.resolve(new Response("bad gateway", { status: 502 })));
    await expect(api.getMe()).rejects.toMatchObject({ status: 502, message: "Unknown error" });
  });

  it("defaults missing storefront_published to false for onboarding responses", async () => {
    vi.mocked(fetch).mockReturnValue(
      ok({
        id: "550e8400-e29b-41d4-a716-446655440000",
        tier_id: "550e8400-e29b-41d4-a716-446655440001",
        name: "Funke Fabrics",
        slug: "funke-fabrics",
        active_modules: {
          inventory: true,
          payments: true,
          logistics: false,
        },
        status: "active",
        created_at: "2026-03-14T10:00:00Z",
        updated_at: "2026-03-14T10:00:00Z",
      }),
    );

    await expect(
      api.onboard({ name: "Funke Fabrics", admin_email: "owner@example.com" }),
    ).resolves.toMatchObject({ storefront_published: false, slug: "funke-fabrics" });
  });

  it("defaults missing storefront_published to false inside auth me tenant payloads", async () => {
    vi.mocked(fetch).mockReturnValue(
      ok({
        onboarded: true,
        tenant: {
          id: "550e8400-e29b-41d4-a716-446655440000",
          tier_id: "550e8400-e29b-41d4-a716-446655440001",
          name: "Funke Fabrics",
          slug: "funke-fabrics",
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
      }),
    );

    await expect(api.getMe()).resolves.toMatchObject({
      onboarded: true,
      tenant: expect.objectContaining({ storefront_published: false }),
    });
  });
});

describe("auth header", () => {
  it("sends Bearer token when set", async () => {
    api.setToken("my-jwt");
    vi.mocked(fetch).mockReturnValue(ok([]));
    await api.getTiers();
    const [, init] = vi.mocked(fetch).mock.calls[0];
    expect((init!.headers as Record<string, string>)["Authorization"]).toBe("Bearer my-jwt");
  });

  it("omits Authorization when no token is set", async () => {
    vi.mocked(fetch).mockReturnValue(ok([]));
    await api.getTiers();
    const [, init] = vi.mocked(fetch).mock.calls[0];
    expect((init!.headers as Record<string, string>)["Authorization"]).toBeUndefined();
  });
});

describe("401 retry", () => {
  it("retries the request with a refreshed token and succeeds", async () => {
    const refreshHandler = vi.fn().mockResolvedValue("new-token");
    api.setToken("expired-token");
    api.setRefreshHandler(refreshHandler);

    vi.mocked(fetch)
      .mockReturnValueOnce(ok({ error: "unauthorized" }, 401))
      .mockReturnValueOnce(
        ok({
          onboarded: true,
          tenant: {
            id: "550e8400-e29b-41d4-a716-446655440000",
            tier_id: "550e8400-e29b-41d4-a716-446655440001",
            name: "Funke Fabrics",
            slug: "funke-fabrics",
            storefront_published: true,
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
        }),
      );

    await api.getMe();

    expect(refreshHandler).toHaveBeenCalledOnce();
    expect(fetch).toHaveBeenCalledTimes(2);
    const [, secondInit] = vi.mocked(fetch).mock.calls[1];
    expect((secondInit!.headers as Record<string, string>)["Authorization"]).toBe(
      "Bearer new-token",
    );

    api.setRefreshHandler(null);
  });

  it("throws when the refresh handler cannot obtain a new token", async () => {
    api.setToken("expired");
    api.setRefreshHandler(vi.fn().mockResolvedValue(null));

    vi.mocked(fetch).mockReturnValue(ok({ error: "unauthorized" }, 401));

    await expect(api.getMe()).rejects.toMatchObject({ status: 401 });
    expect(fetch).toHaveBeenCalledOnce();

    api.setRefreshHandler(null);
  });
});
