"use client";

import { usePathname } from "next/navigation";
import Link from "next/link";
import { useTheme } from "next-themes";
import { MoonIcon, SunIcon, SignOutIcon, UserCircleIcon } from "@phosphor-icons/react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

const labels: Record<string, string> = {
  "/app": "Overview",
  "/app/orders": "Orders",
  "/app/orders/new": "New Order",
  "/app/products": "Products",
  "/app/products/new": "New Product",
  "/app/deliveries": "Deliveries",
  "/app/wallet": "Wallet",
  "/app/settings": "Settings",
  "/app/more": "More",
};

export function Header() {
  const pathname = usePathname();
  const { theme, setTheme } = useTheme();

  const segments = pathname.split("/").filter(Boolean);
  const crumbs = segments.map((_, i) => {
    const href = "/" + segments.slice(0, i + 1).join("/");
    const label = labels[href] ?? segments[i];
    return { href, label };
  });

  const pageTitle = crumbs.length > 0 ? crumbs[crumbs.length - 1].label : "Overview";

  return (
    <header className="flex items-center justify-between h-14 px-4 md:px-6 border-b border-border/50 glass sticky top-0 z-40">
      <div>
        <h1 className="text-base font-semibold md:hidden">{pageTitle}</h1>
        <nav className="hidden md:flex items-center gap-1.5 text-sm text-muted-foreground">
          {crumbs.map((crumb, i) => (
            <span key={crumb.href} className="flex items-center gap-1.5">
              {i > 0 && <span className="text-border">/</span>}
              {i === crumbs.length - 1 ? (
                <span className="font-medium text-foreground">
                  {crumb.label}
                </span>
              ) : (
                <Link
                  href={crumb.href}
                  className="hover:text-foreground transition-colors"
                >
                  {crumb.label}
                </Link>
              )}
            </span>
          ))}
        </nav>
      </div>

      <div className="flex items-center gap-1 md:hidden">
        <Button
          variant="ghost"
          size="icon"
          onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
        >
          <SunIcon className="size-4 rotate-0 scale-100 transition-transform dark:-rotate-90 dark:scale-0" weight="fill" />
          <MoonIcon className="absolute size-4 rotate-90 scale-0 transition-transform dark:rotate-0 dark:scale-100" weight="fill" />
          <span className="sr-only">Toggle theme</span>
        </Button>

        <DropdownMenu>
          <DropdownMenuTrigger
            className="rounded-full focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <Avatar size="sm">
              <AvatarFallback>U</AvatarFallback>
            </Avatar>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-44">
            <DropdownMenuItem render={<Link href="/app/settings" />}>
              <UserCircleIcon className="size-4" weight="fill" />
              Account
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => {}}
              className="text-destructive focus:text-destructive"
            >
              <SignOutIcon className="size-4" />
              Sign out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
