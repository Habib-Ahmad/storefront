"use client";

import { useState } from "react";
import Link from "next/link";
import {
  PlusIcon,
  CaretLeftIcon,
  CaretRightIcon,
  ReceiptIcon,
  ArrowSquareOutIcon,
} from "@phosphor-icons/react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { ReceiptSvg } from "@/components/illustrations";
import { useOrders } from "@/hooks/use-orders";
import type { Order, PaymentStatus, FulfillmentStatus } from "@/lib/types";

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

function OrderCard({ order }: { order: Order }) {
  const customerName = order.customer_name?.trim() || "Walk-in customer";

  return (
    <Link href={`/app/orders/${order.id}`} className="block">
      <div className="card-3d space-y-3 rounded-2xl p-4 transition-all hover:ring-2 hover:ring-primary/20">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 space-y-1">
            <p className="truncate text-sm font-semibold">{customerName}</p>
            <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
              <ReceiptIcon className="size-3.5" />
              <span className="truncate">{order.tracking_slug}</span>
            </div>
          </div>
          <ArrowSquareOutIcon className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <PaymentBadge status={order.payment_status} />
          <FulfillmentBadge status={order.fulfillment_status} />
          <Badge variant="secondary" className="text-xs capitalize">
            {order.payment_method}
          </Badge>
        </div>

        <div className="flex items-end justify-between gap-3">
          <div className="space-y-1">
            <p className="text-base font-semibold text-primary">
              {formatCurrency(order.total_amount)}
            </p>
            <p className="text-xs text-muted-foreground">{formatDate(order.created_at)}</p>
          </div>
          {order.is_delivery && (
            <Badge variant="secondary" className="text-xs">
              Delivery
            </Badge>
          )}
        </div>
      </div>
    </Link>
  );
}

function OrderSkeleton() {
  return (
    <div className="card-3d space-y-3 rounded-2xl p-4">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1 space-y-2">
          <Skeleton className="h-4 w-32" />
          <Skeleton className="h-3 w-24" />
        </div>
        <Skeleton className="h-4 w-4 rounded-full" />
      </div>
      <div className="flex gap-2">
        <Skeleton className="h-5 w-16 rounded-full" />
        <Skeleton className="h-5 w-20 rounded-full" />
        <Skeleton className="h-5 w-16 rounded-full" />
      </div>
      <div className="flex items-end justify-between gap-3">
        <div className="space-y-2">
          <Skeleton className="h-4 w-24" />
          <Skeleton className="h-3 w-20" />
        </div>
        <Skeleton className="h-5 w-14 rounded-full" />
      </div>
    </div>
  );
}

export default function OrdersPage() {
  const [page, setPage] = useState(1);
  const perPage = 12;

  const { data, isLoading } = useOrders({ page, per_page: perPage });

  const orders = data?.data ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / perPage));

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Orders</h1>
        <Link href="/app/orders/new">
          <Button className="gap-2 px-4">
            <PlusIcon className="size-4" weight="bold" />
            New Order
          </Button>
        </Link>
      </div>

      {isLoading && (
        <div className="space-y-3">
          {Array.from({ length: 4 }).map((_, i) => (
            <OrderSkeleton key={i} />
          ))}
        </div>
      )}

      {!isLoading && total === 0 && (
        <div className="card-3d flex flex-col items-center justify-center rounded-2xl p-8 text-center">
          <ReceiptSvg className="size-36" />
          <p className="mt-3 text-sm text-muted-foreground">Your orders will appear here</p>
          <Link href="/app/orders/new" className="mt-3">
            <Button variant="outline" className="px-4">
              Create your first order
            </Button>
          </Link>
        </div>
      )}

      {!isLoading && orders.length > 0 && (
        <div className="space-y-3">
          {orders.map((order) => (
            <OrderCard key={order.id} order={order} />
          ))}
        </div>
      )}

      {!isLoading && totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 pt-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1}
            onClick={() => setPage((p) => p - 1)}
          >
            <CaretLeftIcon className="size-4" />
          </Button>
          <span className="text-sm text-muted-foreground">
            {page} / {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages}
            onClick={() => setPage((p) => p + 1)}
          >
            <CaretRightIcon className="size-4" />
          </Button>
        </div>
      )}
    </div>
  );
}
