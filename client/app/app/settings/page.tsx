"use client";

import { GearSixIcon } from "@phosphor-icons/react";

export default function SettingsPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Settings</h1>
      </div>
      <div className="card-3d flex flex-col items-center justify-center rounded-2xl p-8 text-center">
        <div className="mb-3 flex size-20 items-center justify-center rounded-full bg-primary/10">
          <GearSixIcon className="size-10 text-primary/50" weight="fill" />
        </div>
        <p className="text-sm text-muted-foreground">Store settings will be available here</p>
      </div>
    </div>
  );
}
