"use client";

import { useEffect, useMemo } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  ArrowRightIcon,
  ArrowSquareOutIcon,
  SpinnerGapIcon,
  WarningCircleIcon,
} from "@phosphor-icons/react";
import { useSession } from "@/components/auth-provider";
import { DeliveryTruckSvg } from "@/components/illustrations";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useMe } from "@/hooks/use-auth";
import { useOrders } from "@/hooks/use-orders";
import { missingLogisticsAddressFields, parseLogisticsAddress } from "@/lib/logistics-address";
import {
  cardBadges,
  displayCustomer,
  formatCurrency,
  formatDateTime,
} from "../orders/order-formatters";

export default function DeliveriesPage() {
  const router = useRouter();
  const { loading: authLoading } = useSession();
  const { data: me, isLoading, error: meError } = useMe();
  const {
    data: ordersResponse,
    isLoading: ordersLoading,
    error: ordersError,
  } = useOrders({ page: 1, per_page: 50, view: "all" }, { enabled: me?.onboarded === true });
  const tenant = me?.onboarded ? me.tenant : undefined;
  const isAdmin = me?.onboarded && me.role === "admin";
  const address = parseLogisticsAddress(tenant?.address);
  const missingFields = [
    ...(tenant?.contact_email?.trim() ? [] : ["logistics email"]),
    ...(tenant?.contact_phone?.trim() ? [] : ["pickup phone"]),
    ...missingLogisticsAddressFields(address),
  ];
  const needsSetup = !tenant?.active_modules.logistics;
  const deliveryOrders = useMemo(
    () => (ordersResponse?.data ?? []).filter((order) => order.is_delivery),
    [ordersResponse?.data],
  );
  const summary = useMemo(
    () => [
      {
        label: "Ready to dispatch",
        value: deliveryOrders.filter(
          (order) => order.payment_status === "paid" && order.fulfillment_status === "processing",
        ).length,
      },
      {
        label: "Awaiting payment",
        value: deliveryOrders.filter((order) => order.payment_status === "pending").length,
      },
      {
        label: "In transit",
        value: deliveryOrders.filter((order) => order.fulfillment_status === "shipped").length,
      },
      {
        label: "Delivered",
        value: deliveryOrders.filter((order) => order.fulfillment_status === "delivered").length,
      },
    ],
    [deliveryOrders],
  );

  useEffect(() => {
    if (!isAdmin || !needsSetup) {
      return;
    }

    const timer = window.setTimeout(() => {
      router.replace("/app/settings/logistics");
    }, 1800);

    return () => window.clearTimeout(timer);
  }, [isAdmin, needsSetup, router]);

  if (authLoading || isLoading || (me?.onboarded === true && ordersLoading)) {
    return (
      <div className="card-3d flex min-h-80 flex-col items-center justify-center gap-3 rounded-2xl p-8 text-center">
        <SpinnerGapIcon className="size-5 animate-spin text-primary" />
        <p className="text-sm text-muted-foreground">Loading deliveries</p>
      </div>
    );
  }

  if (meError || ordersError) {
    const loadError = meError ?? ordersError;

    return (
      <div className="card-3d rounded-2xl p-8">
        <div className="flex items-start gap-3">
          <WarningCircleIcon className="mt-1 size-6 text-primary" weight="fill" />
          <div>
            <h1 className="text-2xl font-bold">Deliveries</h1>
            <p className="mt-2 text-sm text-muted-foreground">
              Unable to load delivery activity right now.
            </p>
            <p className="mt-1 text-sm text-muted-foreground">
              {loadError instanceof Error ? loadError.message : "Unexpected deliveries error"}
            </p>
          </div>
        </div>
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

      <div className="grid gap-4 md:grid-cols-4">
        {summary.map((item) => (
          <div key={item.label} className="card-3d rounded-2xl p-5">
            <p className="text-sm text-muted-foreground">{item.label}</p>
            <p className="mt-1 text-3xl font-bold">{item.value}</p>
          </div>
        ))}
      </div>

      {deliveryOrders.length === 0 ? (
        <div className="card-3d flex flex-col items-center justify-center rounded-2xl p-8 text-center">
          <DeliveryTruckSvg className="size-36" />
          <p className="mt-3 text-sm text-muted-foreground">
            Delivery orders will appear here once customers choose delivery at checkout.
          </p>
          <Link href="/app/orders" className="mt-4 inline-flex">
            <Button variant="outline" className="gap-2">
              View all orders
              <ArrowRightIcon className="size-4" />
            </Button>
          </Link>
        </div>
      ) : (
        <div className="space-y-3">
          {deliveryOrders.map((order) => {
            const customerName = displayCustomer(order);
            const badges = cardBadges(order);

            return (
              <Link key={order.id} href={`/app/orders/${order.id}`} className="block">
                <div className="card-3d space-y-3 rounded-2xl p-4 transition-all hover:ring-2 hover:ring-primary/20">
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                    <div className="min-w-0 space-y-1">
                      <p className="truncate text-sm font-semibold">{customerName}</p>
                      <p className="text-xs text-muted-foreground">
                        {formatDateTime(order.created_at)}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        Tracking {order.tracking_slug}
                      </p>
                    </div>

                    <div className="flex items-start gap-3">
                      <div className="text-left sm:text-right">
                        <p className="text-base font-semibold text-primary">
                          {formatCurrency(order.total_amount)}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          Shipping fee {formatCurrency(order.shipping_fee)}
                        </p>
                      </div>
                      <ArrowSquareOutIcon className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
                    </div>
                  </div>

                  <div className="flex flex-wrap items-center gap-2">
                    {badges.map((badge) => (
                      <Badge key={badge.label} variant={badge.variant} className="text-xs">
                        {badge.label}
                      </Badge>
                    ))}
                  </div>

                  <p className="text-sm text-muted-foreground">
                    {order.fulfillment_status === "processing" && order.payment_status === "paid"
                      ? "Ready for courier selection and dispatch."
                      : order.fulfillment_status === "shipped"
                        ? "Already dispatched and awaiting final delivery."
                        : order.fulfillment_status === "delivered"
                          ? "Delivery completed."
                          : "Open the order to manage the current delivery state."}
                  </p>
                </div>
              </Link>
            );
          })}
        </div>
      )}
    </div>
  );
}
