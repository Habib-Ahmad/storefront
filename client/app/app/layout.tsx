"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { SpinnerGapIcon } from "@phosphor-icons/react";
import { Sidebar } from "@/components/layout/sidebar";
import { BottomNav } from "@/components/layout/bottom-nav";
import { Header } from "@/components/layout/header";
import { ShoppingBagSvg } from "@/components/illustrations";
import { useSession } from "@/components/auth-provider";
import { useMe } from "@/hooks/use-auth";

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const { session, loading: sessionLoading } = useSession();
  const { data: me, isLoading: meLoading } = useMe();

  useEffect(() => {
    if (!sessionLoading && session && me && !me.onboarded) {
      window.sessionStorage.setItem("storefront:onboarding-banner", "app-guard");
      router.replace("/onboard");
    }
  }, [me, router, session, sessionLoading]);

  if (sessionLoading || meLoading || (session && me && !me.onboarded)) {
    return (
      <div className="bg-mesh bg-dots flex min-h-screen items-center justify-center bg-background px-4">
        <div className="card-3d w-full max-w-sm space-y-4 rounded-2xl p-6 text-center">
          <div className="flex justify-center">
            <ShoppingBagSvg className="size-24" />
          </div>
          <div className="space-y-1.5">
            <h1 className="text-xl font-semibold">Almost there</h1>
            <p className="text-sm text-muted-foreground">Preparing your workspace</p>
          </div>
          <SpinnerGapIcon className="mx-auto size-5 animate-spin text-primary" />
        </div>
      </div>
    );
  }

  return (
    <div className="bg-mesh bg-dots min-h-screen bg-background">
      <div className="pointer-events-none fixed inset-0" aria-hidden="true">
        <div className="absolute -top-40 -right-40 h-96 w-96 rounded-full bg-primary/8 blur-3xl" />
        <div className="absolute top-1/3 -left-40 h-112 w-md rounded-full bg-chart-4/6 blur-3xl" />
        <div className="absolute right-1/4 bottom-0 h-80 w-80 rounded-full bg-chart-5/6 blur-3xl" />
        <div className="absolute top-2/3 left-1/3 h-64 w-64 rounded-full bg-chart-2/5 blur-3xl" />
      </div>
      <Sidebar />
      <div className="relative md:pl-64">
        <Header />
        <main className="p-4 pb-20 md:p-6 md:pb-6">{children}</main>
      </div>
      <BottomNav />
    </div>
  );
}
