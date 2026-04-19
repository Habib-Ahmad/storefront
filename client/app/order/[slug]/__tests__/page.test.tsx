import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import OrderSummaryPage from "@/app/order/[slug]/page";

const trackMock = vi.fn();
const notFoundMock = vi.fn();

vi.mock("next/navigation", () => ({
  notFound: () => notFoundMock(),
}));

vi.mock("@/lib/api", () => ({
  api: {
    track: (...args: unknown[]) => trackMock(...args),
  },
  ApiError: class ApiError extends Error {
    status: number;
    constructor(status: number, message: string) {
      super(message);
      this.status = status;
    }
  },
}));

vi.mock("@/components/public-storefront-actions", () => ({
  PublicStorefrontActions: () => <div>storefront actions</div>,
}));

vi.mock("@/app/order/[slug]/order-recovery-sync", () => ({
  OrderRecoverySync: () => null,
}));

vi.mock("@/app/order/[slug]/order-payment-status-sync", () => ({
  OrderPaymentStatusSync: () => null,
}));

vi.mock("@/app/order/[slug]/resume-payment-button", () => ({
  ResumePaymentButton: () => <button type="button">Continue payment</button>,
}));

describe("OrderSummaryPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("simplifies pickup confirmations and offers a way back to the store", async () => {
    trackMock.mockResolvedValue({
      tracking_slug: "track-123",
      is_delivery: false,
      storefront_slug: "funke-fabrics",
      payment_status: "paid",
      fulfillment_status: "processing",
    });

    render(
      await OrderSummaryPage({
        params: Promise.resolve({ slug: "track-123" }),
        searchParams: Promise.resolve({}),
      }),
    );

    expect(screen.getByText("Payment confirmed")).toBeInTheDocument();
    expect(
      screen.getByText(
        "Payment is complete. The store has your pickup order and will take it from here.",
      ),
    ).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /continue shopping/i })).toHaveAttribute(
      "href",
      "/funke-fabrics",
    );
    expect(screen.queryByText("Tracking code")).not.toBeInTheDocument();
    expect(screen.queryByText("Delivery updates")).not.toBeInTheDocument();
    expect(screen.queryByText("Order progress")).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /track order/i })).not.toBeInTheDocument();
  });

  it("keeps delivery updates visible for delivery orders", async () => {
    trackMock.mockResolvedValue({
      tracking_slug: "track-123",
      is_delivery: true,
      storefront_slug: "funke-fabrics",
      payment_status: "pending",
      fulfillment_status: "processing",
    });

    render(
      await OrderSummaryPage({
        params: Promise.resolve({ slug: "track-123" }),
        searchParams: Promise.resolve({}),
      }),
    );

    expect(screen.getByText("Tracking code")).toBeInTheDocument();
    expect(screen.getByText("Delivery updates")).toBeInTheDocument();
    expect(screen.getByText("Order progress")).toBeInTheDocument();
  });
});
