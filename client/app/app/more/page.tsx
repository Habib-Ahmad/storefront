"use client";

import Link from "next/link";
import { useTheme } from "next-themes";
import { WalletIcon, GearSixIcon, MoonIcon, SunIcon, SignOutIcon, CaretRightIcon } from "@phosphor-icons/react";
import { Separator } from "@/components/ui/separator";
import { Button } from "@/components/ui/button";

const links = [
  { href: "/app/wallet", label: "Wallet", icon: WalletIcon },
  { href: "/app/settings", label: "Settings", icon: GearSixIcon },
];

export default function MorePage() {
  const { theme, setTheme } = useTheme();

  return (
    <div className="space-y-6">
      <h1 className="text-xl font-bold">More</h1>

      <div className="glass rounded-2xl border border-border/50 divide-y divide-border/50">
        {links.map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className="flex items-center gap-3 px-4 py-3.5 transition-colors hover:bg-accent/50"
          >
            <item.icon className="size-4 text-muted-foreground" weight="fill" />
            <span className="flex-1 text-sm font-medium">{item.label}</span>
            <CaretRightIcon className="size-4 text-muted-foreground" />
          </Link>
        ))}
      </div>

      <div className="glass rounded-2xl border border-border/50 divide-y divide-border/50">
        <button
          onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
          className="flex items-center gap-3 w-full px-4 py-3.5 transition-colors hover:bg-accent/50 text-left"
        >
          <div className="relative size-4 text-muted-foreground">
           <SunIcon className="size-4 rotate-0 scale-100 transition-transform dark:-rotate-90 dark:scale-0" weight="fill" />
           <MoonIcon className="absolute inset-0 size-4 rotate-90 scale-0 transition-transform dark:rotate-0 dark:scale-100" weight="fill" />
          </div>
          <span className="flex-1 text-sm font-medium">Dark mode</span>
          <span className="text-xs text-muted-foreground capitalize">{theme}</span>
        </button>

        <Separator className="my-0!" />

        <button
          onClick={() => {}}
          className="flex items-center gap-3 w-full px-4 py-3.5 transition-colors hover:bg-accent/50 text-left text-destructive"
        >
          <SignOutIcon className="size-4" />
          <span className="flex-1 text-sm font-medium">Sign out</span>
        </button>
      </div>
    </div>
  );
}
