"use client";

import Link from "next/link";
import { ArrowRightIcon, GearSixIcon } from "@phosphor-icons/react";
import { Button } from "@/components/ui/button";

export default function SettingsPage() {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-sm text-muted-foreground">
          Account and business controls live here. Storefront launch now has its own dedicated
          workspace.
        </p>
      </div>

      <div className="card-3d rounded-2xl p-8 text-center">
        <div className="mx-auto flex size-16 items-center justify-center rounded-full bg-primary/10 text-primary">
          <GearSixIcon className="size-7" weight="fill" />
        </div>
        <div className="mt-5 space-y-2">
          <h2 className="text-xl font-semibold">Storefront moved</h2>
          <p className="text-sm text-muted-foreground">
            Draft links, public slugs, and publishing now live under Storefront so this flow stays
            visible and easy to revisit.
          </p>
        </div>
        <Link href="/app/storefront" className="inline-flex">
          <Button className="mt-6 gap-2">
            Open Storefront
            <ArrowRightIcon className="size-4" />
          </Button>
        </Link>
      </div>
    </div>
  );
}
