import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import AppLayout from "@/app/app/layout";

const replaceMock = vi.fn();
const mockUseSession = vi.fn();
const mockUseMe = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: replaceMock }),
}));

vi.mock("@/components/auth-provider", () => ({
  useSession: () => mockUseSession(),
}));

vi.mock("@/hooks/use-auth", () => ({
  useMe: () => mockUseMe(),
}));

vi.mock("@/components/layout/sidebar", () => ({ Sidebar: () => <div>Sidebar</div> }));
vi.mock("@/components/layout/bottom-nav", () => ({ BottomNav: () => <div>BottomNav</div> }));
vi.mock("@/components/layout/header", () => ({ Header: () => <div>Header</div> }));

beforeEach(() => {
  vi.clearAllMocks();
  mockUseSession.mockReturnValue({
    session: { user: { id: "user-1" } },
    loading: false,
  });
  mockUseMe.mockReturnValue({
    data: { onboarded: false },
    isLoading: false,
  });
});

describe("AppLayout", () => {
  it("redirects signed-in users who have not finished onboarding", () => {
    render(
      <AppLayout>
        <div>Dashboard</div>
      </AppLayout>,
    );

    expect(window.sessionStorage.getItem("storefront:onboarding-banner")).toBe("app-guard");
    expect(replaceMock).toHaveBeenCalledWith("/onboard");
    expect(screen.getByText("Almost there")).toBeInTheDocument();
  });
});
