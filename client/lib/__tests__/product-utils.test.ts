import { describe, it, expect } from "vitest";
import { formatPrice, totalStock } from "../product-utils";
import type { ProductVariant } from "../types";

function v(price: string, stock_qty?: number | null): ProductVariant {
  return {
    id: "v1",
    product_id: "p1",
    sku: "S",
    attributes: {},
    price,
    is_default: true,
    created_at: "",
    updated_at: "",
    stock_qty,
  };
}

describe("formatPrice", () => {
  it("returns — with no variants", () => expect(formatPrice([])).toBe("—"));

  it("formats a single price in NGN", () => expect(formatPrice([v("5000")])).toBe("₦5,000"));

  it("returns a range when prices differ, min first", () => {
    const result = formatPrice([v("9000"), v("500"), v("3000")]);
    expect(result).toContain("500");
    expect(result).toContain("9,000");
  });
});

describe("totalStock", () => {
  it("returns null when any variant has unknown stock", () => {
    expect(totalStock([v("100", 5), v("200", null)])).toBeNull();
  });

  it("sums stock across all variants", () => {
    expect(totalStock([v("100", 3), v("200", 7)])).toBe(10);
  });
});
