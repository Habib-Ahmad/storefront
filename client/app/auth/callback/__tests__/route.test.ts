import { beforeEach, describe, expect, it, vi } from "vitest";

const { cookieStoreMock, createServerClientMock, exchangeCodeForSessionMock } = vi.hoisted(() => ({
  cookieStoreMock: {
    getAll: vi.fn(() => []),
    set: vi.fn(),
  },
  createServerClientMock: vi.fn(),
  exchangeCodeForSessionMock: vi.fn(),
}));

vi.mock("next/headers", () => ({
  cookies: vi.fn(async () => cookieStoreMock),
}));

vi.mock("@supabase/ssr", () => ({
  createServerClient: createServerClientMock,
}));

import { GET } from "@/app/auth/callback/route";

beforeEach(() => {
  vi.clearAllMocks();
  createServerClientMock.mockReturnValue({
    auth: {
      exchangeCodeForSession: exchangeCodeForSessionMock,
    },
  });
  exchangeCodeForSessionMock.mockResolvedValue({ error: null });
});

describe("auth callback route", () => {
  it("redirects back to login when the callback arrives without an auth code", async () => {
    const response = await GET(
      new Request("http://localhost:3000/auth/callback?next=%2Fapp%2Forders"),
    );

    expect(createServerClientMock).not.toHaveBeenCalled();
    expect(response.headers.get("location")).toBe(
      "http://localhost:3000/login?error=oauth_callback&redirect=%2Fapp%2Forders",
    );
  });

  it("exchanges the code and redirects to the requested in-app path", async () => {
    const response = await GET(
      new Request("http://localhost:3000/auth/callback?code=auth-code&next=%2Fapp%2Forders"),
    );

    expect(exchangeCodeForSessionMock).toHaveBeenCalledWith("auth-code");
    expect(response.headers.get("location")).toBe("http://localhost:3000/app/orders");
  });

  it("falls back to /app when next is missing or unsafe", async () => {
    const response = await GET(
      new Request("http://localhost:3000/auth/callback?code=auth-code&next=https://evil.test"),
    );

    expect(response.headers.get("location")).toBe("http://localhost:3000/app");
  });

  it("redirects back to login when the callback cannot be exchanged", async () => {
    exchangeCodeForSessionMock.mockResolvedValue({ error: new Error("invalid_grant") });

    const response = await GET(
      new Request("http://localhost:3000/auth/callback?code=bad-code&next=%2Fapp%2Forders"),
    );

    expect(response.headers.get("location")).toBe(
      "http://localhost:3000/login?error=oauth_callback&redirect=%2Fapp%2Forders",
    );
  });
});
