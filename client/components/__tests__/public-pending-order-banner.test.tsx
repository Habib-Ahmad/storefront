import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { PublicPendingOrderBanner } from "@/components/public-pending-order-banner";

const mockGetLatestPendingOrderForStorefront = vi.fn();

vi.mock("@/lib/public-checkout-recovery", () => ({
  getLatestPendingOrderForStorefront: (...args: unknown[]) =>
    mockGetLatestPendingOrderForStorefront(...args),
}));

describe("PublicPendingOrderBanner", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("shows a recovery CTA when an unfinished order exists", async () => {
    mockGetLatestPendingOrderForStorefront.mockReturnValue({
      trackingSlug: "track-123",
      orderPath: "/order/track-123",
    });

    render(<PublicPendingOrderBanner storefrontSlug="funke-fabrics" />);

    await waitFor(() => {
      expect(mockGetLatestPendingOrderForStorefront).toHaveBeenCalledWith("funke-fabrics");
    });

    expect(screen.getByText("You have an unfinished order.")).toBeInTheDocument();
    expect(
      screen.getByText("Open your order to continue payment without starting over."),
    ).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Continue order" })).toHaveAttribute(
      "href",
      "/order/track-123",
    );
  });

  it("stays hidden when there is no unfinished order to recover", async () => {
    mockGetLatestPendingOrderForStorefront.mockReturnValue(null);

    render(<PublicPendingOrderBanner storefrontSlug="funke-fabrics" />);

    await waitFor(() => {
      expect(mockGetLatestPendingOrderForStorefront).toHaveBeenCalledWith("funke-fabrics");
    });

    expect(screen.queryByText("You have an unfinished order.")).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Continue order" })).not.toBeInTheDocument();
  });
});
