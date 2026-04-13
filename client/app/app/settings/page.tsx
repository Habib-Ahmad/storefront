"use client";

import Link from "next/link";
import { ArrowRightIcon, TruckIcon, UserCircleIcon } from "@phosphor-icons/react";
import { Button } from "@/components/ui/button";

export default function SettingsPage() {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-sm text-muted-foreground">
          Keep account and launch controls organized here. Delivery setup now lives in its own
          dedicated subpage.
        </p>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <Link href="/app/settings/logistics" className="block">
          <div className="card-3d h-full rounded-2xl p-6 transition-transform hover:-translate-y-0.5">
            <div className="flex items-start justify-between gap-4">
              <div>
                <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">
                  Delivery
                </p>
                <h2 className="mt-2 text-xl font-semibold">Logistics setup</h2>
                <p className="mt-2 text-sm text-muted-foreground">
                  Fill in pickup address details, contact phone, and logistics email so delivery
                  quotes can be activated.
                </p>
              </div>
              <div className="flex size-12 items-center justify-center rounded-full bg-primary/10 text-primary">
                <TruckIcon className="size-6" weight="fill" />
              </div>
            </div>
            <div className="mt-6 inline-flex items-center gap-2 text-sm font-medium text-primary">
              Open logistics setup
              <ArrowRightIcon className="size-4" />
            </div>
          </div>
        </Link>

        <Link href="/app/storefront" className="block">
          <div className="card-3d h-full rounded-2xl p-6 transition-transform hover:-translate-y-0.5">
            <div className="flex items-start justify-between gap-4">
              <div>
                <p className="text-xs tracking-[0.18em] text-muted-foreground uppercase">Launch</p>
                <h2 className="mt-2 text-xl font-semibold">Storefront controls</h2>
                <p className="mt-2 text-sm text-muted-foreground">
                  Manage your public slug, draft status, and storefront publishing from the
                  dedicated launch workspace.
                </p>
              </div>
              <div className="flex size-12 items-center justify-center rounded-full bg-primary/10 text-primary">
                <UserCircleIcon className="size-6" weight="fill" />
              </div>
            </div>
            <div className="mt-6 inline-flex items-center gap-2 text-sm font-medium text-primary">
              Open storefront controls
              <ArrowRightIcon className="size-4" />
            </div>
          </div>
        </Link>
      </div>
    </div>
  );
}
