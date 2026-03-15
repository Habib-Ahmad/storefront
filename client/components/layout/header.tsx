"use client";

import { usePathname } from "next/navigation";
import Link from "next/link";

const labels: Record<string, string> = {
  "/app": "Overview",
  "/app/orders": "Orders",
  "/app/orders/new": "New Order",
  "/app/products": "Products",
  "/app/products/new": "New Product",
  "/app/wallet": "Wallet",
  "/app/settings": "Settings",
};

export function Header() {
  const pathname = usePathname();

  // Build breadcrumb segments
  const segments = pathname.split("/").filter(Boolean);
  const crumbs = segments.map((_, i) => {
    const href = "/" + segments.slice(0, i + 1).join("/");
    const label = labels[href] ?? segments[i];
    return { href, label };
  });

  return (
    <header className="flex items-center h-14 px-4 md:px-6 border-b glass sticky top-0 z-40">
      <nav className="flex items-center gap-1.5 text-sm text-muted-foreground">
        {crumbs.map((crumb, i) => (
          <span key={crumb.href} className="flex items-center gap-1.5">
            {i > 0 && <span>/</span>}
            {i === crumbs.length - 1 ? (
              <span className="font-medium text-foreground">
                {crumb.label}
              </span>
            ) : (
              <Link href={crumb.href} className="hover:text-foreground">
                {crumb.label}
              </Link>
            )}
          </span>
        ))}
      </nav>
    </header>
  );
}
