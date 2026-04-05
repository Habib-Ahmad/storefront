import * as React from "react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { act, fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import OnboardPage from "@/app/(auth)/onboard/page";
import { getTemporaryStorefrontSlugPreview } from "@/lib/storefront";

const replaceMock = vi.fn();
const routerMock = { replace: replaceMock };
const mockUseSession = vi.fn();
const mockUseMe = vi.fn();
const mockUseOnboardTenant = vi.fn();
const mutateAsyncMock = vi.fn();
const signOutMock = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => routerMock,
}));

vi.mock("framer-motion", () => ({
  motion: new Proxy(
    {},
    {
      get: (_target, tag) => {
        const Component = ({ children, ...props }: React.HTMLAttributes<HTMLElement>) =>
          React.createElement(tag as string, props, children);
        Component.displayName = `Motion${String(tag)}`;
        return Component;
      },
    },
  ),
}));

vi.mock("@/components/auth-provider", () => ({
  useSession: () => mockUseSession(),
}));

vi.mock("@/hooks/use-auth", () => ({
  useMe: () => mockUseMe(),
  useSignOut: () => signOutMock,
}));

vi.mock("@/hooks/use-tenant", () => ({
  useOnboardTenant: () => mockUseOnboardTenant(),
}));

beforeEach(() => {
  vi.clearAllMocks();

  mockUseSession.mockReturnValue({
    session: { user: { email: "owner@example.com" } },
    loading: false,
  });

  mockUseMe.mockReturnValue({
    data: { onboarded: false },
    isLoading: false,
  });

  mockUseOnboardTenant.mockReturnValue({
    mutateAsync: mutateAsyncMock,
  });

  window.sessionStorage.clear();
});

afterEach(() => {
  vi.useRealTimers();
});

describe("OnboardPage", () => {
  it("shows the generated temporary storefront link during onboarding", async () => {
    render(<OnboardPage />);
    const user = userEvent.setup();

    const businessNameInput = screen.getByLabelText("Business name");

    await user.type(businessNameInput, "Amina Fashion House");
    expect(screen.queryByLabelText("Store link")).not.toBeInTheDocument();
    expect(screen.getByText(/private temporary storefront link now/i)).toBeInTheDocument();
    expect(
      screen.getByText(getTemporaryStorefrontSlugPreview("Amina Fashion House")),
    ).toBeInTheDocument();
  });

  it("avoids reserved routes in the temporary storefront preview", async () => {
    render(<OnboardPage />);
    const user = userEvent.setup();

    await user.type(screen.getByLabelText("Business name"), "App");

    expect(screen.getByText("app-store")).toBeInTheDocument();
  });

  it("shows the success state and waits before redirecting after onboarding", async () => {
    mutateAsyncMock.mockResolvedValue({
      name: "Amina Fashion House",
      slug: "amina-fashion-house",
    });
    vi.useFakeTimers();

    render(<OnboardPage />);
    fireEvent.change(screen.getByLabelText("Business name"), {
      target: { value: "Amina Fashion House" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Create workspace" }));

    await act(async () => {
      await Promise.resolve();
    });

    expect(screen.getByText("Setting up your store")).toBeInTheDocument();
    expect(replaceMock).not.toHaveBeenCalled();

    await act(async () => {
      await Promise.resolve();
    });

    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000);
    });

    expect(screen.getByText("Success. Redirecting you now.")).toBeInTheDocument();

    expect(replaceMock).not.toHaveBeenCalled();

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1800);
    });

    expect(replaceMock).toHaveBeenCalledWith("/app/storefront");
  }, 10000);

  it("redirects straight to the dashboard when the user is already onboarded", async () => {
    mockUseMe.mockReturnValue({
      data: { onboarded: true },
      isLoading: false,
    });

    render(<OnboardPage />);

    expect(replaceMock).toHaveBeenCalledWith("/app");
  });

  it("shows a banner when the user is returning to finish onboarding", () => {
    window.sessionStorage.setItem("storefront:onboarding-banner", "app-guard");

    render(<OnboardPage />);

    expect(
      screen.getByText("You already have an account. Finish setting up your store to continue."),
    ).toBeInTheDocument();
  });

  it("submits only the workspace fields during onboarding", async () => {
    mutateAsyncMock.mockResolvedValue({
      name: "Amina Fashion House",
      slug: "amina-fashion-house",
    });
    render(<OnboardPage />);
    const user = userEvent.setup();

    await user.type(screen.getByLabelText("Business name"), "Amina Fashion House");
    await user.click(screen.getByRole("button", { name: "Create workspace" }));

    expect(mutateAsyncMock).toHaveBeenCalledWith({
      name: "Amina Fashion House",
      admin_email: "owner@example.com",
    });
  });

  it("lets the user sign out from onboarding", async () => {
    render(<OnboardPage />);
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: "Sign out" }));

    expect(signOutMock).toHaveBeenCalledOnce();
  });
});
