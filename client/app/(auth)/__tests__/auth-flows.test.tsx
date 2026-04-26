import type { ReactElement } from "react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { act } from "@testing-library/react";
import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import LoginPage from "@/app/(auth)/login/page";
import SignupPage from "@/app/(auth)/signup/page";

const {
  replaceMock,
  pushMock,
  searchParamsGetMock,
  signInWithPasswordMock,
  signUpMock,
  signInWithOAuthMock,
  getMeMock,
} = vi.hoisted(() => ({
  replaceMock: vi.fn(),
  pushMock: vi.fn(),
  searchParamsGetMock: vi.fn(),
  signInWithPasswordMock: vi.fn(),
  signUpMock: vi.fn(),
  signInWithOAuthMock: vi.fn(),
  getMeMock: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: replaceMock, push: pushMock }),
  useSearchParams: () => ({ get: searchParamsGetMock }),
}));

vi.mock("@/lib/supabase", () => ({
  getSupabase: () => ({
    auth: {
      signInWithPassword: signInWithPasswordMock,
      signUp: signUpMock,
      signInWithOAuth: signInWithOAuthMock,
    },
  }),
}));

vi.mock("@/lib/api", async () => {
  const actual = await vi.importActual<typeof import("@/lib/api")>("@/lib/api");
  return {
    ...actual,
    api: {
      ...actual.api,
      getMe: getMeMock,
    },
  };
});

beforeEach(() => {
  vi.clearAllMocks();
  searchParamsGetMock.mockReturnValue(null);
  signInWithPasswordMock.mockResolvedValue({ error: null });
  signUpMock.mockResolvedValue({ data: { session: { access_token: "token" } }, error: null });
  getMeMock.mockResolvedValue({ onboarded: true });
});

afterEach(() => {
  vi.useRealTimers();
});

async function startGoogleOAuth(page: ReactElement) {
  render(page);
  const user = userEvent.setup();
  await user.click(screen.getByRole("button", { name: "Google" }));
}

describe("auth flows", () => {
  it("preserves the requested app redirect when login starts Google OAuth", async () => {
    searchParamsGetMock.mockImplementation((key: string) =>
      key === "redirect" ? "/app/orders" : null,
    );

    await startGoogleOAuth(<LoginPage />);

    expect(signInWithOAuthMock).toHaveBeenCalledWith({
      provider: "google",
      options: {
        redirectTo: `${window.location.origin}/auth/callback?next=${encodeURIComponent("/app/orders")}`,
      },
    });
  });

  it("shows a clear message when the OAuth callback redirects back with an error", () => {
    searchParamsGetMock.mockImplementation((key: string) =>
      key === "error" ? "oauth_callback" : null,
    );

    render(<LoginPage />);

    expect(
      screen.getByText("Google sign-in could not be completed. Please try again."),
    ).toBeInTheDocument();
  });

  it("sends first-time Google signup straight to onboarding", async () => {
    await startGoogleOAuth(<SignupPage />);

    expect(signInWithOAuthMock).toHaveBeenCalledWith({
      provider: "google",
      options: {
        redirectTo: `${window.location.origin}/auth/callback?next=${encodeURIComponent("/onboard")}`,
      },
    });
  });

  it("tells an existing user to sign in when they try to sign up again", async () => {
    signUpMock.mockResolvedValue({
      data: { session: null },
      error: { message: "User already registered" },
    });

    render(<SignupPage />);
    const user = userEvent.setup();

    await user.type(screen.getByLabelText("Email"), "owner@example.com");
    await user.type(screen.getByLabelText("Password"), "password123");
    await user.type(screen.getByLabelText("Confirm password"), "password123");
    await user.click(screen.getByRole("button", { name: "Create account" }));

    expect(
      await screen.findByText("You already have an account. Sign in to continue."),
    ).toBeInTheDocument();
    expect(
      screen.getByText("Sign in instead and continue setting up your store."),
    ).toBeInTheDocument();
  });

  it("redirects a user with incomplete onboarding back to onboarding after login", async () => {
    getMeMock.mockResolvedValue({ onboarded: false });

    render(<LoginPage />);
    const user = userEvent.setup();

    await user.type(screen.getByLabelText("Email"), "owner@example.com");
    await user.type(screen.getByLabelText("Password"), "password123");
    await user.click(screen.getByRole("button", { name: "Sign in" }));

    expect(await signInWithPasswordMock).toHaveBeenCalledWith({
      email: "owner@example.com",
      password: "password123",
    });
    expect(replaceMock).toHaveBeenCalledWith("/onboard");
  });

  it("redirects straight to onboarding after signup when a session is created", async () => {
    vi.useFakeTimers();
    render(<SignupPage />);

    fireEvent.change(screen.getByLabelText("Email"), {
      target: { value: "owner@example.com" },
    });
    fireEvent.change(screen.getByLabelText("Password"), {
      target: { value: "password123" },
    });
    fireEvent.change(screen.getByLabelText("Confirm password"), {
      target: { value: "password123" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Create account" }));

    await act(async () => {
      await Promise.resolve();
      await vi.advanceTimersByTimeAsync(2000);
    });

    expect(pushMock).toHaveBeenCalledWith("/onboard");
  }, 10000);
});
