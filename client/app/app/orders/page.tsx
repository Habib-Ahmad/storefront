"use client";

import { useState } from "react";
import Link from "next/link";
import { PlusIcon, CaretLeftIcon, CaretRightIcon, ArrowSquareOutIcon } from "@phosphor-icons/react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { ReceiptSvg } from "@/components/illustrations";
import { useOrders } from "@/hooks/use-orders";
import type { Order } from "@/lib/types";

function formatCurrency(amount: string) {
  return new Intl.NumberFormat("en-NG", {
    style: "currency",
    currency: "NGN",
    minimumFractionDigits: 0,
  }).format(parseFloat(amount));
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("en-NG", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
    hour12: true,
    timeZone: "Africa/Lagos",
  }).format(new Date(value));
}

type CardBadge = {
  label: string;
  variant: "default" | "secondary" | "destructive";
};

function primaryStatusBadge(order: Order): CardBadge {
  if (order.fulfillment_status === "cancelled") {
    return { label: "Cancelled", variant: "destructive" };
  }

  if (order.payment_status === "failed") {
    return { label: "Payment failed", variant: "destructive" };
  }

  if (order.payment_status === "refunded") {
    return { label: "Refunded", variant: "destructive" };
  }

  if (order.fulfillment_status === "completed") {
    return { label: "Completed", variant: "default" };
  }

  if (order.fulfillment_status === "delivered") {
    return { label: "Delivered", variant: "default" };
  }

  if (order.fulfillment_status === "shipped") {
    return { label: "Shipped", variant: "default" };
  }

  if (order.payment_status === "pending") {
    return { label: "Awaiting payment", variant: "secondary" };
  }

  if (order.is_delivery && order.fulfillment_status === "processing") {
    return { label: "Ready for delivery", variant: "secondary" };
  }

  return { label: "Processing", variant: "secondary" };
}

function cardBadges(order: Order): CardBadge[] {
  const badges = [primaryStatusBadge(order)];

  if (order.is_delivery) {
    badges.push({ label: "Delivery", variant: "secondary" });
  }

  return badges;
}

function OrderCard({ order }: { order: Order }) {
  const customerName = order.customer_name?.trim() || "Walk-in customer";
  const badges = cardBadges(order);

  return (
    <Link href={`/app/orders/${order.id}`} className="block">
      <div className="card-3d space-y-3 rounded-2xl p-4 transition-all hover:ring-2 hover:ring-primary/20">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 space-y-1">
            <p className="truncate text-sm font-semibold">{customerName}</p>
            <p className="text-xs text-muted-foreground">{formatDateTime(order.created_at)}</p>
          </div>
          <div className="flex items-start gap-3">
            <div className="text-right">
              <p className="text-base font-semibold text-primary">
                {formatCurrency(order.total_amount)}
              </p>
            </div>
            <ArrowSquareOutIcon className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {badges.map((badge) => (
            <Badge key={badge.label} variant={badge.variant} className="text-xs">
              {badge.label}
            </Badge>
          ))}
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
