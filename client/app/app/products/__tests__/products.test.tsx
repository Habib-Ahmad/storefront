import { describe, it, expect, vi, beforeEach } from "vitest";
import { fireEvent, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { PRODUCTS_KNOWN_STORAGE_KEY } from "@/lib/storage";

const pushMock = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: pushMock, back: vi.fn() }),
  usePathname: () => "/app/products",
  useParams: () => ({}),
}));

const mockUseProducts = vi.fn();
const mockCreateProduct = vi.fn();
const mockAddImage = vi.fn();
const mockUploadImageFile = vi.fn();
const mockUseSession = vi.fn();
vi.mock("@/hooks/use-products", () => ({
  useProducts: (...args: unknown[]) => mockUseProducts(...args),
  useCreateProduct: () => ({
    mutateAsync: mockCreateProduct,
    isPending: false,
  }),
  useAddImage: () => ({
    mutateAsync: mockAddImage,
    isPending: false,
  }),
}));

vi.mock("@/components/auth-provider", () => ({
  useSession: () => mockUseSession(),
}));

vi.mock("@/lib/media-upload", () => ({
  uploadImageFile: (...args: unknown[]) => mockUploadImageFile(...args),
}));

import ProductsPage from "@/app/app/products/page";
import NewProductPage from "@/app/app/products/new/page";

beforeEach(() => {
  vi.clearAllMocks();
  window.localStorage.clear();
  mockUseSession.mockReturnValue({
    session: { user: { id: "user-1" } },
    loading: false,
  });
  mockAddImage.mockResolvedValue(undefined);
  mockUploadImageFile.mockResolvedValue("https://cdn.example.com/image.jpg");
});

describe("ProductsPage", () => {
  it("shows empty state instead of product skeletons for first-time users while loading", () => {
    mockUseProducts.mockReturnValue({ data: undefined, isLoading: true });

    render(<ProductsPage />);

    expect(screen.getByText("Add your first product to get started")).toBeInTheDocument();
  });

  it("shows product skeletons while loading once products have existed before", () => {
    window.localStorage.setItem(`${PRODUCTS_KNOWN_STORAGE_KEY}:user-1`, "1");
    mockUseProducts.mockReturnValue({ data: undefined, isLoading: true });

    const { container } = render(<ProductsPage />);

    expect(screen.queryByText("Add your first product to get started")).not.toBeInTheDocument();
    expect(container.querySelectorAll(".card-3d [data-slot='skeleton']")).toHaveLength(12);
  });

  it("ignores another user's stored product flag while loading", () => {
    window.localStorage.setItem(`${PRODUCTS_KNOWN_STORAGE_KEY}:user-2`, "1");
    mockUseProducts.mockReturnValue({ data: undefined, isLoading: true });

    render(<ProductsPage />);

    expect(screen.getByText("Add your first product to get started")).toBeInTheDocument();
  });

  it("shows empty state when there are no products", () => {
    mockUseProducts.mockReturnValue({ data: { data: [], total: 0 }, isLoading: false });
    render(<ProductsPage />);
    expect(screen.getByText("Add your first product to get started")).toBeInTheDocument();
  });

  it("renders product cards with name, price, and availability", () => {
    mockUseProducts.mockReturnValue({
      data: {
        data: [
          {
            id: "p1",
            name: "Ankara Shirt",
            is_available: true,
            variants: [{ id: "v1", price: "5000", stock_qty: 10 }],
            images: [],
          },
          {
            id: "p2",
            name: "Leather Bag",
            is_available: false,
            variants: [{ id: "v2", price: "12000", stock_qty: 0 }],
            images: [],
          },
        ],
        total: 2,
      },
      isLoading: false,
    });
    render(<ProductsPage />);

    expect(screen.getByText("Ankara Shirt")).toBeInTheDocument();
    expect(screen.getByText("Leather Bag")).toBeInTheDocument();
    expect(screen.getByText("Active")).toBeInTheDocument();
    expect(screen.getByText("Draft")).toBeInTheDocument();
    expect(screen.getByText("10 in stock")).toBeInTheDocument();
    expect(screen.getByText("Out of stock")).toBeInTheDocument();
  });

  it("filters products by search input", async () => {
    mockUseProducts.mockReturnValue({
      data: {
        data: [
          {
            id: "p1",
            name: "Ankara Shirt",
            is_available: true,
            variants: [{ price: "5000" }],
            images: [],
          },
          {
            id: "p2",
            name: "Leather Bag",
            is_available: true,
            variants: [{ price: "12000" }],
            images: [],
          },
        ],
        total: 2,
      },
      isLoading: false,
    });
    render(<ProductsPage />);

    const searchInput = screen.getByPlaceholderText("Search products…");
    await userEvent.type(searchInput, "ankara");

    expect(screen.getByText("Ankara Shirt")).toBeInTheDocument();
    expect(screen.queryByText("Leather Bag")).not.toBeInTheDocument();
  });

  it("shows 'no results' message when search matches nothing", async () => {
    mockUseProducts.mockReturnValue({
      data: {
        data: [
          {
            id: "p1",
            name: "Ankara Shirt",
            is_available: true,
            variants: [{ price: "5000" }],
            images: [],
          },
        ],
        total: 1,
      },
      isLoading: false,
    });
    render(<ProductsPage />);

    await userEvent.type(screen.getByPlaceholderText("Search products…"), "zzz");

    expect(screen.getByText(/No products matching/)).toBeInTheDocument();
  });

  it("links each product card to its detail page", () => {
    mockUseProducts.mockReturnValue({
      data: {
        data: [
          {
            id: "p1",
            name: "Ankara Shirt",
            is_available: true,
            variants: [{ price: "5000" }],
            images: [],
          },
        ],
        total: 1,
      },
      isLoading: false,
    });
    render(<ProductsPage />);

    const link = screen.getByText("Ankara Shirt").closest("a");
    expect(link).toHaveAttribute("href", "/app/products/p1");
  });
});

