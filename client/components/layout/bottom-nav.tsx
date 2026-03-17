"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  HouseIcon,
  ShoppingBagIcon,
  PackageIcon,
  DotsThreeCircleIcon,
  PlusIcon,
  TruckIcon,
} from "@phosphor-icons/react";
import { cn } from "@/lib/utils";

const leftItems = [
  { href: "/app", label: "Home", icon: HouseIcon },
  { href: "/app/orders", label: "Orders", icon: ShoppingBagIcon },
];

const rightItems = [
  { href: "/app/deliveries", label: "Deliveries", icon: TruckIcon },
  { href: "/app/more", label: "More", icon: DotsThreeCircleIcon },
];

export function BottomNav() {
  const pathname = usePathname();

  const renderItem = (item: (typeof leftItems)[0]) => {
    const isActive = item.href === "/app" ? pathname === "/app" : pathname.startsWith(item.href);
    return (
      <Link
        key={item.href}
        href={item.href}
        className={cn(
          "flex flex-1 flex-col items-center justify-center gap-0.5 py-3 text-[11px] transition-colors",
          isActive ? "text-primary" : "text-muted-foreground",
        )}
      >
        <item.icon className="size-5" weight={isActive ? "fill" : "regular"} />
        <span className={cn(isActive && "font-medium")}>{item.label}</span>
      </Link>
    );
  };

  return (
    <div
      className="fixed right-0 bottom-0 left-0 z-50 backdrop-blur-xl backdrop-saturate-150 md:hidden"
      style={{ paddingBottom: "env(safe-area-inset-bottom)" }}
    >
      <div className="relative h-16">
        {/* SVG notched background */}
        <svg
          aria-hidden
          className="pointer-events-none absolute inset-0 h-full w-full"
          viewBox="0 0 390 64"
          preserveAspectRatio="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          {/* Bar fill — circular notch centred on FAB (radius 28, centre at 195,8) */}
          <path
            d="M0,0 H168 A28,28 0 0 1 222,0 H390 V64 H0 Z"
            className="fill-background"
            fillOpacity="0.92"
          />
          {/* Top border following the notch arc */}
          <path
            d="M0,0.5 H168 A28,28 0 0 1 222,0.5 H390"
            fill="none"
            className="stroke-border"
            strokeWidth="1"
            vectorEffect="non-scaling-stroke"
          />
        </svg>

        {/* Nav items — left and right of FAB zone */}
        <div className="absolute inset-0 flex items-center">
          <div className="flex flex-1 justify-around">{leftItems.map(renderItem)}</div>
          {/* Spacer so items don't crowd the FAB zone */}
          <div className="w-20" />
          <div className="flex flex-1 justify-around">{rightItems.map(renderItem)}</div>
        </div>

        {/* FAB — rises above the bar, sits in the notch */}
        <Link
          href="/app/orders/new"
          className="absolute -top-4 left-1/2 z-10 flex size-12 -translate-x-1/2 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-xl shadow-primary/40 transition-transform active:scale-95"
        >
          <PlusIcon className="size-6" weight="bold" />
        </Link>
      </div>
    </div>
  );
}
