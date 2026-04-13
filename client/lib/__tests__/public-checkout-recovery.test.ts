import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  basketRecoveryKey,
  clearPendingOrderByTrackingSlug,
  getLatestPendingOrderForStorefront,
  getOrCreateCheckoutId,
  keepPendingOrderByTrackingSlug,
  rememberPendingOrder,
} from "../public-checkout-recovery";

const slug = "funke-fabrics";
const recoveryKey = basketRecoveryKey(slug);

beforeEach(() => {
  vi.useRealTimers();
  window.localStorage.clear();
});

describe("public checkout recovery", () => {
  it("reuses the same checkout id for the same recovery context", () => {
    const first = getOrCreateCheckoutId(recoveryKey, slug);
    const second = getOrCreateCheckoutId(recoveryKey, slug);

    expect(second).toBe(first);
  });

  it("returns the latest pending order for a storefront", () => {
    rememberPendingOrder(recoveryKey, slug, "track-old");
    rememberPendingOrder("basket:other-store", "other-store", "track-other");
    rememberPendingOrder(recoveryKey, slug, "track-new");

    expect(getLatestPendingOrderForStorefront(slug)).toEqual(
      expect.objectContaining({
        storefrontSlug: slug,
        trackingSlug: "track-new",
        orderPath: "/order/track-new",
      }),
    );
  });

  it("drops recovery records once the pending order is cleared", () => {
    rememberPendingOrder(recoveryKey, slug, "track-123");

    clearPendingOrderByTrackingSlug("track-123");

    expect(getLatestPendingOrderForStorefront(slug)).toBeNull();
  });

  it("refreshes the pending order timestamp when it is kept alive", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2025-01-01T00:00:00.000Z"));

    rememberPendingOrder(recoveryKey, slug, "track-123");

    vi.setSystemTime(new Date("2025-01-01T00:44:00.000Z"));
    keepPendingOrderByTrackingSlug("track-123");

    vi.setSystemTime(new Date("2025-01-01T00:46:00.000Z"));

    expect(getLatestPendingOrderForStorefront(slug)).toEqual(
      expect.objectContaining({ trackingSlug: "track-123" }),
    );
  });
});
