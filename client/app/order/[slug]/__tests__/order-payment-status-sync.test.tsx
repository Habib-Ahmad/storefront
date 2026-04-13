import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { OrderPaymentStatusSync } from "@/app/order/[slug]/order-payment-status-sync";

const replaceMock = vi.fn();
const refreshMock = vi.fn();
const mockSearchParams = vi.fn();
const confirmTrackedOrderPaymentMock = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    replace: replaceMock,
    refresh: refreshMock,
  }),
  usePathname: () => "/order/track-123",
  useSearchParams: () => mockSearchParams(),
}));

vi.mock("@/lib/api", () => ({
  api: {
    confirmTrackedOrderPayment: (...args: unknown[]) => confirmTrackedOrderPaymentMock(...args),
  },
  ApiError: class ApiError extends Error {
    status: number;
    constructor(status: number, message: string) {
      super(message);
      this.status = status;
    }
  },
}));

describe("OrderPaymentStatusSync", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSearchParams.mockReturnValue(new URLSearchParams("reference=abc&trxref=def"));
  });

  it("confirms the returned payment and refreshes when the payment is now paid", async () => {
    confirmTrackedOrderPaymentMock.mockResolvedValue({
      tracking_slug: "track-123",
      payment_status: "paid",
      fulfillment_status: "processing",
    });

    render(
      <OrderPaymentStatusSync
        slug="track-123"
        paymentStatus="pending"
        reference="abc"
        trxref="def"
      />,
    );

    expect(screen.getByText("Confirming payment")).toBeInTheDocument();

    await waitFor(() => {
      expect(confirmTrackedOrderPaymentMock).toHaveBeenCalledWith("track-123", {
        reference: "abc",
        trxref: "def",
      });
    });

    await waitFor(() => {
      expect(replaceMock).toHaveBeenCalledWith("/order/track-123", { scroll: false });
      expect(refreshMock).toHaveBeenCalled();
    });
  });

  it("does not render when there is no returned payment reference", () => {
    const { container } = render(
      <OrderPaymentStatusSync slug="track-123" paymentStatus="pending" />,
    );

    expect(container).toBeEmptyDOMElement();
    expect(confirmTrackedOrderPaymentMock).not.toHaveBeenCalled();
  });
});
