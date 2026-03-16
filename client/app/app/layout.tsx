import { Sidebar } from "@/components/layout/sidebar";
import { BottomNav } from "@/components/layout/bottom-nav";
import { Header } from "@/components/layout/header";

export default function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-background bg-mesh bg-dots">
      <div className="fixed inset-0 pointer-events-none" aria-hidden="true">
        <div className="absolute -top-40 -right-40 h-96 w-96 rounded-full bg-primary/8 blur-3xl" />
        <div className="absolute top-1/3 -left-40 h-[28rem] w-[28rem] rounded-full bg-chart-4/6 blur-3xl" />
        <div className="absolute bottom-0 right-1/4 h-80 w-80 rounded-full bg-chart-5/6 blur-3xl" />
        <div className="absolute top-2/3 left-1/3 h-64 w-64 rounded-full bg-chart-2/5 blur-3xl" />
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
