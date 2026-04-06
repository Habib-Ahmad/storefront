export function formatCurrency(amount: string) {
  return new Intl.NumberFormat("en-NG", {
    style: "currency",
    currency: "NGN",
    maximumFractionDigits: 0,
  }).format(Number(amount));
}

export function getInitials(name: string) {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? "")
    .join("");
}

export function getProductDescription(description?: string | null) {
  if (!description) {
    return "Ask the store for more details about this item.";
  }

  return description.length > 120 ? `${description.slice(0, 117)}...` : description;
}
