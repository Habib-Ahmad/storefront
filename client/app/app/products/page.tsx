"use client";

import { useState } from "react";
import Link from "next/link";
import {
  PlusIcon,
  MagnifyingGlassIcon,
  TagIcon,
  CaretLeftIcon,
  CaretRightIcon,
} from "@phosphor-icons/react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { OpenBoxSvg } from "@/components/illustrations";
import { useProducts } from "@/hooks/use-products";
import type { Product } from "@/lib/types";

function formatPrice(variants?: Product["variants"]) {
  if (!variants || variants.length === 0) return "—";
  const prices = variants.map((v) => parseFloat(v.price));
  const min = Math.min(...prices);
  const max = Math.max(...prices);
  const fmt = (n: number) =>
    new Intl.NumberFormat("en-NG", { style: "currency", currency: "NGN", minimumFractionDigits: 0 }).format(n);
  return min === max ? fmt(min) : `${fmt(min)} – ${fmt(max)}`;
}

function totalStock(variants?: Product["variants"]) {
  if (!variants || variants.length === 0) return null;
  if (variants.some((v) => v.stock_qty === null || v.stock_qty === undefined)) return null;
  return variants.reduce((sum, v) => sum + (v.stock_qty ?? 0), 0);
}

function ProductCard({ product }: { product: Product }) {
  const stock = totalStock(product.variants);
  const primary = product.images?.find((i) => i.is_primary) ?? product.images?.[0];

  return (
    <Link href={`/app/products/${product.id}`} className="block">
      <div className="card-3d rounded-2xl overflow-hidden hover:ring-2 hover:ring-primary/20 transition-all">
        <div className="aspect-square bg-muted flex items-center justify-center">
          {primary ? (
            <img
              src={primary.url}
              alt={product.name}
              className="size-full object-cover"
            />
          ) : (
            <TagIcon className="size-10 text-muted-foreground/40" />
          )}
        </div>
        <div className="p-3 space-y-1.5">
          <p className="text-sm font-medium truncate">{product.name}</p>
          <p className="text-sm font-semibold text-primary">
            {formatPrice(product.variants)}
          </p>
          <div className="flex items-center gap-2">
            <Badge variant={product.is_available ? "default" : "secondary"} className="text-xs">
              {product.is_available ? "Active" : "Draft"}
            </Badge>
            {stock !== null && (
              <span className={`text-xs ${stock === 0 ? "text-destructive" : "text-muted-foreground"}`}>
                {stock === 0 ? "Out of stock" : `${stock} in stock`}
              </span>
            )}
          </div>
        </div>
      </div>
    </Link>
  );
}

function ProductSkeleton() {
  return (
    <div className="card-3d rounded-2xl overflow-hidden">
      <Skeleton className="aspect-square" />
      <div className="p-3 space-y-2">
        <Skeleton className="h-4 w-3/4" />
        <Skeleton className="h-4 w-1/2" />
        <Skeleton className="h-5 w-16 rounded-full" />
      </div>
    </div>
  );
}

export default function ProductsPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const perPage = 12;

  const { data, isLoading } = useProducts({ page, per_page: perPage });

  const products = data?.data ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.ceil(total / perPage);

  const filtered = search
    ? products.filter((p) => p.name.toLowerCase().includes(search.toLowerCase()))
    : products;

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Products</h1>
        <Link href="/app/products/new">
          <Button size="sm" className="gap-1.5">
            <PlusIcon className="size-4" weight="bold" />
            Add Product
          </Button>
        </Link>
      </div>

      {/* Search */}
      {!isLoading && total > 0 && (
        <div className="relative max-w-sm">
          <MagnifyingGlassIcon className="absolute left-2.5 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <Input
            placeholder="Search products…"
            value={search}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSearch(e.target.value)}
            className="pl-8 h-9"
          />
        </div>
      )}

      {/* Loading */}
      {isLoading && (
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <ProductSkeleton key={i} />
          ))}
        </div>
      )}

      {/* Empty state */}
      {!isLoading && total === 0 && (
        <div className="card-3d rounded-2xl p-8 flex flex-col items-center justify-center text-center">
          <OpenBoxSvg className="size-36" />
          <p className="text-sm text-muted-foreground mt-3">
            Add your first product to get started
          </p>
          <Link href="/app/products/new" className="mt-3">
            <Button variant="outline" size="sm">Add product</Button>
          </Link>
        </div>
      )}

      {/* Product grid */}
      {!isLoading && filtered.length > 0 && (
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
          {filtered.map((product) => (
            <ProductCard key={product.id} product={product} />
          ))}
        </div>
      )}

      {/* No search results */}
      {!isLoading && total > 0 && filtered.length === 0 && (
        <p className="text-sm text-muted-foreground text-center py-8">
          No products matching &ldquo;{search}&rdquo;
        </p>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 pt-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1}
            onClick={() => setPage((p) => p - 1)}
          >
            <CaretLeftIcon className="size-4" />
          </Button>
          <span className="text-sm text-muted-foreground">
            {page} / {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages}
            onClick={() => setPage((p) => p + 1)}
          >
            <CaretRightIcon className="size-4" />
          </Button>
        </div>
      )}
    </div>
  );
}
