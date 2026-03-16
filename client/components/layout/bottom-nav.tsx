"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { HouseIcon, ShoppingBagIcon, PackageIcon, TruckIcon, DotsThreeCircleIcon } from "@phosphor-icons/react";
import { cn } from "@/lib/utils";

const items = [
  { href: "/app", label: "Home", icon: HouseIcon },
  { href: "/app/orders", label: "Orders", icon: ShoppingBagIcon },
  { href: "/app/products", label: "Products", icon: PackageIcon },
  { href: "/app/deliveries", label: "Deliveries", icon: TruckIcon },
  { href: "/app/more", label: "More", icon: DotsThreeCircleIcon },
];

export function BottomNav() {
  const pathname = usePathname();

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 border-t border-border/50 glass pb-[env(safe-area-inset-bottom)] md:hidden">
      <div className="flex items-center justify-around h-16">
        {items.map((item) => {
          const isActive =
            item.href === "/app"
              ? pathname === "/app"
              : pathname.startsWith(item.href);

          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex flex-col items-center justify-center gap-0.5 w-16 py-1.5 text-[11px] transition-colors",
                isActive
                  ? "text-primary"
                  : "text-muted-foreground active:text-foreground",
              )}
            >
              <item.icon
                className="size-5"
                weight={isActive ? "fill" : "regular"}
              />
              <span className={cn(isActive && "font-medium")}>{item.label}</span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
