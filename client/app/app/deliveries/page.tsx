"use client";

import { useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowRightIcon, SpinnerGapIcon, WarningCircleIcon } from "@phosphor-icons/react";
import { DeliveryTruckSvg } from "@/components/illustrations";
import { Button } from "@/components/ui/button";
import { useMe } from "@/hooks/use-auth";
import { missingLogisticsAddressFields, parseLogisticsAddress } from "@/lib/logistics-address";

export default function DeliveriesPage() {
  const router = useRouter();
  const { data: me, isLoading } = useMe();
  const tenant = me?.onboarded ? me.tenant : undefined;
  const isAdmin = me?.onboarded && me.role === "admin";
  const address = parseLogisticsAddress(tenant?.address);
  const missingFields = [
    ...(tenant?.contact_email?.trim() ? [] : ["logistics email"]),
    ...(tenant?.contact_phone?.trim() ? [] : ["pickup phone"]),
    ...missingLogisticsAddressFields(address),
  ];
  const needsSetup = !tenant?.active_modules.logistics;

  useEffect(() => {
    if (!isAdmin || !needsSetup) {
      return;
    }

    const timer = window.setTimeout(() => {
      router.replace("/app/settings/logistics");
    }, 1800);

    return () => window.clearTimeout(timer);
  }, [isAdmin, needsSetup, router]);

  if (isLoading) {
    return (
      <div className="card-3d flex min-h-80 flex-col items-center justify-center gap-3 rounded-2xl p-8 text-center">
        <SpinnerGapIcon className="size-5 animate-spin text-primary" />
        <p className="text-sm text-muted-foreground">Loading deliveries</p>
      </div>
    );
  }

  if (needsSetup) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">Deliveries</h1>
        </div>

        <div className="card-3d rounded-2xl p-8">
          <div className="flex items-start gap-3">
            <WarningCircleIcon className="mt-1 size-6 text-primary" weight="fill" />
            <div>
              <h2 className="text-xl font-semibold">Finish delivery setup first</h2>
              <p className="mt-2 text-sm text-muted-foreground">
                To start offering deliveries, the store needs a complete logistics profile with a
                clear pickup address, phone number, and logistics email.
              </p>
            </div>
          </div>

          {missingFields.length > 0 ? (
            <div className="mt-5 rounded-xl border border-border/60 bg-muted/40 p-4 text-sm text-muted-foreground">
              Missing: {missingFields.join(", ")}
            </div>
          ) : null}

          <div className="mt-6 flex flex-col gap-3 sm:flex-row sm:items-center">
            {isAdmin ? (
              <>
                <Link href="/app/settings/logistics" className="inline-flex">
                  <Button className="gap-2">
                    Open logistics setup
                    <ArrowRightIcon className="size-4" />
                  </Button>
                </Link>
                <p className="text-sm text-muted-foreground">
                  Redirecting you to logistics setup now.
                </p>
              </>
            ) : (
              <p className="text-sm text-muted-foreground">
                Ask an admin to complete logistics setup before deliveries can be used.
              </p>
            )}
          </div>
        </div>
      </div>
    );
  }

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
