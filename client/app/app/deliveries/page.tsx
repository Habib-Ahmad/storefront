"use client";

import { DeliveryTruckSvg } from "@/components/illustrations";

export default function DeliveriesPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Deliveries</h1>
      </div>
      <div className="card-3d flex flex-col items-center justify-center rounded-2xl p-8 text-center">
        <DeliveryTruckSvg className="size-36" />
        <p className="mt-3 text-sm text-muted-foreground">
          Shipments and deliveries will be tracked here
        </p>
      </div>
    </div>
  );
}
