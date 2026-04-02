"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Field, FieldArray, Form, Formik, type FormikErrors, type FormikTouched } from "formik";
import * as Yup from "yup";
import {
  ArrowLeftIcon,
  PlusIcon,
  SpinnerGapIcon,
  TrashIcon,
  CaretDownIcon,
  CaretUpIcon,
} from "@phosphor-icons/react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { useCreateOrder } from "@/hooks/use-orders";
import { useProducts, useVariants } from "@/hooks/use-products";
import { ApiError } from "@/lib/api";
import type { CreateOrderRequest, PaymentMethod, Product } from "@/lib/types";

type FulfillmentMode = "delivery" | "none";

type LineItemForm = {
  product_id: string;
  variant_id: string;
  quantity: string;
};

type FormValues = {
  quick_sale: boolean;
  fulfillment: FulfillmentMode;
  payment_method: PaymentMethod;
  customer_name: string;
  customer_phone: string;
  customer_email: string;
  shipping_address: string;
  note: string;
  shipping_fee: string;
  total_amount: string;
  items: LineItemForm[];
};

type ResolvedLineItem = LineItemForm & {
  resolvedVariants: ProductVariantLike[];
  resolvedLoading: boolean;
};

type ProductVariantLike = NonNullable<Product["variants"]>[number];

const emptyLineItem: LineItemForm = {
  product_id: "",
  variant_id: "",
  quantity: "1",
};

const schema = Yup.object({
  quick_sale: Yup.boolean().required(),
  fulfillment: Yup.mixed<FulfillmentMode>().oneOf(["none", "delivery"]).required(),
  payment_method: Yup.mixed<PaymentMethod>()
    .oneOf(["online", "cash", "transfer"])
    .required("Choose a payment method"),
  customer_name: Yup.string(),
  customer_phone: Yup.string().when("fulfillment", {
    is: "delivery",
    then: (s) => s.required("Phone is required for delivery"),
  }),
  customer_email: Yup.string().email("Enter a valid email"),
  shipping_address: Yup.string().when("fulfillment", {
    is: "delivery",
    then: (s) => s.required("Address is required for delivery"),
  }),
  note: Yup.string(),
  shipping_fee: Yup.string().test("shipping-fee", "Shipping fee must be 0 or more", (value) => {
    if (!value) return true;
    const n = Number(value);
    return Number.isFinite(n) && n >= 0;
  }),
  total_amount: Yup.string().when("quick_sale", {
    is: true,
    then: (s) =>
      s
        .required("Enter the order amount")
        .test("quick-sale-total", "Amount must be greater than 0", (value) => {
          const n = Number(value);
          return Number.isFinite(n) && n > 0;
        }),
  }),
  items: Yup.array().of(
    Yup.object({
      product_id: Yup.string().when("$quick_sale", {
        is: false,
        then: (s) => s.required("Choose a product"),
      }),
      variant_id: Yup.string().when("$quick_sale", {
        is: false,
        then: (s) => s.required("Choose an option"),
      }),
      quantity: Yup.string().when("$quick_sale", {
        is: false,
        then: (s) =>
          s
            .required("Enter quantity")
            .test("positive-qty", "Quantity must be at least 1", (value) => {
              const n = Number(value);
              return Number.isInteger(n) && n > 0;
            }),
      }),
    }),
  ),
}).test("items-or-quick-sale", "Add an item or switch to quick order", (values) => {
  if (!values) return false;
  return values.quick_sale || (values.items?.length ?? 0) > 0;
});

function formatCurrency(amount: number) {
  return new Intl.NumberFormat("en-NG", {
    style: "currency",
    currency: "NGN",
    minimumFractionDigits: 0,
  }).format(amount);
}

function parsePositiveNumber(value: string) {
  const n = Number(value);
  return Number.isFinite(n) && n > 0 ? n : 0;
}

function parseNonNegativeNumber(value: string) {
  const n = Number(value);
  return Number.isFinite(n) && n >= 0 ? n : 0;
}

