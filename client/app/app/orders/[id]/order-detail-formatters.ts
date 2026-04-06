import type { FulfillmentStatus, PaymentStatus } from "@/lib/types";

export function formatCurrency(amount: string) {
  return new Intl.NumberFormat("en-NG", {
    style: "currency",
    currency: "NGN",
    minimumFractionDigits: 0,
  }).format(parseFloat(amount));
}

export function formatDate(value: string) {
  return new Intl.DateTimeFormat("en-NG", {
    day: "numeric",
    month: "short",
    year: "numeric",
  }).format(new Date(value));
}

export function paymentBadgeVariant(
  status: PaymentStatus,
): "default" | "secondary" | "destructive" {
  switch (status) {
    case "paid":
      return "default";
    case "failed":
    case "refunded":
      return "destructive";
    case "pending":
    default:
      return "secondary";
  }
}

export function fulfillmentBadgeVariant(
  status: FulfillmentStatus,
): "default" | "secondary" | "destructive" {
  switch (status) {
    case "completed":
    case "delivered":
    case "shipped":
      return "default";
    case "cancelled":
      return "destructive";
    case "processing":
    default:
      return "secondary";
  }
}
