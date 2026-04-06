import {
  CurrencyNgnIcon,
  EnvelopeSimpleIcon,
  MapPinIcon,
  NotePencilIcon,
  PhoneIcon,
  ReceiptIcon,
  XCircleIcon,
} from "@phosphor-icons/react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import type { FulfillmentStatus, Order, OrderItem, PaymentStatus } from "@/lib/types";
import {
  formatCurrency,
  formatDate,
  fulfillmentBadgeVariant,
  paymentBadgeVariant,
} from "./order-detail-formatters";

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

export function OrderSummaryCard({ order }: { order: Order }) {
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

export function DeliveryCard({ order }: { order: Order }) {
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

export function ItemsCard({ items, order }: { items: OrderItem[]; order: Order }) {
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

export function ActionCard({
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
      if (order.fulfillment_status === "completed") {
        return "This pickup order is complete.";
      }

      if (order.payment_status !== "paid") {
        return "This pickup order is waiting for payment. You can still cancel it before payment is completed.";
      }

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

export function OrderDetailSkeleton() {
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
