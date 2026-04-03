"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import {
  ArrowLeftIcon,
  XCircleIcon,
  SpinnerGapIcon,
  NotePencilIcon,
  ReceiptIcon,
  PhoneIcon,
  EnvelopeSimpleIcon,
  MapPinIcon,
  CurrencyNgnIcon,
} from "@phosphor-icons/react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { useCancelOrder, useOrder, useOrderItems } from "@/hooks/use-orders";
import { ApiError } from "@/lib/api";
import type { FulfillmentStatus, Order, OrderItem, PaymentStatus } from "@/lib/types";

function formatCurrency(amount: string) {
  return new Intl.NumberFormat("en-NG", {
    style: "currency",
    currency: "NGN",
    minimumFractionDigits: 0,
  }).format(parseFloat(amount));
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("en-NG", {
    day: "numeric",
    month: "short",
    year: "numeric",
  }).format(new Date(value));
}

function paymentBadgeVariant(status: PaymentStatus): "default" | "secondary" | "destructive" {
  switch (status) {
    case "paid":
      return "default";
    case "failed":
    case "refunded":
      return "destructive";
    case "pending":
    default:
      return "secondary";
  }
}

function fulfillmentBadgeVariant(
  status: FulfillmentStatus,
): "default" | "secondary" | "destructive" {
  switch (status) {
    case "delivered":
    case "shipped":
      return "default";
    case "cancelled":
      return "destructive";
    case "processing":
    default:
      return "secondary";
  }
}

function PaymentBadge({ status }: { status: PaymentStatus }) {
  return (
    <Badge variant={paymentBadgeVariant(status)} className="text-xs capitalize">
      {status}
    </Badge>
  );
}

function FulfillmentBadge({ status }: { status: FulfillmentStatus }) {
  return (
    <Badge variant={fulfillmentBadgeVariant(status)} className="text-xs capitalize">
      {status}
    </Badge>
  );
}

function DetailRow({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode;
  label: string;
  value: React.ReactNode;
}) {
  return (
    <div className="flex items-start gap-3">
      <div className="mt-0.5 text-muted-foreground">{icon}</div>
      <div className="min-w-0">
        <p className="text-xs text-muted-foreground">{label}</p>
        <div className="text-sm">{value}</div>
      </div>
    </div>
  );
}

function OrderSummaryCard({ order }: { order: Order }) {
  const customerName = order.customer_name?.trim() || "Walk-in customer";

  return (
    <div className="card-3d space-y-4 rounded-2xl p-5">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 space-y-1">
          <h2 className="truncate text-lg font-semibold">{customerName}</h2>
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <ReceiptIcon className="size-4" />
            <span>{order.tracking_slug}</span>
          </div>
        </div>

        <div className="flex flex-wrap items-center justify-end gap-2">
          <PaymentBadge status={order.payment_status} />
          <FulfillmentBadge status={order.fulfillment_status} />
          <Badge variant="secondary" className="text-xs capitalize">
            {order.payment_method}
          </Badge>
          <Badge variant="secondary" className="text-xs">
            {order.is_delivery ? "Delivery" : "No delivery"}
          </Badge>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <DetailRow
          icon={<CurrencyNgnIcon className="size-4" />}
          label="Total"
          value={
            <span className="font-medium text-primary">{formatCurrency(order.total_amount)}</span>
          }
        />
        <DetailRow
          label="Created"
          icon={<NotePencilIcon className="size-4" />}
          value={formatDate(order.created_at)}
        />
      </div>

      {((!order.is_delivery && order.customer_phone) || order.customer_email || order.note) && (
        <>
          <div className="border-t border-border/60 pt-4" />
          <div className="grid gap-4 md:grid-cols-2">
            {!order.is_delivery && order.customer_phone && (
              <DetailRow
                icon={<PhoneIcon className="size-4" />}
                label="Phone"
                value={order.customer_phone}
              />
            )}
            {order.customer_email && (
              <DetailRow
                icon={<EnvelopeSimpleIcon className="size-4" />}
                label="Email"
                value={order.customer_email}
              />
            )}
            {order.note && (
              <DetailRow
                icon={<NotePencilIcon className="size-4" />}
                label="Note"
                value={order.note}
              />
            )}
          </div>
        </>
      )}
    </div>
  );
}

