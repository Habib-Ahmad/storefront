"use client";

import { DeliveryTruckSvg } from "@/components/illustrations";

export default function DeliveriesPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Deliveries</h1>
      </div>
      <div className="card-3d rounded-2xl p-8 flex flex-col items-center justify-center text-center">
        <DeliveryTruckSvg className="size-36" />
        <p className="text-sm text-muted-foreground mt-3">
          Shipments and deliveries will be tracked here
        </p>
      </div>
    </div>
  );
}
