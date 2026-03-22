import * as React from "react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { act, fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import OnboardPage from "@/app/(auth)/onboard/page";

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
  it("auto-generates the store link from the business name until the user edits it manually", async () => {
    render(<OnboardPage />);
    const user = userEvent.setup();

    const businessNameInput = screen.getByLabelText("Business name");
    const storeLinkInput = screen.getByLabelText("Store link");

    await user.type(businessNameInput, "Amina Fashion House");
    expect(storeLinkInput).toHaveValue("amina-fashion-house");

    await user.clear(storeLinkInput);
    await user.type(storeLinkInput, "amina style");
    await user.clear(businessNameInput);
    await user.type(businessNameInput, "Amina Luxe Atelier");

    expect(storeLinkInput).toHaveValue("amina-style");
    expect(screen.getByText("storefront.com/")).toBeInTheDocument();
    expect(screen.getByText("amina-style")).toBeInTheDocument();
  });

  it("shows the success state and waits before redirecting after onboarding", async () => {
    mutateAsyncMock.mockResolvedValue({ name: "Amina Fashion House" });
    vi.useFakeTimers();

    render(<OnboardPage />);
    fireEvent.change(screen.getByLabelText("Business name"), {
      target: { value: "Amina Fashion House" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Create my store" }));

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

    expect(replaceMock).toHaveBeenCalledWith("/app");
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

  it("lets the user sign out from onboarding", async () => {
    render(<OnboardPage />);
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: "Sign out" }));

    expect(signOutMock).toHaveBeenCalledOnce();
  });
});
