import Link from "next/link";
import { TagIcon } from "@phosphor-icons/react";
import { Badge } from "@/components/ui/badge";
import type { Product } from "@/lib/types";
import { formatPrice, totalStock } from "./product-formatters";

interface ProductCardProps {
  product: Product;
}

export function ProductCard({ product }: ProductCardProps) {
  const stock = totalStock(product.variants);
  const primaryImage = product.images?.find((image) => image.is_primary) ?? product.images?.[0];

  return (
    <Link href={`/app/products/${product.id}`} className="block">
      <div className="card-3d overflow-hidden rounded-2xl transition-all hover:ring-2 hover:ring-primary/20">
        <div className="flex aspect-square items-center justify-center bg-muted">
          {primaryImage ? (
            <img src={primaryImage.url} alt={product.name} className="size-full object-cover" />
          ) : (
            <TagIcon className="size-10 text-muted-foreground/40" />
          )}
        </div>
        <div className="space-y-1.5 p-3">
          <p className="truncate text-sm font-medium">{product.name}</p>
          <p className="text-sm font-semibold text-primary">{formatPrice(product.variants)}</p>
          <div className="flex items-center gap-2">
            <Badge variant={product.is_available ? "default" : "secondary"} className="text-xs">
              {product.is_available ? "Active" : "Draft"}
            </Badge>
            {stock !== null && (
              <span
                className={`text-xs ${stock === 0 ? "text-destructive" : "text-muted-foreground"}`}
              >
                {stock === 0 ? "Out of stock" : `${stock} in stock`}
              </span>
            )}
          </div>
        </div>
      </div>
    </Link>
  );
}
