import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const pushMock = vi.fn();
const refreshMock = vi.fn();
const mockUseParams = vi.fn();
const mockCancelOrderMutate = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: pushMock,
    replace: vi.fn(),
    refresh: refreshMock,
    back: vi.fn(),
  }),
  usePathname: () => "/app/orders",
  useParams: () => mockUseParams(),
}));

const mockUseOrders = vi.fn();
const mockUseOrder = vi.fn();
const mockUseOrderItems = vi.fn();
const mockUseCreateOrder = vi.fn();
const mockUseCancelOrder = vi.fn();
const mockUseProducts = vi.fn();
const mockUseVariants = vi.fn();

vi.mock("@/hooks/use-orders", () => ({
  useOrders: (...args: unknown[]) => mockUseOrders(...args),
  useOrder: (...args: unknown[]) => mockUseOrder(...args),
  useOrderItems: (...args: unknown[]) => mockUseOrderItems(...args),
  useCreateOrder: () => ({
    mutateAsync: mockUseCreateOrder,
    isPending: false,
  }),
  useCancelOrder: () => mockUseCancelOrder(),
}));

vi.mock("@/hooks/use-products", () => ({
  useProducts: (...args: unknown[]) => mockUseProducts(...args),
  useVariants: (...args: unknown[]) => mockUseVariants(...args),
}));

import OrdersPage from "@/app/app/orders/page";
import OrderDetailPage from "@/app/app/orders/[id]/page";
import NewOrderPage from "@/app/app/orders/new/page";

beforeEach(() => {
  vi.clearAllMocks();
  mockUseParams.mockReturnValue({ id: "order-1" });
  mockUseOrders.mockReturnValue({
    data: { data: [], total: 0, page: 1, per_page: 12 },
    isLoading: false,
  });
  mockUseOrder.mockReturnValue({
    data: null,
    isLoading: false,
  });
  mockUseOrderItems.mockReturnValue({
    data: [],
    isLoading: false,
  });
  mockUseProducts.mockReturnValue({
    data: { data: [], total: 0, page: 1, per_page: 100 },
    isLoading: false,
  });
  mockUseVariants.mockReturnValue({
    data: [],
    isLoading: false,
  });
  mockUseCancelOrder.mockReturnValue({
    mutateAsync: mockCancelOrderMutate,
    isPending: false,
  });
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe("OrdersPage", () => {
  it("shows empty state when there are no orders", () => {
    render(<OrdersPage />);

    expect(screen.getByText("Your orders will appear here")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /create your first order/i })).toHaveAttribute(
      "href",
      "/app/orders/new",
    );
  });

  it("renders order cards with customer, statuses, total, and link", () => {
    mockUseOrders.mockReturnValue({
      data: {
        data: [
          {
            id: "order-1",
            tenant_id: "tenant-1",
            tracking_slug: "abc123def456",
            is_delivery: true,
            customer_name: "Amina Bello",
            customer_phone: "+2348012345678",
            customer_email: "amina@example.com",
            shipping_address: "12 Allen Avenue, Ikeja",
            note: "Handle with care",
            total_amount: "18500",
            shipping_fee: "1500",
            payment_method: "cash",
            payment_status: "paid",
            fulfillment_status: "processing",
            created_at: "2026-03-14T10:00:00Z",
            updated_at: "2026-03-14T10:00:00Z",
          },
          {
            id: "order-2",
            tenant_id: "tenant-1",
            tracking_slug: "xyz987uvw654",
            is_delivery: false,
            customer_name: null,
            customer_phone: null,
            customer_email: null,
            shipping_address: null,
            note: null,
            total_amount: "7200",
            shipping_fee: "0",
            payment_method: "cash",
            payment_status: "paid",
            fulfillment_status: "completed",
            created_at: "2026-03-12T10:00:00Z",
            updated_at: "2026-03-12T10:00:00Z",
          },
          {
            id: "order-3",
            tenant_id: "tenant-1",
            tracking_slug: "pqr765stu321",
            is_delivery: true,
            customer_name: "Kehinde Musa",
            customer_phone: "+2348099999999",
            customer_email: null,
            shipping_address: "8 Admiralty Way, Lekki",
            note: null,
            total_amount: "12000",
            shipping_fee: "1000",
            payment_method: "online",
            payment_status: "pending",
            fulfillment_status: "cancelled",
            created_at: "2026-03-11T10:00:00Z",
            updated_at: "2026-03-11T10:00:00Z",
          },
        ],
        total: 3,
        page: 1,
        per_page: 12,
      },
      isLoading: false,
    });

    render(<OrdersPage />);

    expect(screen.getByText("Amina Bello")).toBeInTheDocument();
    expect(screen.getByText("Walk-in customer")).toBeInTheDocument();
    expect(screen.getByText("Kehinde Musa")).toBeInTheDocument();
    expect(screen.getByText(/14 Mar 2026, 11:00 AM/i)).toBeInTheDocument();
    expect(screen.getByText(/12 Mar 2026, 11:00 AM/i)).toBeInTheDocument();
    expect(screen.getByText(/11 Mar 2026, 11:00 AM/i)).toBeInTheDocument();
    expect(screen.queryByText("abc123def456")).not.toBeInTheDocument();
    expect(screen.queryByText("xyz987uvw654")).not.toBeInTheDocument();
    expect(screen.getAllByText("Ready for delivery").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Completed").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Cancelled").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Delivery").length).toBeGreaterThan(0);
    expect(screen.queryByText("Pickup")).not.toBeInTheDocument();
    expect(screen.queryByText("Awaiting payment")).not.toBeInTheDocument();
    expect(screen.queryByText("Paid")).not.toBeInTheDocument();

    const firstLink = screen.getByText("Amina Bello").closest("a");
    const secondLink = screen.getByText("Walk-in customer").closest("a");
    const thirdLink = screen.getByText("Kehinde Musa").closest("a");

    expect(firstLink).toHaveAttribute("href", "/app/orders/order-1");
    expect(secondLink).toHaveAttribute("href", "/app/orders/order-2");
    expect(thirdLink).toHaveAttribute("href", "/app/orders/order-3");
  });

  it("shows pagination controls when there are multiple pages", async () => {
    mockUseOrders.mockReturnValue({
      data: {
        data: [
          {
            id: "order-1",
            tenant_id: "tenant-1",
            tracking_slug: "abc123def456",
            is_delivery: false,
            customer_name: "Amina Bello",
            customer_phone: null,
            customer_email: null,
            shipping_address: null,
            note: null,
            total_amount: "18500",
            shipping_fee: "0",
            payment_method: "cash",
            payment_status: "paid",
            fulfillment_status: "processing",
            created_at: "2026-03-14T10:00:00Z",
            updated_at: "2026-03-14T10:00:00Z",
          },
        ],
        total: 24,
        page: 1,
        per_page: 12,
      },
      isLoading: false,
    });

    render(<OrdersPage />);

    expect(screen.getByText("1 / 2")).toBeInTheDocument();

    const buttons = screen.getAllByRole("button");
    const prevButton = buttons[1];
    const nextButton = buttons[2];

    expect(prevButton).toBeDisabled();
    expect(nextButton).not.toBeDisabled();

    await userEvent.click(nextButton);

    expect(mockUseOrders).toHaveBeenLastCalledWith({ page: 2, per_page: 12 });
  });
});

