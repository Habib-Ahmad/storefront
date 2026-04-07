"use client";

import { Moon, Sun } from "lucide-react";
import { useEffect, useState } from "react";
import { useTheme } from "next-themes";
import { Button } from "@/components/ui/button";
import { StorefrontBasketLink } from "@/app/[slug]/storefront-basket-link";

interface PublicStorefrontActionsProps {
  slug?: string;
}

export function PublicStorefrontActions({ slug }: PublicStorefrontActionsProps) {
  const { resolvedTheme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const isDark = resolvedTheme === "dark";

  return (
    <div className="flex items-center gap-2">
      {slug ? <StorefrontBasketLink slug={slug} /> : null}
      <Button
        type="button"
        variant="outline"
        size="icon-sm"
        className="rounded-full border-border/70 bg-card"
        onClick={() => setTheme(isDark ? "light" : "dark")}
        disabled={!mounted}
        aria-label="Toggle color theme"
      >
        {mounted ? (
          isDark ? (
            <Sun className="h-4 w-4" />
          ) : (
            <Moon className="h-4 w-4" />
          )
        ) : (
          <Sun className="h-4 w-4 opacity-0" />
        )}
      </Button>
    </div>
  );
}
