"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { PlusIcon, CaretLeftIcon, CaretRightIcon } from "@phosphor-icons/react";
import { Button } from "@/components/ui/button";
import { ReceiptSvg } from "@/components/illustrations";
import { useOrders } from "@/hooks/use-orders";
import type { OrderListView } from "@/lib/types/orders";
import { OrderCard } from "./order-card";
import { OrderSkeleton } from "./order-skeleton";
import { OrderViewSwitcher } from "./order-view-switcher";

const orderViewOptions = [
  {
    value: "actionable",
    label: "Action needed",
    description: "Paid orders still waiting on you to prepare, confirm, or dispatch.",
  },
  {
    value: "active",
    label: "Open orders",
    description: "Everything in flight, including unpaid and already moving orders.",
  },
  {
    value: "cancelled",
    label: "Cancelled",
    description: "Failed, refunded, or cancelled orders that no longer need work.",
  },
  {
    value: "all",
    label: "All orders",
    description: "The full ledger when you need a complete operational view.",
  },
] as const satisfies ReadonlyArray<{
  value: OrderListView;
  label: string;
  description: string;
}>;

export default function OrdersPage() {
  const [page, setPage] = useState(1);
  const [view, setView] = useState<OrderListView>("actionable");
  const perPage = 12;

  const { data, isLoading } = useOrders({ page, per_page: perPage, view });

  useEffect(() => {
    setPage(1);
  }, [view]);

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

      <OrderViewSwitcher value={view} onChange={setView} options={orderViewOptions} />

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
          <p className="mt-3 text-sm text-muted-foreground">
            {view === "actionable"
              ? "Your paid orders that still need action will appear here"
              : view === "active"
                ? "Your open orders will appear here"
                : view === "cancelled"
                  ? "No cancelled or failed orders yet"
                  : "Your orders will appear here"}
          </p>
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
