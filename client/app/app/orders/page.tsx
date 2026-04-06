"use client";

import { useState } from "react";
import Link from "next/link";
import { PlusIcon, CaretLeftIcon, CaretRightIcon } from "@phosphor-icons/react";
import { Button } from "@/components/ui/button";
import { ReceiptSvg } from "@/components/illustrations";
import { useOrders } from "@/hooks/use-orders";
import { OrderCard } from "./order-card";
import { OrderSkeleton } from "./order-skeleton";

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