describe("NewProductPage", () => {
  it("shows validation errors when submitting an empty form", async () => {
    render(<NewProductPage />);

    await userEvent.click(screen.getByText("Create product"));

    expect(await screen.findByText("Name is required")).toBeInTheDocument();
    expect(await screen.findByText("Description is required")).toBeInTheDocument();
    expect(await screen.findByText("Price is required")).toBeInTheDocument();
  });

  it("submits a valid product and lets the user continue from the success modal", async () => {
    mockCreateProduct.mockResolvedValue({ id: "new-id" });
    render(<NewProductPage />);

    await userEvent.type(screen.getByLabelText("Name"), "Ankara Shirt");
    await userEvent.type(
      screen.getByLabelText("Description"),
      "A soft Ankara shirt for everyday wear",
    );
    await userEvent.type(screen.getByLabelText("Price (₦)"), "5000");

    await userEvent.click(screen.getByText("Create product"));

    await vi.waitFor(() => {
      expect(mockCreateProduct).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "Ankara Shirt",
          is_available: true,
          variants: [expect.objectContaining({ sku: "Default" })],
        }),
      );
    });

    expect(await screen.findByText("Product created")).toBeInTheDocument();

    await userEvent.click(screen.getByText("View product"));
    expect(pushMock).toHaveBeenCalledWith("/app/products/new-id");
  });

  it("adds and removes variant options", async () => {
    render(<NewProductPage />);

    expect(
      screen.getByText("No option name needed unless you add more choices."),
    ).toBeInTheDocument();
    expect(screen.queryByLabelText(/Option name/)).not.toBeInTheDocument();

    await userEvent.click(screen.getByText("Add size, color, or other option"));
    expect(screen.getByText("Option 2")).toBeInTheDocument();

    // After adding, every option should ask for a customer-facing name.
    const optionNameLabels = screen.getAllByLabelText(/Option name/);
    expect(optionNameLabels).toHaveLength(2);
  });

  it("shows API error message when creation fails", async () => {
    const { ApiError } = await import("@/lib/api");
    mockCreateProduct.mockRejectedValue(new ApiError(422, "slug already taken"));
    render(<NewProductPage />);

    await userEvent.type(screen.getByLabelText("Name"), "Test");
    await userEvent.type(screen.getByLabelText("Description"), "Useful test product");
    await userEvent.type(screen.getByLabelText("Price (₦)"), "1000");

    await userEvent.click(screen.getByText("Create product"));

    expect(await screen.findByText("slug already taken")).toBeInTheDocument();
  });

  it("uploads selected images after product creation", async () => {
    mockCreateProduct.mockResolvedValue({ id: "new-id" });
    render(<NewProductPage />);

    await userEvent.type(screen.getByLabelText("Name"), "Ankara Shirt");
    await userEvent.type(
      screen.getByLabelText("Description"),
      "A soft Ankara shirt for everyday wear",
    );
    await userEvent.type(screen.getByLabelText("Price (₦)"), "5000");

    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement | null;
    expect(fileInput).not.toBeNull();

    const file = new File(["image"], "shirt.jpg", { type: "image/jpeg" });
    const createObjectURLSpy = vi.spyOn(URL, "createObjectURL").mockReturnValue("blob:shirt");

    fireEvent.change(fileInput!, { target: { files: [file] } });

    await userEvent.click(screen.getByText("Create product"));

    await vi.waitFor(() => {
      expect(mockUploadImageFile).toHaveBeenCalledWith(file, "new-id");
      expect(mockAddImage).toHaveBeenCalledWith({
        productId: "new-id",
        data: {
          url: "https://cdn.example.com/image.jpg",
          sort_order: 0,
          is_primary: true,
        },
      });
      expect(screen.getByText("Product created")).toBeInTheDocument();
    });

    await userEvent.click(screen.getByText("View product"));

    await vi.waitFor(() => {
      expect(pushMock).toHaveBeenCalledWith("/app/products/new-id");
    });

    createObjectURLSpy.mockRestore();
  });
});
