"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Home,
  ShoppingCart,
  Package,
  Wallet,
  Settings,
  PlusCircle,
} from "lucide-react";
import { cn } from "@/lib/utils";

const items = [
  { href: "/app", label: "Overview", icon: Home },
  { href: "/app/orders", label: "Orders", icon: ShoppingCart },
  { href: "/app/products", label: "Products", icon: Package },
  { href: "/app/wallet", label: "Wallet", icon: Wallet },
  { href: "/app/settings", label: "Settings", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="hidden md:flex md:flex-col md:w-64 md:border-r md:h-screen md:fixed md:left-0 md:top-0 glass">
      {/* Logo / brand */}
      <div className="flex items-center h-16 px-6 border-b">
        <Link href="/app" className="text-lg font-semibold">
          Storefront
        </Link>
      </div>

      {/* Nav links */}
      <nav className="flex-1 px-3 py-4 space-y-1">
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
                "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors",
                isActive
                  ? "bg-accent text-accent-foreground"
                  : "text-muted-foreground hover:bg-accent/50 hover:text-foreground",
              )}
            >
              <item.icon className="w-5 h-5" />
              {item.label}
            </Link>
          );
        })}
      </nav>

      {/* Quick action */}
      <div className="p-4 border-t">
        <Link
          href="/app/orders/new"
          className="flex items-center justify-center gap-2 w-full rounded-md bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground"
        >
          <PlusCircle className="w-4 h-4" />
          New Order
        </Link>
      </div>
    </aside>
  );
}
