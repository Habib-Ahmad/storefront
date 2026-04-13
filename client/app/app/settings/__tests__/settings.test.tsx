import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import SettingsPage from "@/app/app/settings/page";

describe("SettingsPage", () => {
  it("links to logistics setup and storefront controls", () => {
    render(<SettingsPage />);

    expect(screen.getByRole("link", { name: /open logistics setup/i })).toHaveAttribute(
      "href",
      "/app/settings/logistics",
    );
    expect(screen.getByRole("link", { name: /open storefront controls/i })).toHaveAttribute(
      "href",
      "/app/storefront",
    );
  });
});
