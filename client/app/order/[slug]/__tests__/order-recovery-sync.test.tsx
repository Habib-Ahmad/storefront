import { render } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { OrderRecoverySync } from "@/app/order/[slug]/order-recovery-sync";

const clearPendingOrderByTrackingSlug = vi.fn();
const keepPendingOrderByTrackingSlug = vi.fn();

vi.mock("@/lib/public-checkout-recovery", () => ({
  clearPendingOrderByTrackingSlug: (...args: unknown[]) => clearPendingOrderByTrackingSlug(...args),
  keepPendingOrderByTrackingSlug: (...args: unknown[]) => keepPendingOrderByTrackingSlug(...args),
}));

describe("OrderRecoverySync", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("keeps recovery active while the order can still resume payment", () => {
    render(
      <OrderRecoverySync
        trackingSlug="track-123"
        paymentStatus="pending"
        fulfillmentStatus="processing"
      />,
    );

    expect(keepPendingOrderByTrackingSlug).toHaveBeenCalledWith("track-123");
    expect(clearPendingOrderByTrackingSlug).not.toHaveBeenCalled();
  });

  it("clears recovery once the order leaves the resumable state", () => {
    const { rerender } = render(
      <OrderRecoverySync
        trackingSlug="track-123"
        paymentStatus="pending"
        fulfillmentStatus="processing"
      />,
    );

    rerender(
      <OrderRecoverySync
        trackingSlug="track-123"
        paymentStatus="paid"
        fulfillmentStatus="processing"
      />,
    );

    expect(clearPendingOrderByTrackingSlug).toHaveBeenCalledWith("track-123");
  });
});
