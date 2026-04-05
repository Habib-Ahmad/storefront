import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import SettingsPage from "@/app/app/settings/page";

beforeEach(() => {
  vi.clearAllMocks();
});

describe("SettingsPage", () => {
  it("points the user to the dedicated storefront workspace", async () => {
    render(<SettingsPage />);

    expect(screen.getByText("Storefront moved")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Open Storefront" })).toHaveAttribute(
      "href",
      "/app/storefront",
    );
  });
});