function fieldError(err: unknown, touched: unknown, tried: boolean): string | undefined {
  if (typeof err !== "string") return undefined;
  if (tried) return err;
  return touched ? err : undefined;
}

function variantPrice(variants: ProductVariantLike[], variantId: string) {
  const variant = variants.find((v) => v.id === variantId);
  return variant ? Number(variant.price) : 0;
}

function lineSubtotal(variants: ProductVariantLike[], item: LineItemForm) {
  const qty = Number(item.quantity);
  if (!Number.isInteger(qty) || qty <= 0) return 0;
  return variantPrice(variants, item.variant_id) * qty;
}

function sortSelectableVariants(variants: Product["variants"] | undefined) {
  if (!variants?.length) return [];
  return [...variants].sort((a, b) => {
    if (a.is_default === b.is_default) return 0;
    return a.is_default ? -1 : 1;
  });
}

function SectionHeader({ title, description }: { title: string; description: string }) {
  return (
    <div className="space-y-1 border-b border-border/60 pb-4">
      <h2 className="text-xl font-semibold tracking-tight">{title}</h2>
      <p className="text-sm text-muted-foreground">{description}</p>
    </div>
  );
}

function ChoiceCard({
  active,
  title,
  description,
  onClick,
}: {
  active: boolean;
  title: string;
  description: string;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`rounded-2xl border p-4 text-left transition ${
        active
          ? "border-primary bg-primary/6 shadow-sm ring-2 ring-primary/20"
          : "border-border/60 bg-background/50 hover:border-primary/40"
      }`}
    >
      <div className="space-y-1">
        <p className="font-medium">{title}</p>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
    </button>
  );
}

function PaymentPill({
  active,
  label,
  onClick,
}: {
  active: boolean;
  label: PaymentMethod;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`rounded-full border px-4 py-2 text-sm font-medium capitalize transition ${
        active
          ? "border-primary bg-primary text-primary-foreground"
          : "border-border/60 bg-background/50 hover:border-primary/40"
      }`}
    >
      {label}
    </button>
  );
}

function VariantLoader({
  productId,
  cachedVariants,
  onResolved,
}: {
  productId: string;
  cachedVariants: ProductVariantLike[] | undefined;
  onResolved: (variants: ProductVariantLike[]) => void;
}) {
  const { data = [], isLoading } = useVariants(productId, {
    enabled: !!productId && !cachedVariants,
  });

  useEffect(() => {
    if (!productId || isLoading) return;
    onResolved(sortSelectableVariants(data));
  }, [data, isLoading, onResolved, productId]);

  return null;
}

