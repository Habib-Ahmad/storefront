import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
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

describe("auth flows", () => {
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
    render(<SignupPage />);
    const user = userEvent.setup();

    await user.type(screen.getByLabelText("Email"), "owner@example.com");
    await user.type(screen.getByLabelText("Password"), "password123");
    await user.type(screen.getByLabelText("Confirm password"), "password123");
    await user.click(screen.getByRole("button", { name: "Create account" }));

    await new Promise((resolve) => window.setTimeout(resolve, 2100));

    expect(pushMock).toHaveBeenCalledWith("/onboard");
  }, 10000);
});