function DeliveryCard({ order }: { order: Order }) {
  return (
    <div className="card-3d space-y-4 rounded-2xl p-5">
      <div>
        <h2 className="text-base font-semibold">Delivery</h2>
        <p className="text-sm text-muted-foreground">
          {order.is_delivery
            ? "Delivery details for this order."
            : "This order was saved without delivery."}
        </p>
      </div>

      {!order.is_delivery ? (
        <div className="rounded-xl border border-border/60 bg-background/50 p-4 text-sm text-muted-foreground">
          No delivery was added to this order.
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          <DetailRow
            icon={<PhoneIcon className="size-4" />}
            label="Phone"
            value={order.customer_phone || "No phone added"}
          />
          <DetailRow
            icon={<CurrencyNgnIcon className="size-4" />}
            label="Shipping fee"
            value={formatCurrency(order.shipping_fee)}
          />
          <div className="md:col-span-2">
            <DetailRow
              icon={<MapPinIcon className="size-4" />}
              label="Address"
              value={order.shipping_address || "No address added"}
            />
          </div>
        </div>
      )}
    </div>
  );
}

function ItemsCard({ items, order }: { items: OrderItem[]; order: Order }) {
  return (
    <div className="card-3d space-y-4 rounded-2xl p-5">
      <div className="flex items-center justify-between">
        <h2 className="text-base font-semibold">Items</h2>
        <Badge variant="secondary" className="text-xs">
          {items.length} {items.length === 1 ? "item" : "items"}
        </Badge>
      </div>

      {items.length === 0 ? (
        <div className="rounded-xl border border-border/60 bg-background/50 p-4 text-sm text-muted-foreground">
          This was saved as a quick order. No items were added.
        </div>
      ) : (
        <div className="space-y-3">
          {items.map((item) => (
            <div key={item.id} className="rounded-xl border border-border/60 bg-background/50 p-4">
              <div className="flex items-start justify-between gap-4">
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium">
                    {item.product_name || "Unnamed product"}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    {item.variant_label || "Default option"}
                  </p>
                  <p className="mt-1 text-xs text-muted-foreground">Qty: {item.quantity}</p>
                </div>
                <div className="text-right">
                  <p className="text-sm font-medium text-primary">
                    {formatCurrency(item.price_at_sale)}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    Total{" "}
                    {formatCurrency((parseFloat(item.price_at_sale) * item.quantity).toString())}
                  </p>
                </div>
              </div>
            </div>
          ))}

          {parseFloat(order.shipping_fee) > 0 && (
            <div className="flex items-center justify-between rounded-xl border border-border/60 bg-background/50 px-4 py-3 text-sm">
              <span className="text-muted-foreground">Shipping fee</span>
              <span className="font-medium">{formatCurrency(order.shipping_fee)}</span>
            </div>
          )}

          <div className="flex items-center justify-between rounded-xl border border-border/60 bg-background px-4 py-3">
            <span className="font-medium">Grand total</span>
            <span className="text-lg font-semibold text-primary">
              {formatCurrency(order.total_amount)}
            </span>
          </div>
        </div>
      )}
    </div>
  );
}

function ActionCard({
  order,
  onCancel,
  actionError,
}: {
  order: Order;
  onCancel: () => void;
  actionError: string | null;
}) {
  const canCancel = order.fulfillment_status === "processing";

  const helperText = (() => {
    if (!order.is_delivery) {
      return canCancel
        ? "No delivery was added to this order. You can still cancel it if it was created by mistake."
        : "No delivery was added to this order.";
    }

    if (order.fulfillment_status === "cancelled") {
      return "This delivery order has already been cancelled.";
    }

    if (order.fulfillment_status === "delivered") {
      return "This delivery order has already been delivered.";
    }

    if (order.fulfillment_status === "shipped") {
      return "This delivery order has already been dispatched.";
    }

    if (order.payment_status !== "paid") {
      return "Delivery was added to this order. Payment must be completed before dispatch.";
    }

    return "Delivery was added to this order. Dispatch setup is not ready in this screen yet.";
  })();

  return (
    <div className="card-3d space-y-4 rounded-2xl p-5">
      <div>
        <h2 className="text-base font-semibold">Order actions</h2>
        <p className="text-sm text-muted-foreground">
          Only actions that make sense for this order appear here.
        </p>
      </div>

      {actionError && (
        <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {actionError}
        </p>
      )}

      {canCancel ? (
        <div className="flex flex-wrap gap-3">
          <Button type="button" variant="destructive" onClick={onCancel} className="gap-2">
            <XCircleIcon className="size-4" />
            Cancel order
          </Button>
        </div>
      ) : null}

      <div className="rounded-xl border border-border/60 bg-background/50 p-4 text-sm text-muted-foreground">
        {helperText}
      </div>
    </div>
  );
}

