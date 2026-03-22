"use client";

import { usePathname } from "next/navigation";
import Link from "next/link";
import { useTheme } from "next-themes";
import { MoonIcon, SunIcon, SignOutIcon, UserCircleIcon } from "@phosphor-icons/react";
import { useSignOut } from "@/hooks/use-auth";
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

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

export function Header() {
  const pathname = usePathname();
  const { theme, setTheme } = useTheme();
  const signOut = useSignOut();

  const segments = pathname.split("/").filter(Boolean);
  const crumbs = segments.map((_, i) => {
    const href = "/" + segments.slice(0, i + 1).join("/");
    const raw = segments[i];
    const label = labels[href] ?? (UUID_RE.test(raw) ? "Details" : raw);
    return { href, label };
  });

  return (
    <header className="glass sticky top-0 z-40 flex h-14 items-center justify-between border-b border-border/50 px-4 backdrop-blur-xl backdrop-saturate-150 md:px-6">
      <div>
        {/* Mobile: no title — each page renders its own h1 */}
        <nav className="hidden items-center gap-1.5 text-sm text-muted-foreground md:flex">
          {crumbs.map((crumb, i) => (
            <span key={crumb.href} className="flex items-center gap-1.5">
              {i > 0 && <span className="text-border">/</span>}
              {i === crumbs.length - 1 ? (
                <span className="font-medium text-foreground">{crumb.label}</span>
              ) : (
                <Link href={crumb.href} className="transition-colors hover:text-foreground">
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
          <SunIcon
            className="size-4 scale-100 rotate-0 transition-transform dark:scale-0 dark:-rotate-90"
            weight="fill"
          />
          <MoonIcon
            className="absolute size-4 scale-0 rotate-90 transition-transform dark:scale-100 dark:rotate-0"
            weight="fill"
          />
          <span className="sr-only">Toggle theme</span>
        </Button>

        <DropdownMenu>
          <DropdownMenuTrigger className="rounded-full focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none">
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
            <DropdownMenuItem onClick={signOut} className="text-destructive focus:text-destructive">
              <SignOutIcon className="size-4" />
              Sign out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