export default function NewOrderPage() {
  const router = useRouter();
  const createOrder = useCreateOrder();
  const { data: productsRes, isLoading: productsLoading } = useProducts({ page: 1, per_page: 100 });

  const products = useMemo(() => productsRes?.data ?? [], [productsRes]);
  const productsById = useMemo(
    () => Object.fromEntries(products.map((p) => [p.id, p])),
    [products],
  );

  const [variantCache, setVariantCache] = useState<Record<string, ProductVariantLike[]>>({});
  const [loadingProductIds, setLoadingProductIds] = useState<Record<string, boolean>>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [showExtras, setShowExtras] = useState(false);

  const initialValues: FormValues = {
    quick_sale: true,
    fulfillment: "none",
    payment_method: "cash",
    customer_name: "",
    customer_phone: "",
    customer_email: "",
    shipping_address: "",
    note: "",
    shipping_fee: "",
    total_amount: "",
    items: [{ ...emptyLineItem }],
  };

  const fallbackVariantsByProductId = useMemo(
    () =>
      Object.fromEntries(
        products.map((product) => [product.id, sortSelectableVariants(product.variants)]),
      ) as Record<string, ProductVariantLike[]>,
    [products],
  );

  return (
    <div className="mx-auto max-w-3xl space-y-5">
      <div>
        <Link href="/app/orders">
          <Button variant="ghost" size="sm" className="-ml-2 gap-1">
            <ArrowLeftIcon className="size-4" />
            Back
          </Button>
        </Link>
        <h1 className="mt-1 text-2xl font-bold">New Order</h1>
        <p className="text-sm text-muted-foreground">
          Record a new order in the quickest way that fits this customer.
        </p>
      </div>

      <Formik
        initialValues={initialValues}
        validationSchema={schema}
        validateOnBlur
        validateOnChange={false}
        validateOnMount={false}
        context={{ quick_sale: false }}
        onSubmit={async (values) => {
          setFormError(null);

          const isDelivery = values.fulfillment === "delivery";
          const payload: CreateOrderRequest = {
            is_delivery: isDelivery,
            payment_method: values.payment_method,
            customer_name: values.customer_name.trim() || null,
            customer_phone: values.customer_phone.trim() || null,
            customer_email: values.customer_email.trim() || null,
            shipping_address: isDelivery ? values.shipping_address.trim() || null : null,
            note: values.note.trim() || null,
            shipping_fee: isDelivery ? parseNonNegativeNumber(values.shipping_fee) : 0,
          };

          if (values.quick_sale) {
            payload.total_amount = parsePositiveNumber(values.total_amount);
            payload.items = [];
          } else {
            payload.items = values.items
              .filter((item) => item.variant_id)
              .map((item) => ({
                variant_id: item.variant_id,
                quantity: Number(item.quantity),
              }));
          }

          try {
            const created = await createOrder.mutateAsync(payload);

            if (created.authorization_url) {
              window.location.href = created.authorization_url;
              return;
            }

            router.push(`/app/orders/${created.id}`);
          } catch (err) {
            if (err instanceof ApiError) {
              setFormError(err.message);
              return;
            }
            setFormError("Unable to create order");
          }
        }}
      >
        {({
          values,
          errors,
          touched,
          isSubmitting,
          submitCount,
          setFieldValue,
          setFieldTouched,
        }) => {
          const tried = submitCount > 0;
          const isDelivery = values.fulfillment === "delivery";

          const itemsWithResolvedVariants: ResolvedLineItem[] = values.items.map((item) => {
            const variants =
              variantCache[item.product_id] ?? fallbackVariantsByProductId[item.product_id] ?? [];

            return {
              ...item,
              resolvedVariants: variants,
              resolvedLoading: !!loadingProductIds[item.product_id] && variants.length === 0,
            };
          });

          const itemsTotal = itemsWithResolvedVariants.reduce((sum, item) => {
            return sum + lineSubtotal(item.resolvedVariants, item);
          }, 0);

          const shippingFee = isDelivery ? parseNonNegativeNumber(values.shipping_fee) : 0;
          const quickSaleAmount = parsePositiveNumber(values.total_amount);
          const baseTotal = values.quick_sale ? quickSaleAmount : itemsTotal;
          const grandTotal = baseTotal + shippingFee;

          const productIdsNeedingVariants = Array.from(
            new Set(
              values.items
                .map((item) => item.product_id)
                .filter((productId) => {
                  if (!productId) return false;
                  const cached = variantCache[productId];
                  const fallback = fallbackVariantsByProductId[productId];
                  return !cached && (!fallback || fallback.length === 0);
                }),
            ),
          );

          return (
            <Form className="space-y-5">
              {productIdsNeedingVariants.map((productId) => (
                <VariantLoader
                  key={productId}
                  productId={productId}
                  cachedVariants={variantCache[productId]}
                  onResolved={(variants) => {
                    setVariantCache((prev) => {
                      if (prev[productId]) return prev;
                      return { ...prev, [productId]: variants };
                    });

                    setLoadingProductIds((prev) => {
                      if (!prev[productId]) return prev;
                      return { ...prev, [productId]: false };
                    });

                    values.items.forEach((item, index) => {
                      if (item.product_id !== productId) return;

                      const currentVariantId = values.items[index].variant_id;
                      const currentStillExists = variants.some(
                        (variant) => variant.id === currentVariantId,
                      );
                      const defaultVariant =
                        variants.find((variant) => variant.is_default) ?? variants[0];

                      if (!currentStillExists) {
                        setFieldValue(`items.${index}.variant_id`, defaultVariant?.id ?? "", false);
                      }
                    });
                  }}
                />
              ))}

              {formError && (
                <p className="rounded-lg bg-destructive/10 px-3 py-2 text-center text-sm text-destructive">
                  {formError}
                </p>
              )}

              <div className="card-3d space-y-5 rounded-3xl border p-6">
                <SectionHeader
                  title="What are you selling?"
                  description="Choose the quickest way to record this order."
                />

                <div className="grid gap-3 md:grid-cols-2">
                  <ChoiceCard
                    active={values.quick_sale}
                    title="Quick order"
                    description="Enter the amount for this order without adding products."
                    onClick={() => setFieldValue("quick_sale", true)}
                  />

                  <ChoiceCard
                    active={!values.quick_sale}
                    title="Choose products"
                    description="Pick products and quantities from your products."
                    onClick={() => setFieldValue("quick_sale", false)}
                  />
                </div>
              </div>

              {!values.quick_sale ? (
                <div className="card-3d space-y-5 rounded-3xl border p-6">
                  <SectionHeader title="Items" description="Choose what the customer is buying." />

                  <div className="flex items-center justify-between gap-3">
                    <p className="text-sm text-muted-foreground">Add one or more products.</p>
                    <Button
                      type="button"
                      size="default"
                      className="gap-2"
                      onClick={() =>
                        setFieldValue("items", [...values.items, { ...emptyLineItem }])
                      }
                    >
                      <PlusIcon className="size-4" />
                      Add item
                    </Button>
                  </div>

                  {productsLoading ? (
                    <p className="text-sm text-muted-foreground">Loading products…</p>
                  ) : products.length === 0 ? (
                    <div className="rounded-xl border border-border/60 bg-background/50 p-4 text-sm text-muted-foreground">
                      You don’t have any products yet. Add products first or switch to quick order.
                    </div>
                  ) : (
                    <FieldArray name="items">
                      {() => (
                        <div className="space-y-4">
                          {itemsWithResolvedVariants.map((item, index) => {
                            const itemErrors =
                              (errors.items?.[index] as FormikErrors<LineItemForm> | undefined) ??
                              undefined;
                            const itemTouched =
                              (touched.items?.[index] as FormikTouched<LineItemForm> | undefined) ??
                              undefined;

                            const product = productsById[item.product_id];
                            const variants = item.resolvedVariants;
                            const subtotal = lineSubtotal(variants, item);

                            return (
                              <div
                                key={index}
                                className="rounded-2xl border border-border/60 bg-background/50 p-4"
                              >
                                <div className="mb-4 flex items-center justify-between">
                                  <h3 className="text-base font-semibold">Item {index + 1}</h3>
                                  {values.items.length > 1 && (
                                    <Button
                                      type="button"
                                      variant="ghost"
                                      size="sm"
                                      className="h-8 px-2 text-destructive hover:text-destructive"
                                      onClick={() => {
                                        const next = values.items.filter((_, i) => i !== index);
                                        setFieldValue(
                                          "items",
                                          next.length ? next : [{ ...emptyLineItem }],
                                        );
                                      }}
                                    >
                                      <TrashIcon className="size-4" />
                                    </Button>
                                  )}
                                </div>

                                <div className="grid gap-4 md:grid-cols-3">
                                  <div className="space-y-1.5">
                                    <Label htmlFor={`items.${index}.product_id`}>Product</Label>
                                    <Field
                                      as="select"
                                      id={`items.${index}.product_id`}
                                      name={`items.${index}.product_id`}
                                      className="flex h-10 w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm outline-none"
                                      onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {
                                        const nextProductId = e.target.value;
                                        const cachedVariants =
                                          variantCache[nextProductId] ??
                                          fallbackVariantsByProductId[nextProductId] ??
                                          [];
                                        const defaultVariant =
                                          cachedVariants.find((variant) => variant.is_default) ??
                                          cachedVariants[0];

                                        setFieldValue(`items.${index}.product_id`, nextProductId);
                                        setFieldTouched(`items.${index}.product_id`, true, false);
                                        setFieldTouched(`items.${index}.variant_id`, true, false);

                                        if (!nextProductId) {
                                          setFieldValue(`items.${index}.variant_id`, "", false);
                                          return;
                                        }

                                        if (cachedVariants.length > 0) {
                                          setFieldValue(
                                            `items.${index}.variant_id`,
                                            defaultVariant?.id ?? "",
                                            false,
                                          );
                                          return;
                                        }

                                        setLoadingProductIds((prev) => ({
                                          ...prev,
                                          [nextProductId]: true,
                                        }));
                                        setFieldValue(`items.${index}.variant_id`, "", false);
                                      }}
                                    >
                                      <option value="">Choose a product</option>
                                      {products.map((productOption) => (
                                        <option key={productOption.id} value={productOption.id}>
                                          {productOption.name}
                                        </option>
                                      ))}
                                    </Field>
                                    {fieldError(
                                      itemErrors?.product_id,
                                      itemTouched?.product_id,
                                      tried,
                                    ) && (
                                      <p className="text-xs text-destructive">
                                        {itemErrors?.product_id}
                                      </p>
                                    )}
                                  </div>

                                  <div className="space-y-1.5">
                                    <Label htmlFor={`items.${index}.variant_id`}>Option</Label>
                                    <Field
                                      as="select"
                                      id={`items.${index}.variant_id`}
                                      name={`items.${index}.variant_id`}
                                      className="flex h-10 w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm outline-none"
                                      disabled={
                                        !product || item.resolvedLoading || variants.length === 0
                                      }
                                      onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {
                                        setFieldValue(`items.${index}.variant_id`, e.target.value);
                                        setFieldTouched(`items.${index}.variant_id`, true, false);
                                      }}
                                    >
                                      <option value="">
                                        {!product
                                          ? "Choose product first"
                                          : item.resolvedLoading
                                            ? "Loading options..."
                                            : variants.length === 0
                                              ? "No options available"
                                              : "Choose an option"}
                                      </option>
                                      {variants.map((variant) => (
                                        <option key={variant.id} value={variant.id}>
                                          {variant.sku} — {formatCurrency(Number(variant.price))}
                                        </option>
                                      ))}
                                    </Field>
                                    {fieldError(
                                      itemErrors?.variant_id,
                                      itemTouched?.variant_id,
                                      tried,
                                    ) && (
                                      <p className="text-xs text-destructive">
                                        {itemErrors?.variant_id}
                                      </p>
                                    )}
                                  </div>

                                  <div className="space-y-1.5">
                                    <Label htmlFor={`items.${index}.quantity`}>Quantity</Label>
                                    <Field
                                      as={Input}
                                      id={`items.${index}.quantity`}
                                      name={`items.${index}.quantity`}
                                      type="number"
                                      min="1"
                                      className="h-10"
                                    />
                                    {fieldError(
                                      itemErrors?.quantity,
                                      itemTouched?.quantity,
                                      tried,
                                    ) && (
                                      <p className="text-xs text-destructive">
                                        {itemErrors?.quantity}
                                      </p>
                                    )}
                                  </div>
                                </div>

                                {item.product_id && item.variant_id && (
                                  <div className="mt-4 rounded-xl border border-border/60 bg-background px-3 py-3 text-sm">
                                    <div className="flex items-center justify-between gap-4">
                                      <span className="text-muted-foreground">
                                        {variants.find((v) => v.id === item.variant_id)?.sku ??
                                          "Selected option"}
                                      </span>
                                      <span className="font-medium text-foreground">
                                        {formatCurrency(subtotal)}
                                      </span>
                                    </div>
                                  </div>
                                )}
                              </div>
                            );
                          })}

                          {typeof errors.items === "string" && (
                            <p className="text-xs text-destructive">{errors.items}</p>
                          )}
                        </div>
                      )}
                    </FieldArray>
                  )}
                </div>
              ) : null}

              <div className="card-3d space-y-5 rounded-3xl border p-6">
                <SectionHeader
                  title="Payment"
                  description={
                    values.quick_sale
                      ? "Enter the amount and choose how the customer will pay."
                      : "Choose how the customer will pay for this order."
                  }
                />

                {values.quick_sale && (
                  <div className="space-y-1.5">
                    <Label htmlFor="total_amount">Amount (₦)</Label>
                    <Field
                      as={Input}
                      id="total_amount"
                      name="total_amount"
                      type="number"
                      min="0"
                      placeholder="0.00"
                      className="h-11"
                    />
                    {fieldError(errors.total_amount, touched.total_amount, tried) && (
                      <p className="text-xs text-destructive">{errors.total_amount}</p>
                    )}
                  </div>
                )}

                <div className="space-y-1.5">
                  <Label>Payment method</Label>
                  <div className="flex flex-wrap gap-2">
                    {(["cash", "transfer", "online"] as PaymentMethod[]).map((method) => (
                      <PaymentPill
                        key={method}
                        label={method}
                        active={values.payment_method === method}
                        onClick={() => setFieldValue("payment_method", method)}
                      />
                    ))}
                  </div>
                </div>
              </div>

              <div className="card-3d space-y-5 rounded-3xl border p-6">
                <SectionHeader
                  title="Fulfillment"
                  description={
                    isDelivery
                      ? "Add the delivery details for this order."
                      : "Add delivery details only if this order needs them."
                  }
                />

                {!isDelivery ? (
                  <div className="rounded-2xl border border-border/60 bg-background/50 p-4">
                    <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                      <div className="space-y-1">
                        <p className="font-medium">No delivery</p>
                        <p className="text-sm text-muted-foreground">
                          Only add delivery if this order needs an address and fee.
                        </p>
                      </div>
                      <Button
                        type="button"
                        variant="outline"
                        onClick={() => setFieldValue("fulfillment", "delivery")}
                      >
                        Add delivery
                      </Button>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-4">
                    <div className="flex items-center justify-between gap-3 rounded-2xl border border-primary/20 bg-primary/5 p-4">
                      <div className="space-y-1">
                        <p className="font-medium">Delivery</p>
                        <p className="text-sm text-muted-foreground">
                          Add the address and delivery fee for this order.
                        </p>
                      </div>
                      <Button
                        type="button"
                        variant="ghost"
                        onClick={() => {
                          setFieldValue("fulfillment", "none");
                          setFieldValue("shipping_address", "");
                          setFieldValue("shipping_fee", "");
                        }}
                      >
                        Remove delivery
                      </Button>
                    </div>

                    <div className="rounded-2xl border border-primary/20 bg-primary/5 p-4">
                      <div className="mb-4">
                        <h3 className="text-base font-semibold">Delivery details</h3>
                        <p className="text-sm text-muted-foreground">
                          Add only the basics for now.
                        </p>
                      </div>

                      <div className="grid gap-4 md:grid-cols-2">
                        <div className="space-y-1.5">
                          <Label htmlFor="customer_phone">Phone</Label>
                          <Field
                            as={Input}
                            id="customer_phone"
                            name="customer_phone"
                            placeholder="Customer phone number"
                            className="h-10"
                          />
                          {fieldError(errors.customer_phone, touched.customer_phone, tried) && (
                            <p className="text-xs text-destructive">{errors.customer_phone}</p>
                          )}
                        </div>

                        <div className="space-y-1.5">
                          <Label htmlFor="shipping_fee">Shipping fee (₦)</Label>
                          <Field
                            as={Input}
                            id="shipping_fee"
                            name="shipping_fee"
                            type="number"
                            min="0"
                            placeholder="0.00"
                            className="h-10"
                          />
                          {fieldError(errors.shipping_fee, touched.shipping_fee, tried) && (
                            <p className="text-xs text-destructive">{errors.shipping_fee}</p>
                          )}
                        </div>

                        <div className="space-y-1.5 md:col-span-2">
                          <Label htmlFor="shipping_address">Address</Label>
                          <Field
                            as="textarea"
                            id="shipping_address"
                            name="shipping_address"
                            className="flex min-h-24 w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm outline-none"
                            placeholder="Enter the delivery address"
                          />
                          {fieldError(errors.shipping_address, touched.shipping_address, tried) && (
                            <p className="text-xs text-destructive">{errors.shipping_address}</p>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </div>

              <div className="card-3d space-y-5 rounded-3xl border p-6">
                <div className="flex items-center justify-between gap-3">
                  <SectionHeader
                    title="More details"
                    description="Add customer details or a note only if you need them."
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="gap-1"
                    onClick={() => setShowExtras((prev) => !prev)}
                  >
                    {showExtras ? (
                      <>
                        Hide
                        <CaretUpIcon className="size-4" />
                      </>
                    ) : (
                      <>
                        Add details
                        <CaretDownIcon className="size-4" />
                      </>
                    )}
                  </Button>
                </div>

                {showExtras && (
                  <div className="space-y-4 rounded-2xl border border-border/60 bg-background/50 p-4">
                    <div className="grid gap-4 md:grid-cols-2">
                      <div className="space-y-1.5">
                        <Label htmlFor="customer_name">Customer name</Label>
                        <Field
                          as={Input}
                          id="customer_name"
                          name="customer_name"
                          placeholder="Optional"
                          className="h-10"
                        />
                      </div>

                      <div className="space-y-1.5">
                        <Label htmlFor="customer_phone">Phone</Label>
                        <Field
                          as={Input}
                          id="customer_phone"
                          name="customer_phone"
                          placeholder="Optional"
                          className="h-10"
                        />
                      </div>

                      <div className="space-y-1.5 md:col-span-2">
                        <Label htmlFor="customer_email">Email</Label>
                        <Field
                          as={Input}
                          id="customer_email"
                          name="customer_email"
                          type="email"
                          placeholder="Optional"
                          className="h-10"
                        />
                        {fieldError(errors.customer_email, touched.customer_email, tried) && (
                          <p className="text-xs text-destructive">{errors.customer_email}</p>
                        )}
                      </div>

                      <div className="space-y-1.5 md:col-span-2">
                        <Label htmlFor="note">Note</Label>
                        <Field
                          as="textarea"
                          id="note"
                          name="note"
                          className="flex min-h-20 w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm outline-none"
                          placeholder="Optional note"
                        />
                      </div>
                    </div>
                  </div>
                )}
              </div>

              <div className="card-3d space-y-5 rounded-3xl border p-6">
                <SectionHeader
                  title="Total"
                  description="Check the total before you save the order."
                />

                <div className="rounded-2xl border border-border/60 bg-background/50 p-4">
                  <div className="flex items-center justify-between text-sm text-muted-foreground">
                    <span>{values.quick_sale ? "Order amount" : "Items total"}</span>
                    <span>{formatCurrency(baseTotal)}</span>
                  </div>

                  {isDelivery && (
                    <div className="mt-2 flex items-center justify-between text-sm text-muted-foreground">
                      <span>Shipping fee</span>
                      <span>{formatCurrency(shippingFee)}</span>
                    </div>
                  )}

                  <Separator className="my-3" />

                  <div className="flex items-center justify-between">
                    <span className="font-medium">Grand total</span>
                    <span className="text-lg font-semibold text-primary">
                      {formatCurrency(grandTotal)}
                    </span>
                  </div>
                </div>
              </div>

              <div className="flex justify-end gap-3">
                <Link href="/app/orders">
                  <Button type="button" variant="outline">
                    Cancel
                  </Button>
                </Link>
                <Button type="submit" size="lg" disabled={isSubmitting || createOrder.isPending}>
                  {(isSubmitting || createOrder.isPending) && (
                    <SpinnerGapIcon className="size-4 animate-spin" />
                  )}
                  {values.payment_method === "online" ? "Continue to payment" : "Save order"}
                </Button>
              </div>
            </Form>
          );
        }}
      </Formik>
    </div>
  );
}
