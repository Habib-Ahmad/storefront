import Link from "next/link";
import { ArrowSquareOutIcon } from "@phosphor-icons/react";
import { Badge } from "@/components/ui/badge";
import type { Order } from "@/lib/types";
import { cardBadges, formatCurrency, formatDateTime } from "./order-formatters";

interface OrderCardProps {
  order: Order;
}

export function OrderCard({ order }: OrderCardProps) {
  const customerName = order.customer_name?.trim() || "Walk-in customer";
  const badges = cardBadges(order);

  return (
    <Link href={`/app/orders/${order.id}`} className="block">
      <div className="card-3d space-y-3 rounded-2xl p-4 transition-all hover:ring-2 hover:ring-primary/20">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 space-y-1">
            <p className="truncate text-sm font-semibold">{customerName}</p>
            <p className="text-xs text-muted-foreground">{formatDateTime(order.created_at)}</p>
          </div>
          <div className="flex items-start gap-3">
            <div className="text-right">
              <p className="text-base font-semibold text-primary">
                {formatCurrency(order.total_amount)}
              </p>
            </div>
            <ArrowSquareOutIcon className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {badges.map((badge) => (
            <Badge key={badge.label} variant={badge.variant} className="text-xs">
              {badge.label}
            </Badge>
          ))}
        </div>
      </div>
    </Link>
  );
}
