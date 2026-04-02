import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const pushMock = vi.fn();
const refreshMock = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: pushMock,
    replace: vi.fn(),
    refresh: refreshMock,
    back: vi.fn(),
  }),
  usePathname: () => "/app/orders",
  useParams: () => ({}),
}));

const mockUseOrders = vi.fn();
const mockUseCreateOrder = vi.fn();
const mockUseProducts = vi.fn();
const mockUseVariants = vi.fn();

vi.mock("@/hooks/use-orders", () => ({
  useOrders: (...args: unknown[]) => mockUseOrders(...args),
  useCreateOrder: () => ({
    mutateAsync: mockUseCreateOrder,
    isPending: false,
  }),
}));

vi.mock("@/hooks/use-products", () => ({
  useProducts: (...args: unknown[]) => mockUseProducts(...args),
  useVariants: (...args: unknown[]) => mockUseVariants(...args),
}));

import OrdersPage from "@/app/app/orders/page";
import NewOrderPage from "@/app/app/orders/new/page";

beforeEach(() => {
  vi.clearAllMocks();
  mockUseOrders.mockReturnValue({
    data: { data: [], total: 0, page: 1, per_page: 12 },
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
            payment_method: "online",
            payment_status: "pending",
            fulfillment_status: "cancelled",
            created_at: "2026-03-12T10:00:00Z",
            updated_at: "2026-03-12T10:00:00Z",
          },
        ],
        total: 2,
        page: 1,
        per_page: 12,
      },
      isLoading: false,
    });

    render(<OrdersPage />);

    expect(screen.getByText("Amina Bello")).toBeInTheDocument();
    expect(screen.getByText("Walk-in customer")).toBeInTheDocument();
    expect(screen.getByText("abc123def456")).toBeInTheDocument();
    expect(screen.getByText("xyz987uvw654")).toBeInTheDocument();
    expect(screen.getAllByText("paid").length).toBeGreaterThan(0);
    expect(screen.getAllByText("pending").length).toBeGreaterThan(0);
    expect(screen.getAllByText("processing").length).toBeGreaterThan(0);
    expect(screen.getAllByText("cancelled").length).toBeGreaterThan(0);
    expect(screen.getByText("Delivery")).toBeInTheDocument();

    const firstLink = screen.getByText("Amina Bello").closest("a");
    const secondLink = screen.getByText("Walk-in customer").closest("a");

    expect(firstLink).toHaveAttribute("href", "/app/orders/order-1");
    expect(secondLink).toHaveAttribute("href", "/app/orders/order-2");
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

  it("creates a vendor-entered catalog order and navigates to its detail page", async () => {
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
      expect(screen.getByLabelText("Option")).toHaveValue("variant-1");
    });

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

    expect(pushMock).toHaveBeenCalledWith("/app/orders/new-order-id");
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
    expect(
      screen.getByText("Enter the amount and choose how the customer will pay."),
    ).toBeInTheDocument();
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
      expect(screen.getByLabelText("Option")).toHaveValue("variant-1");
    });

    await userEvent.click(screen.getByRole("button", { name: /add item/i }));

    const productSelects = screen.getAllByLabelText("Product");
    const optionSelects = screen.getAllByLabelText("Option");

    await userEvent.selectOptions(productSelects[1], "product-1");

    await waitFor(() => {
      expect(optionSelects[0]).toHaveValue("variant-1");
      expect(optionSelects[1]).toHaveValue("variant-1");
    });
  });

  it("does not show delivery fields by default", () => {
    mockUseProducts.mockReturnValue(productsPayload);
    mockUseVariants.mockReturnValue({
      data: [],
      isLoading: false,
    });

    render(<NewOrderPage />);

    expect(screen.getByText("No delivery")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /add delivery/i })).toBeInTheDocument();
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

    expect(pushMock).toHaveBeenCalledWith("/app/orders/quick-sale-order");
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
      screen.getByText(
        /You don’t have any products yet. Add products first or switch to quick sale./i,
      ),
    ).toBeInTheDocument();
  });
});
