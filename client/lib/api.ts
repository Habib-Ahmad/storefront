import type {
  AddImageRequest,
  CreateOrderRequest,
  CreateOrderResponse,
  DispatchShipmentOption,
  CreateProductRequest,
  CreateVariantRequest,
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
  ResumePaymentResponse,
  SetModulesRequest,
  Shipment,
  Tenant,
  Tier,
  TrackingResponse,
  Transaction,
  UpdateStorefrontRequest,
  UpdateProductRequest,
  UpdateTenantRequest,
  UpdateUserRequest,
  User,
  Wallet,
} from "./types";
import {
  MeResponseSchema,
  DispatchShipmentOptionSchema,
  OrderItemSchema,
  OrderSchema,
  PaginatedOrdersResponseSchema,
  PaginatedProductsResponseSchema,
  ProductDetailResponseSchema,
  ProductImageSchema,
  ProductSchema,
  AnalyticsSummarySchema,
  ProductVariantSchema,
  ResumePaymentResponseSchema,
  TenantSchema,
  TierSchema,
  TrackingResponseSchema,
  TransactionSchema,
  UpdateTenantRequestSchema,
  UpdateStorefrontRequestSchema,
  UpdateUserRequestSchema,
  UserSchema,
  WalletSchema,
  createPaginatedResponseSchema,
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
  const page = params.page ?? 1;
  const perPage = params.per_page ?? 20;
  s.set("limit", String(perPage));
  s.set("offset", String((page - 1) * perPage));
  return s.toString();
}

function withStorefrontPublishedDefault(value: unknown): unknown {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return value;
  }

  if ("storefront_published" in value) {
    return value;
  }

  return {
    ...value,
    storefront_published: false,
  };
}

