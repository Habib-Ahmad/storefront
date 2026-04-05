import { describe, it, expect, vi, beforeEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import StorefrontPage from "@/app/app/storefront/page";

const mockUseMe = vi.fn();
const mockUseUpdateStorefront = vi.fn();
const mutateAsyncMock = vi.fn();
const refetchMock = vi.fn();

vi.mock("@/hooks/use-auth", () => ({
  useMe: () => mockUseMe(),
}));

vi.mock("@/hooks/use-tenant", () => ({
  useUpdateStorefront: () => mockUseUpdateStorefront(),
}));

beforeEach(() => {
  vi.clearAllMocks();

  mockUseMe.mockReturnValue({
    data: {
      onboarded: true,
      role: "admin",
      tenant: {
        id: "550e8400-e29b-41d4-a716-446655440000",
        tier_id: "550e8400-e29b-41d4-a716-446655440001",
        name: "Funke Fabrics",
        slug: "funke-fabrics",
        storefront_published: false,
        active_modules: {
          inventory: true,
          payments: true,
          logistics: false,
        },
        status: "active",
        created_at: "2026-03-14T10:00:00Z",
        updated_at: "2026-03-14T10:00:00Z",
      },
    },
    isLoading: false,
    isError: false,
    error: null,
    refetch: refetchMock,
  });

  mockUseUpdateStorefront.mockReturnValue({
    mutateAsync: mutateAsyncMock,
  });
});

describe("StorefrontPage", () => {
  it("lets the merchant claim and save a public storefront slug", async () => {
    render(<StorefrontPage />);

    fireEvent.change(screen.getByLabelText("Public storefront slug"), {
      target: { value: "Funke Fabrics HQ" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Save slug" }));

    await waitFor(() => {
      expect(mutateAsyncMock).toHaveBeenCalledWith({
        slug: "funke-fabrics-hq",
        storefront_published: false,
      });
    });
  });

  it("shows the reserved storefront link before the storefront is published", () => {
    render(<StorefrontPage />);

    expect(screen.getByRole("button", { name: "Publish storefront" })).toBeInTheDocument();
    expect(screen.getByText("Reserved storefront link")).toBeInTheDocument();
    expect(screen.getByText("storefront.com/funke-fabrics")).toBeInTheDocument();
  });

  it("shows an error state instead of an endless loader when tenant data is missing", () => {
    mockUseMe.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: new Error("tenant lookup failed"),
      refetch: refetchMock,
    });

    render(<StorefrontPage />);

    expect(screen.getByText("We couldn't load your storefront")).toBeInTheDocument();
    expect(screen.getByText("tenant lookup failed")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Retry" })).toBeInTheDocument();
  });
});
