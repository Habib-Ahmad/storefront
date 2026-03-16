import type {
  AnalyticsSummary,
  CreateOrderRequest,
  CreateOrderResponse,
  CreateProductRequest,
  CreateVariantRequest,
  AddImageRequest,
  MeResponse,
  OnboardRequest,
  Order,
  OrderItem,
  PaginatedResponse,
  PaginationParams,
  Product,
  ProductDetailResponse,
  ProductImage,
  ProductVariant,
  SetModulesRequest,
  Shipment,
  Tenant,
  Tier,
  TrackingResponse,
  Transaction,
  UpdateProductRequest,
  UpdateTenantRequest,
  UpdateUserRequest,
  User,
  Wallet,
} from "./types";

// ── Error ──────────────────────────────────────────────

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
    public fields?: Record<string, string>,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

// ── Helpers ────────────────────────────────────────────

function qs(params: PaginationParams): string {
  const s = new URLSearchParams();
  if (params.page) s.set("page", String(params.page));
  if (params.per_page) s.set("per_page", String(params.per_page));
  return s.toString();
}

// ── Client ─────────────────────────────────────────────

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "";

class ApiClient {
  private token: string | null = null;
  private refreshHandler: (() => Promise<string | null>) | null = null;

  setToken(token: string | null) {
    this.token = token;
  }

  setRefreshHandler(handler: (() => Promise<string | null>) | null) {
    this.refreshHandler = handler;
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown,
  ): Promise<T> {
    const doFetch = () => {
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
      };
      if (this.token) headers["Authorization"] = `Bearer ${this.token}`;
      return fetch(`${API_BASE}${path}`, {
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined,
      });
    };

    let res = await doFetch();

    if (res.status === 401 && this.refreshHandler) {
      const newToken = await this.refreshHandler();
      if (newToken) {
        this.token = newToken;
        res = await doFetch();
      }
    }

    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new ApiError(res.status, err.error ?? "Unknown error", err.errors);
    }
    if (res.status === 204) return undefined as T;
    return res.json();
  }

  // Auth
  getMe = () => this.request<MeResponse>("GET", "/auth/me");

  // Tiers (public)
  getTiers = () => this.request<Tier[]>("GET", "/tiers");

  // Tracking (public)
  track = (slug: string) =>
    this.request<TrackingResponse>("GET", `/track/${encodeURIComponent(slug)}`);

  // Tenants
  onboard = (data: OnboardRequest) =>
    this.request<Tenant>("POST", "/tenants/onboard", data);

  getTenant = () => this.request<Tenant>("GET", "/tenants/me");

  updateTenant = (data: UpdateTenantRequest) =>
    this.request<void>("PUT", "/tenants/me", data);

  setModules = (data: SetModulesRequest) =>
    this.request<void>("PUT", "/tenants/me/modules", data);

  // Users
  getUser = () => this.request<User>("GET", "/users/me");

  updateUser = (data: UpdateUserRequest) =>
    this.request<void>("PUT", "/users/me", data);

  // Products
  getProducts = (params: PaginationParams) =>
    this.request<PaginatedResponse<Product>>("GET", `/products?${qs(params)}`);

  getProduct = (id: string) =>
    this.request<ProductDetailResponse>("GET", `/products/${id}`);

  createProduct = (data: CreateProductRequest) =>
    this.request<Product>("POST", "/products", data);

  updateProduct = (id: string, data: UpdateProductRequest) =>
    this.request<void>("PUT", `/products/${id}`, data);

  deleteProduct = (id: string) =>
    this.request<void>("DELETE", `/products/${id}`);

  // Variants
  getVariants = (productId: string) =>
    this.request<ProductVariant[]>("GET", `/products/${productId}/variants`);

  createVariant = (productId: string, data: CreateVariantRequest) =>
    this.request<ProductVariant>(
      "POST",
      `/products/${productId}/variants`,
      data,
    );

  updateVariant = (productId: string, variantId: string, data: CreateVariantRequest) =>
    this.request<void>(
      "PUT",
      `/products/${productId}/variants/${variantId}`,
      data,
    );

  deleteVariant = (productId: string, variantId: string) =>
    this.request<void>(
      "DELETE",
      `/products/${productId}/variants/${variantId}`,
    );

  // Images
  getImages = (productId: string) =>
    this.request<ProductImage[]>("GET", `/products/${productId}/images`);

  addImage = (productId: string, data: AddImageRequest) =>
    this.request<ProductImage>("POST", `/products/${productId}/images`, data);

  updateImage = (productId: string, imageId: string, data: AddImageRequest) =>
    this.request<void>(
      "PUT",
      `/products/${productId}/images/${imageId}`,
      data,
    );

  deleteImage = (productId: string, imageId: string) =>
    this.request<void>(
      "DELETE",
      `/products/${productId}/images/${imageId}`,
    );

  // Orders
  getOrders = (params: PaginationParams) =>
    this.request<PaginatedResponse<Order>>("GET", `/orders?${qs(params)}`);

  getOrder = (id: string) => this.request<Order>("GET", `/orders/${id}`);

  getOrderItems = (orderId: string) =>
    this.request<OrderItem[]>("GET", `/orders/${orderId}/items`);

  createOrder = (data: CreateOrderRequest) =>
    this.request<CreateOrderResponse>("POST", "/orders", data);

  cancelOrder = (id: string) =>
    this.request<{ status: string }>("POST", `/orders/${id}/cancel`);

  dispatchOrder = (id: string, data: unknown) =>
    this.request<Shipment>("POST", `/orders/${id}/dispatch`, data);

  // Wallet
  getWallet = () => this.request<Wallet>("GET", "/wallet");

  getTransactions = (params: PaginationParams) =>
    this.request<PaginatedResponse<Transaction>>(
      "GET",
      `/wallet/transactions?${qs(params)}`,
    );

  // Analytics
  getAnalyticsSummary = (from: string, to: string) =>
    this.request<AnalyticsSummary>(
      "GET",
      `/analytics/summary?from=${from}&to=${to}`,
    );
}

export const api = new ApiClient();