function normalizeMeResponse(value: unknown): unknown {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return value;
  }

  const payload = value as Record<string, unknown>;
  if (payload.onboarded !== true) {
    return value;
  }

  return {
    ...payload,
    tenant: withStorefrontPublishedDefault(payload.tenant),
  };
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

  private async request<T>(method: string, path: string, body?: unknown): Promise<T> {
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
  getMe = async (): Promise<MeResponse> =>
    MeResponseSchema.parse(normalizeMeResponse(await this.request<unknown>("GET", "/auth/me")));

  // Tiers (public)
  getTiers = async (): Promise<Tier[]> =>
    TierSchema.array().parse(await this.request<unknown>("GET", "/tiers"));

  // Tracking (public)
  track = async (slug: string): Promise<TrackingResponse> =>
    TrackingResponseSchema.parse(
      await this.request<unknown>("GET", `/track/${encodeURIComponent(slug)}`),
    );

  confirmTrackedOrderPayment = async (
    slug: string,
    data: { reference?: string; trxref?: string },
  ): Promise<TrackingResponse> =>
    TrackingResponseSchema.parse(
      await this.request<unknown>(
        "POST",
        `/track/${encodeURIComponent(slug)}/confirm-payment`,
        data,
      ),
    );

  resumeTrackedOrderPayment = async (slug: string): Promise<ResumePaymentResponse> =>
    ResumePaymentResponseSchema.parse(
      await this.request<unknown>("POST", `/track/${encodeURIComponent(slug)}/resume-payment`),
    );

  // Tenants
  onboard = async (data: OnboardRequest): Promise<Tenant> =>
    TenantSchema.parse(
      withStorefrontPublishedDefault(await this.request<unknown>("POST", "/tenants/onboard", data)),
    );

  updateTenant = async (data: UpdateTenantRequest): Promise<void> => {
    UpdateTenantRequestSchema.parse(data);
    await this.request<void>("PUT", "/tenants/me", data);
  };

  updateStorefront = async (data: UpdateStorefrontRequest): Promise<void> => {
    UpdateStorefrontRequestSchema.parse(data);
    await this.request<void>("PUT", "/tenants/me/storefront", data);
  };

  setModules = (data: SetModulesRequest) => this.request<void>("PUT", "/tenants/me/modules", data);

  // Users
  getUser = async (): Promise<User> =>
    UserSchema.parse(await this.request<unknown>("GET", "/users/me"));

  updateUser = async (data: UpdateUserRequest): Promise<void> => {
    UpdateUserRequestSchema.parse(data);
    await this.request<void>("PUT", "/users/me", data);
  };

  // Products
  getProducts = async (params: PaginationParams): Promise<PaginatedResponse<Product>> =>
    PaginatedProductsResponseSchema.parse(
      await this.request<unknown>("GET", `/products?${qs(params)}`),
    );

  getProduct = async (id: string): Promise<ProductDetailResponse> =>
    ProductDetailResponseSchema.parse(await this.request<unknown>("GET", `/products/${id}`));

  createProduct = async (data: CreateProductRequest): Promise<Product> =>
    ProductSchema.parse(await this.request<unknown>("POST", "/products", data));

  updateProduct = (id: string, data: UpdateProductRequest) =>
    this.request<void>("PUT", `/products/${id}`, data);

  deleteProduct = (id: string) => this.request<void>("DELETE", `/products/${id}`);

  // Variants
  getVariants = async (productId: string): Promise<ProductVariant[]> =>
    ProductVariantSchema.array().parse(
      await this.request<unknown>("GET", `/products/${productId}/variants`),
    );

  createVariant = async (productId: string, data: CreateVariantRequest): Promise<ProductVariant> =>
    ProductVariantSchema.parse(
      await this.request<unknown>("POST", `/products/${productId}/variants`, data),
    );

  updateVariant = (productId: string, variantId: string, data: CreateVariantRequest) =>
    this.request<void>("PUT", `/products/${productId}/variants/${variantId}`, data);

  deleteVariant = (productId: string, variantId: string) =>
    this.request<void>("DELETE", `/products/${productId}/variants/${variantId}`);

  // Images
  getImages = async (productId: string): Promise<ProductImage[]> =>
    ProductImageSchema.array().parse(
      await this.request<unknown>("GET", `/products/${productId}/images`),
    );

  addImage = async (productId: string, data: AddImageRequest): Promise<ProductImage> =>
    ProductImageSchema.parse(
      await this.request<unknown>("POST", `/products/${productId}/images`, data),
    );

  updateImage = (productId: string, imageId: string, data: AddImageRequest) =>
    this.request<void>("PUT", `/products/${productId}/images/${imageId}`, data);

  deleteImage = (productId: string, imageId: string) =>
    this.request<void>("DELETE", `/products/${productId}/images/${imageId}`);

  // Orders
  getOrders = async (params: PaginationParams): Promise<PaginatedResponse<Order>> =>
    PaginatedOrdersResponseSchema.parse(
      await this.request<unknown>("GET", `/orders?${qs(params)}`),
    );

  getOrder = async (id: string): Promise<Order> =>
    OrderSchema.parse(await this.request<unknown>("GET", `/orders/${id}`));

  getOrderItems = async (orderId: string): Promise<OrderItem[]> =>
    OrderItemSchema.array().parse(await this.request<unknown>("GET", `/orders/${orderId}/items`));

  getOrderDispatchOptions = async (orderId: string): Promise<DispatchShipmentOption[]> =>
    DispatchShipmentOptionSchema.array().parse(
      await this.request<unknown>("GET", `/orders/${orderId}/dispatch-options`),
    );

  createOrder = (data: CreateOrderRequest) =>
    this.request<CreateOrderResponse>("POST", "/orders", data);

  cancelOrder = (id: string) => this.request<{ status: string }>("POST", `/orders/${id}/cancel`);

  resumeOrderPayment = async (id: string): Promise<ResumePaymentResponse> =>
    ResumePaymentResponseSchema.parse(
      await this.request<unknown>("POST", `/orders/${id}/resume-payment`),
    );

  dispatchOrder = (id: string, data: unknown) =>
    this.request<Shipment>("POST", `/orders/${id}/dispatch`, data);

  // Wallet
  getWallet = async (): Promise<Wallet> =>
    WalletSchema.parse(await this.request<unknown>("GET", "/wallet"));

  getTransactions = async (params: PaginationParams): Promise<PaginatedResponse<Transaction>> =>
    createPaginatedResponseSchema(TransactionSchema).parse(
      await this.request<unknown>("GET", `/wallet/transactions?${qs(params)}`),
    );

  // Analytics
  getAnalyticsSummary = async (from: string, to: string) =>
    AnalyticsSummarySchema.parse(
      await this.request<unknown>("GET", `/analytics/summary?from=${from}&to=${to}`),
    );

  // Media
  getUploadUrl = () =>
    this.request<{ id: string; upload_url: string }>("POST", "/media/upload-url");
}

export const api = new ApiClient();
