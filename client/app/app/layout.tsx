import { Sidebar } from "@/components/layout/sidebar";
import { BottomNav } from "@/components/layout/bottom-nav";
import { Header } from "@/components/layout/header";

export default function AppLayout({ children }: { children: React.ReactNode }) {
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