describe("NewOrderPage", () => {
  const productsPayload = {
    data: {
      data: [
        {
          id: "product-1",
          tenant_id: "tenant-1",
          name: "Ankara Shirt",
          description: "Bright patterned shirt",
          category: "Fashion",
          is_available: true,
          created_at: "2026-03-14T10:00:00Z",
          updated_at: "2026-03-14T10:00:00Z",
          images: [],
        },
      ],
      total: 1,
      page: 1,
      per_page: 100,
    },
    isLoading: false,
  };

  it("shows validation errors when submitting an incomplete catalog order", async () => {
    mockUseProducts.mockReturnValue(productsPayload);
    mockUseVariants.mockReturnValue({
      data: [],
      isLoading: false,
    });

    render(<NewOrderPage />);

    await userEvent.click(screen.getByRole("button", { name: /choose products/i }));
    await userEvent.click(screen.getByRole("button", { name: /save order/i }));

    expect(
      await screen.findByText((content, element) => {
        return element?.tagName.toLowerCase() === "p" && content === "Choose a product";
      }),
    ).toBeInTheDocument();
  });

  it("creates a vendor-entered catalog order and returns to the order list", async () => {
    mockUseProducts.mockReturnValue(productsPayload);
    mockUseVariants.mockImplementation((productId: string) => ({
      data:
        productId === "product-1"
          ? [
              {
                id: "variant-1",
                product_id: "product-1",
                sku: "Standard",
                attributes: {},
                price: "5000",
                cost_price: "3000",
                stock_qty: 10,
                is_default: true,
                created_at: "2026-03-14T10:00:00Z",
                updated_at: "2026-03-14T10:00:00Z",
              },
            ]
          : [],
      isLoading: false,
    }));
    mockUseCreateOrder.mockResolvedValue({
      id: "new-order-id",
      payment_method: "cash",
    });

    render(<NewOrderPage />);

    await userEvent.click(screen.getByRole("button", { name: /choose products/i }));
    await userEvent.selectOptions(screen.getByLabelText("Product"), "product-1");

    await waitFor(() => {
      expect(screen.getByText("Option added automatically")).toBeInTheDocument();
    });

    expect(screen.queryByLabelText("Option")).not.toBeInTheDocument();

    await userEvent.clear(screen.getByLabelText("Quantity"));
    await userEvent.type(screen.getByLabelText("Quantity"), "2");
    await userEvent.click(screen.getByRole("button", { name: /add details/i }));
    await userEvent.type(screen.getByLabelText("Customer name"), "Amina Bello");
    await userEvent.click(screen.getByRole("button", { name: /^cash$/i }));
    await userEvent.click(screen.getByRole("button", { name: /save order/i }));

    await waitFor(() => {
      expect(mockUseCreateOrder).toHaveBeenCalledWith(
        expect.objectContaining({
          is_delivery: false,
          payment_method: "cash",
          customer_name: "Amina Bello",
          items: [{ variant_id: "variant-1", quantity: 2 }],
        }),
      );
    });

    expect(pushMock).toHaveBeenCalledWith("/app/orders");
  });

  it("shows payment details for quick orders", async () => {
    mockUseProducts.mockReturnValue(productsPayload);
    mockUseVariants.mockReturnValue({
      data: [],
      isLoading: false,
    });

    render(<NewOrderPage />);

    await userEvent.click(screen.getByRole("button", { name: /quick order/i }));

    expect(screen.getByRole("heading", { name: /^payment$/i })).toBeInTheDocument();
    expect(screen.getByLabelText("Amount (₦)")).toBeInTheDocument();
    expect(screen.getByLabelText("Payment method")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^cash$/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^transfer$/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^online$/i })).toBeInTheDocument();
  });

  it("keeps variants isolated per item when adding another product line", async () => {
    mockUseProducts.mockReturnValue(productsPayload);
    mockUseVariants.mockImplementation((productId: string) => ({
      data:
        productId === "product-1"
          ? [
              {
                id: "variant-1",
                product_id: "product-1",
                sku: "Standard",
                attributes: {},
                price: "5000",
                cost_price: "3000",
                stock_qty: 10,
                is_default: true,
                created_at: "2026-03-14T10:00:00Z",
                updated_at: "2026-03-14T10:00:00Z",
              },
            ]
          : [],
      isLoading: false,
    }));

    render(<NewOrderPage />);

    await userEvent.click(screen.getByRole("button", { name: /choose products/i }));
    await userEvent.selectOptions(screen.getByLabelText("Product"), "product-1");

    await waitFor(() => {
      expect(screen.getByText("Option added automatically")).toBeInTheDocument();
    });

    await userEvent.click(screen.getByRole("button", { name: /^add item$/i }));

    const productSelects = screen.getAllByLabelText("Product");

    await userEvent.selectOptions(productSelects[1], "product-1");

    await waitFor(() => {
      expect(screen.getAllByText("Option added automatically")).toHaveLength(2);
    });
  });

  it("does not show delivery fields by default", () => {
    mockUseProducts.mockReturnValue(productsPayload);
    mockUseVariants.mockReturnValue({
      data: [],
      isLoading: false,
    });

    render(<NewOrderPage />);

    expect(screen.getByRole("button", { name: /^no delivery$/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^add delivery$/i })).toBeInTheDocument();
    expect(screen.queryByLabelText("Shipping fee (₦)")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("Address")).not.toBeInTheDocument();
  });

  it("lets you ignore delivery after turning it on", async () => {
    mockUseProducts.mockReturnValue(productsPayload);
    mockUseVariants.mockReturnValue({
      data: [],
      isLoading: false,
    });

    render(<NewOrderPage />);

    await userEvent.click(screen.getByRole("button", { name: /^add delivery$/i }));

    expect(screen.getByLabelText("Shipping fee (₦)")).toBeInTheDocument();
    expect(screen.getByLabelText("Address")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: /^no delivery$/i }));

    expect(screen.queryByLabelText("Shipping fee (₦)")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("Address")).not.toBeInTheDocument();
  });

  it("supports quick sale orders", async () => {
    mockUseProducts.mockReturnValue(productsPayload);
    mockUseVariants.mockReturnValue({
      data: [],
      isLoading: false,
    });
    mockUseCreateOrder.mockResolvedValue({
      id: "quick-sale-order",
      payment_method: "cash",
    });

    render(<NewOrderPage />);

    await userEvent.click(screen.getByRole("button", { name: /quick order/i }));
    await userEvent.type(screen.getByLabelText("Amount (₦)"), "7500");
    await userEvent.click(screen.getByRole("button", { name: /^transfer$/i }));

    await userEvent.click(screen.getByRole("button", { name: /save order/i }));

    await waitFor(() => {
      expect(mockUseCreateOrder).toHaveBeenCalledWith(
        expect.objectContaining({
          payment_method: "transfer",
          total_amount: 7500,
          items: [],
        }),
      );
    });

    expect(pushMock).toHaveBeenCalledWith("/app/orders");
  });

  it("redirects to the authorization URL for online payments", async () => {
    mockUseProducts.mockReturnValue(productsPayload);
    mockUseVariants.mockReturnValue({
      data: [],
      isLoading: false,
    });
    mockUseCreateOrder.mockResolvedValue({
      id: "online-order",
      payment_method: "online",
      authorization_url: "https://paystack.test/authorize",
    });

    const assignMock = vi.fn();
    Object.defineProperty(window, "location", {
      value: {
        ...window.location,
        href: "",
        assign: assignMock,
      },
      writable: true,
      configurable: true,
    });

    render(<NewOrderPage />);

    await userEvent.click(screen.getByRole("button", { name: /quick order/i }));
    await userEvent.type(screen.getByLabelText("Amount (₦)"), "12000");
    await userEvent.click(screen.getByRole("button", { name: /^online$/i }));

    await userEvent.click(screen.getByRole("button", { name: /continue to payment/i }));

    await waitFor(() => {
      expect(mockUseCreateOrder).toHaveBeenCalled();
    });

    expect(window.location.href).toBe("https://paystack.test/authorize");
  });

  it("shows an API error when order creation fails", async () => {
    mockUseProducts.mockReturnValue(productsPayload);
    mockUseVariants.mockReturnValue({
      data: [],
      isLoading: false,
    });
    const { ApiError } = await import("@/lib/api");
    mockUseCreateOrder.mockRejectedValue(new ApiError(422, "inventory module not enabled"));

    render(<NewOrderPage />);

    await userEvent.click(screen.getByRole("button", { name: /quick order/i }));
    await userEvent.type(screen.getByLabelText("Amount (₦)"), "9000");

    await userEvent.click(screen.getByRole("button", { name: /save order/i }));

    expect(await screen.findByText("inventory module not enabled")).toBeInTheDocument();
  });

  it("shows helpful empty-state guidance when there are no products for catalog orders", async () => {
    mockUseProducts.mockReturnValue({
      data: { data: [], total: 0, page: 1, per_page: 100 },
      isLoading: false,
    });
    mockUseVariants.mockReturnValue({
      data: [],
      isLoading: false,
    });

    render(<NewOrderPage />);

    await userEvent.click(screen.getByRole("button", { name: /choose products/i }));

    expect(
      screen.getByText(/You don’t have any products yet. Add products first or use quick order./i),
    ).toBeInTheDocument();
  });
});

