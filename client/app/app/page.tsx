"use client";

import { TrendUpIcon, ShoppingBagIcon, ChartPieIcon, CurrencyNgnIcon } from "@phosphor-icons/react";
import { ChartRiseSvg } from "@/components/illustrations";

const stats = [
  { label: "Revenue", value: "—", icon: CurrencyNgnIcon, color: "text-chart-5" },
  { label: "Orders", value: "—", icon: ShoppingBagIcon, color: "text-primary" },
  { label: "Profit", value: "—", icon: TrendUpIcon, color: "text-chart-4" },
  { label: "Avg. Order", value: "—", icon: ChartPieIcon, color: "text-chart-2" },
];

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Overview</h1>
      </div>

      <div className="grid gap-4 grid-cols-2 lg:grid-cols-4">
        {stats.map((stat) => (
          <div key={stat.label} className="card-3d rounded-2xl p-4 space-y-2">
            <div className="flex items-center justify-between">
              <p className="text-sm text-muted-foreground">{stat.label}</p>
              <stat.icon className={`size-5 ${stat.color}`} weight="fill" />
            </div>
            <p className="text-2xl font-bold">{stat.value}</p>
          </div>
        ))}
      </div>

      <div className="card-3d rounded-2xl p-6 flex flex-col items-center justify-center text-center">
        <ChartRiseSvg className="size-40" />
        <p className="text-sm text-muted-foreground mt-2">
          Sales analytics will appear here once you start selling
        </p>
      </div>
    </div>
  );
}
