"use client";

import Link from "next/link";
import {
  TrendUpIcon,
  ShoppingBagIcon,
  ChartPieIcon,
  CurrencyNgnIcon,
  PackageIcon,
  CaretRightIcon,
} from "@phosphor-icons/react";
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

      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {stats.map((stat) => (
          <div key={stat.label} className="card-3d space-y-2 rounded-2xl p-4">
            <div className="flex items-center justify-between">
              <p className="text-sm text-muted-foreground">{stat.label}</p>
              <stat.icon className={`size-5 ${stat.color}`} weight="fill" />
            </div>
            <p className="text-2xl font-bold">{stat.value}</p>
          </div>
        ))}
      </div>

      <div className="card-3d flex flex-col items-center justify-center rounded-2xl p-6 text-center">
        <ChartRiseSvg className="size-40" />
        <p className="mt-2 text-sm text-muted-foreground">
          Sales analytics will appear here once you start selling
        </p>
      </div>

      <div>
        <h2 className="mb-3 text-sm font-semibold tracking-wider text-muted-foreground uppercase">
          Quick access
        </h2>
        <div className="glass divide-y divide-border/50 rounded-2xl border border-border/50">
          <Link
            href="/app/products"
            className="flex items-center gap-3 px-4 py-3.5 transition-colors hover:bg-accent/50"
          >
            <PackageIcon className="size-4 text-muted-foreground" weight="fill" />
            <span className="flex-1 text-sm font-medium">Products</span>
            <CaretRightIcon className="size-4 text-muted-foreground" />
          </Link>
        </div>
      </div>
    </div>
  );
}
