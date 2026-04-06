import type { Order } from "@/lib/types";

export function formatCurrency(amount: string) {
  return new Intl.NumberFormat("en-NG", {
    style: "currency",
    currency: "NGN",
    minimumFractionDigits: 0,
  }).format(parseFloat(amount));
}

export function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("en-NG", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
    hour12: true,
    timeZone: "Africa/Lagos",
  }).format(new Date(value));
}

type CardBadge = {
  label: string;
  variant: "default" | "secondary" | "destructive";
};

function primaryStatusBadge(order: Order): CardBadge {
  if (order.fulfillment_status === "cancelled") {
    return { label: "Cancelled", variant: "destructive" };
  }

  if (order.payment_status === "failed") {
    return { label: "Payment failed", variant: "destructive" };
  }

  if (order.payment_status === "refunded") {
    return { label: "Refunded", variant: "destructive" };
  }

  if (order.fulfillment_status === "completed") {
    return { label: "Completed", variant: "default" };
  }

  if (order.fulfillment_status === "delivered") {
    return { label: "Delivered", variant: "default" };
  }

  if (order.fulfillment_status === "shipped") {
    return { label: "Shipped", variant: "default" };
  }

  if (order.payment_status === "pending") {
    return { label: "Awaiting payment", variant: "secondary" };
  }

  if (order.is_delivery && order.fulfillment_status === "processing") {
    return { label: "Ready for delivery", variant: "secondary" };
  }

  return { label: "Processing", variant: "secondary" };
}

export function cardBadges(order: Order): CardBadge[] {
  const badges = [primaryStatusBadge(order)];

  if (order.is_delivery) {
    badges.push({ label: "Delivery", variant: "secondary" });
  }

  return badges;
}

export type { CardBadge };
