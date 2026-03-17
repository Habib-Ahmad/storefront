"use client";

import Link from "next/link";
import { PlusIcon } from "@phosphor-icons/react";
import { Button } from "@/components/ui/button";
import { ReceiptSvg } from "@/components/illustrations";

export default function OrdersPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Orders</h1>
        <Link href="/app/orders/new">
          <Button size="sm" className="gap-1.5">
            <PlusIcon className="size-4" weight="bold" />
            New Order
          </Button>
        </Link>
      </div>
      <div className="card-3d flex flex-col items-center justify-center rounded-2xl p-8 text-center">
        <ReceiptSvg className="size-36" />
        <p className="mt-3 text-sm text-muted-foreground">Your orders will appear here</p>
        <Link href="/app/orders/new" className="mt-3">
          <Button variant="outline" size="sm">
            Create your first order
          </Button>
        </Link>
      </div>
    </div>
  );
}