describe("OrderDetailPage", () => {
  const baseOrder = {
    id: "order-1",
    tenant_id: "tenant-1",
    tracking_slug: "abc123def456",
    customer_name: "Amina Bello",
    customer_phone: "+2348012345678",
    customer_email: "amina@example.com",
    note: "Handle with care",
    total_amount: "18500",
    shipping_fee: "0",
    payment_method: "cash",
    payment_status: "paid",
    fulfillment_status: "completed",
    created_at: "2026-03-14T10:00:00Z",
    updated_at: "2026-03-14T10:00:00Z",
  };

  it("shows completed pickup orders without actions", () => {
    mockUseOrder.mockReturnValue({
      data: {
        ...baseOrder,
        is_delivery: false,
        shipping_address: null,
      },
      isLoading: false,
    });
    mockUseOrderItems.mockReturnValue({
      data: [],
      isLoading: false,
    });

    render(<OrderDetailPage />);

    expect(screen.getAllByText("No delivery").length).toBeGreaterThan(0);
    expect(
      screen.getByText("This was saved as a quick order. No items were added."),
    ).toBeInTheDocument();
    expect(screen.getAllByText("completed").length).toBeGreaterThan(0);
    expect(screen.queryByRole("button", { name: /dispatch order/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /cancel order/i })).not.toBeInTheDocument();
    expect(screen.getByText("This pickup order is complete.")).toBeInTheDocument();
  });

  it("shows delivery details for delivery orders", () => {
    mockUseOrder.mockReturnValue({
      data: {
        ...baseOrder,
        is_delivery: true,
        fulfillment_status: "processing",
        shipping_address: "12 Allen Avenue, Ikeja",
        shipping_fee: "1500",
      },
      isLoading: false,
    });
    mockUseOrderItems.mockReturnValue({
      data: [],
      isLoading: false,
    });

    render(<OrderDetailPage />);

    expect(screen.getByText("Delivery details for this order.")).toBeInTheDocument();
    expect(screen.getByText("12 Allen Avenue, Ikeja")).toBeInTheDocument();
    expect(screen.getByText(/dispatch setup is not ready in this screen yet/i)).toBeInTheDocument();
  });
});
