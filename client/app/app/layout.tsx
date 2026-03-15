import { Sidebar } from "@/components/layout/sidebar";
import { BottomNav } from "@/components/layout/bottom-nav";
import { Header } from "@/components/layout/header";

export default function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-background">
      {/* Subtle gradient orbs — visible through frosted glass */}
      <div className="fixed inset-0 pointer-events-none" aria-hidden="true">
        <div className="absolute -top-40 -right-40 h-80 w-80 rounded-full bg-primary/[0.07] blur-3xl" />
        <div className="absolute top-1/2 -left-40 h-96 w-96 rounded-full bg-primary/[0.05] blur-3xl" />
      </div>
      <Sidebar />
      <div className="md:pl-64 relative">
        <Header />
        <main className="p-4 pb-20 md:p-6 md:pb-6">{children}</main>
      </div>
      <BottomNav />
    </div>
  );
}
