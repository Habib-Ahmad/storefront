import { describe, it, expect, vi, beforeEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import LogisticsSettingsPage from "@/app/app/settings/logistics/page";

const mockUseMe = vi.fn();
const mockUseSession = vi.fn();
const mockUseUpdateTenant = vi.fn();
const mutateAsyncMock = vi.fn();
const refetchMock = vi.fn();

vi.mock("@/hooks/use-auth", () => ({
  useMe: () => mockUseMe(),
}));

vi.mock("@/components/auth-provider", () => ({
  useSession: () => mockUseSession(),
}));

vi.mock("@/hooks/use-tenant", () => ({
  useUpdateTenant: () => mockUseUpdateTenant(),
}));

beforeEach(() => {
  vi.clearAllMocks();

  mockUseSession.mockReturnValue({
    session: {
      user: {
        email: "owner@example.com",
      },
    },
  });

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
        contact_email: null,
        contact_phone: null,
        address: null,
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

  mockUseUpdateTenant.mockReturnValue({
    mutateAsync: mutateAsyncMock,
  });
});

describe("LogisticsSettingsPage", () => {
  it("prefills the logistics email from the signed-in admin when tenant email is empty", () => {
    render(<LogisticsSettingsPage />);

    expect(screen.getByLabelText("Logistics email")).toHaveValue("owner@example.com");
    expect(screen.getByLabelText("Country")).toHaveValue("Nigeria");
    expect(screen.getByLabelText("State")).toHaveDisplayValue("Select state");
    expect(screen.getByLabelText("City or area")).toHaveValue("");
    expect(screen.getByText("Shipbubble wallet funding")).toBeInTheDocument();
  });

  it("saves a structured logistics profile for admins", async () => {
    render(<LogisticsSettingsPage />);

    fireEvent.change(screen.getByLabelText("Pickup phone"), {
      target: { value: "08012345678" },
    });
    fireEvent.change(screen.getByLabelText("Street address"), {
      target: { value: "16 Owerri Street, War College, Gwarinpa" },
    });
    fireEvent.change(screen.getByLabelText("State"), {
      target: { value: "FCT" },
    });
    fireEvent.change(screen.getByLabelText("City or area"), {
      target: { value: "Abuja" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Save logistics setup" }));

    await waitFor(() => {
      expect(mutateAsyncMock).toHaveBeenCalledWith({
        name: "Funke Fabrics",
        contact_email: "owner@example.com",
        contact_phone: "08012345678",
        address: "16 Owerri Street, War College, Gwarinpa, Abuja, FCT, Nigeria",
      });
    });
  });

  it("blocks non-admin users from editing logistics settings", () => {
    mockUseMe.mockReturnValue({
      data: {
        onboarded: true,
        role: "staff",
        tenant: {
          id: "550e8400-e29b-41d4-a716-446655440000",
          tier_id: "550e8400-e29b-41d4-a716-446655440001",
          name: "Funke Fabrics",
          slug: "funke-fabrics",
          storefront_published: false,
          contact_email: null,
          contact_phone: null,
          address: null,
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

    render(<LogisticsSettingsPage />);

    expect(
      screen.getByText("Only store admins can update delivery and logistics setup."),
    ).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Save logistics setup" })).not.toBeInTheDocument();
  });
});
