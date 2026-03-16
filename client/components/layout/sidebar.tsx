"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTheme } from "next-themes";
import {
  HouseIcon,
  ShoppingBagIcon,
  PackageIcon,
  TruckIcon,
  WalletIcon,
  GearSixIcon,
  PlusCircleIcon,
  MoonIcon,
  SunIcon,
  SignOutIcon,
  UserCircleIcon,
} from "@phosphor-icons/react";
import { useSignOut } from "@/hooks/use-auth";
import { cn } from "@/lib/utils";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

const navItems = [
  { href: "/app", label: "Overview", icon: HouseIcon },
  { href: "/app/orders", label: "Orders", icon: ShoppingBagIcon },
  { href: "/app/products", label: "Products", icon: PackageIcon },
  { href: "/app/deliveries", label: "Deliveries", icon: TruckIcon },
  { href: "/app/wallet", label: "Wallet", icon: WalletIcon },
  { href: "/app/settings", label: "Settings", icon: GearSixIcon },
];

export function Sidebar() {
  const pathname = usePathname();
  const { theme, setTheme } = useTheme();
  const signOut = useSignOut();

  return (
    <aside className="hidden md:flex md:flex-col md:w-64 md:border-r md:h-screen md:fixed md:left-0 md:top-0 glass z-30">
      <div className="flex items-center h-14 px-5 border-b border-border/50">
        <Link href="/app" className="text-lg font-bold tracking-tight">
          Storefront
        </Link>
      </div>

      <nav className="flex-1 px-3 py-3 space-y-0.5 overflow-y-auto">
        {navItems.map((item) => {
          const isActive =
            item.href === "/app"
              ? pathname === "/app"
              : pathname.startsWith(item.href);

          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                isActive
                  ? "bg-primary/10 text-primary"
                  : "text-muted-foreground hover:bg-accent hover:text-foreground",
              )}
            >
              <item.icon className="size-4" weight={isActive ? "fill" : "regular"} />
              {item.label}
            </Link>
          );
        })}
      </nav>

      <div className="px-3 pb-2">
        <Link href="/app/orders/new">
          <Button className="w-full gap-2" size="lg">
            <PlusCircleIcon className="size-4" weight="fill" />
            New Order
          </Button>
        </Link>
      </div>

      <div className="flex items-center gap-2 px-3 py-3 border-t border-border/50">
        <DropdownMenu>
          <DropdownMenuTrigger
            className="flex items-center gap-2.5 flex-1 rounded-lg px-2 py-1.5 text-sm hover:bg-accent transition-colors text-left"
          >
            <Avatar size="sm">
              <AvatarFallback>U</AvatarFallback>
            </Avatar>
            <div className="flex-1 truncate">
              <p className="font-medium text-xs truncate">My Store</p>
              <p className="text-[11px] text-muted-foreground truncate">owner</p>
            </div>
          </DropdownMenuTrigger>
          <DropdownMenuContent side="top" align="start" className="w-48">
            <DropdownMenuItem render={<Link href="/app/settings" />}>
              <UserCircleIcon className="size-4" weight="fill" />
              Account
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={signOut}
              className="text-destructive focus:text-destructive"
            >
              <SignOutIcon className="size-4" />
              Sign out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        <Button
          variant="ghost"
          size="icon"
          className="shrink-0"
          onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
        >
          <SunIcon className="size-4 rotate-0 scale-100 transition-transform dark:-rotate-90 dark:scale-0" weight="fill" />
          <MoonIcon className="absolute size-4 rotate-90 scale-0 transition-transform dark:rotate-0 dark:scale-100" weight="fill" />
          <span className="sr-only">Toggle theme</span>
        </Button>
      </div>
    </aside>
  );
}
