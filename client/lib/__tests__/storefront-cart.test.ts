import { beforeEach, describe, expect, it } from "vitest";
import {
  addStorefrontCartItem,
  getStorefrontCart,
  updateStorefrontCartItemQuantity,
} from "../storefront-cart";

const slug = "funke-fabrics";

beforeEach(() => {
  window.localStorage.clear();
});

describe("storefront cart", () => {
  it("merges duplicate variants into one cart line", () => {
    addStorefrontCartItem(slug, {
      productId: "product-1",
      productName: "Ankara Set",
      variantId: "variant-1",
      variantLabel: "Blue / M",
      unitPrice: "24500",
      quantity: 1,
      imageUrl: null,
    });

    addStorefrontCartItem(slug, {
      productId: "product-1",
      productName: "Ankara Set",
      variantId: "variant-1",
      variantLabel: "Blue / M",
      unitPrice: "24500",
      quantity: 2,
      imageUrl: null,
    });

    expect(getStorefrontCart(slug)).toEqual([
      expect.objectContaining({ variantId: "variant-1", quantity: 3 }),
    ]);
  });

  it("removes a line when quantity is reduced to zero", () => {
    addStorefrontCartItem(slug, {
      productId: "product-1",
      productName: "Ankara Set",
      variantId: "variant-1",
      variantLabel: "Blue / M",
      unitPrice: "24500",
      quantity: 1,
      imageUrl: null,
    });

    updateStorefrontCartItemQuantity(slug, "variant-1", 0);

    expect(getStorefrontCart(slug)).toEqual([]);
  });
});