function OrderDetailSkeleton() {
  return (
    <div className="mx-auto max-w-3xl space-y-4">
      <Skeleton className="h-8 w-28" />
      <div className="card-3d space-y-4 rounded-2xl p-5">
        <Skeleton className="h-5 w-40" />
        <Skeleton className="h-4 w-28" />
        <div className="grid gap-4 md:grid-cols-2">
          <Skeleton className="h-12 w-full" />
          <Skeleton className="h-12 w-full" />
        </div>
      </div>
      <div className="card-3d space-y-4 rounded-2xl p-5">
        <Skeleton className="h-5 w-20" />
        <Skeleton className="h-20 w-full" />
        <Skeleton className="h-20 w-full" />
      </div>
      <div className="card-3d space-y-4 rounded-2xl p-5">
        <Skeleton className="h-5 w-20" />
        <div className="flex gap-3">
          <Skeleton className="h-10 w-32" />
          <Skeleton className="h-10 w-32" />
        </div>
      </div>
    </div>
  );
}

export default function OrderDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();

  const { data: order, isLoading: orderLoading } = useOrder(id);
  const { data: items, isLoading: itemsLoading } = useOrderItems(id);
  const cancelOrder = useCancelOrder();

  const [cancelDialogOpen, setCancelDialogOpen] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  const loading = orderLoading || itemsLoading;
  const orderItems = useMemo(() => items ?? [], [items]);

  if (loading) {
    return <OrderDetailSkeleton />;
  }

  if (!order) {
    return (
      <div className="mx-auto max-w-3xl space-y-4">
        <Link href="/app/orders">
          <Button variant="ghost" size="sm" className="-ml-2 gap-1">
            <ArrowLeftIcon className="size-4" />
            Back
          </Button>
        </Link>
        <div className="card-3d rounded-2xl p-6">
          <p className="text-sm text-muted-foreground">Order not found.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl space-y-4">
      <div>
        <Link href="/app/orders">
          <Button variant="ghost" size="sm" className="-ml-2 gap-1">
            <ArrowLeftIcon className="size-4" />
            Back
          </Button>
        </Link>
        <h1 className="mt-1 text-2xl font-bold">Order details</h1>
        <p className="text-sm text-muted-foreground">Review this order and its current status.</p>
      </div>

      <OrderSummaryCard order={order} />
      <DeliveryCard order={order} />
      <ItemsCard items={orderItems} order={order} />
      <ActionCard
        order={order}
        actionError={actionError}
        onCancel={() => {
          setActionError(null);
          setCancelDialogOpen(true);
        }}
      />

      <Dialog open={cancelDialogOpen} onOpenChange={setCancelDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Cancel this order?</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            This will mark the order as cancelled and may reverse inventory or payment effects
            depending on backend rules.
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCancelDialogOpen(false)}>
              Go back
            </Button>
            <Button
              variant="destructive"
              disabled={cancelOrder.isPending}
              onClick={async () => {
                try {
                  await cancelOrder.mutateAsync(order.id);
                  setCancelDialogOpen(false);
                  router.refresh();
                } catch (err) {
                  if (err instanceof ApiError) {
                    setActionError(err.message);
                  } else {
                    setActionError("Unable to cancel order");
                  }
                  setCancelDialogOpen(false);
                }
              }}
            >
              {cancelOrder.isPending && <SpinnerGapIcon className="size-4 animate-spin" />}
              Cancel order
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
