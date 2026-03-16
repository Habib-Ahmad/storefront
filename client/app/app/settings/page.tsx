"use client";

import { GearSixIcon } from "@phosphor-icons/react";

export default function SettingsPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Settings</h1>
      </div>
      <div className="card-3d rounded-2xl p-8 flex flex-col items-center justify-center text-center">
        <div className="size-20 rounded-full bg-primary/10 flex items-center justify-center mb-3">
          <GearSixIcon className="size-10 text-primary/50" weight="fill" />
        </div>
        <p className="text-sm text-muted-foreground">
          Store settings will be available here
        </p>
      </div>
    </div>
  );
}
