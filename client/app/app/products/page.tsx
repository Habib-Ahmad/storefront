"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import {
  PlusIcon,
  MagnifyingGlassIcon,
  CaretLeftIcon,
  CaretRightIcon,
} from "@phosphor-icons/react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { OpenBoxSvg } from "@/components/illustrations";
import { useProducts } from "@/hooks/use-products";
import { ProductCard } from "./product-card";
import { ProductSkeleton } from "./product-skeleton";

const PRODUCTS_KNOWN_STORAGE_KEY = "storefront.products.has-items";

export default function ProductsPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [hasKnownProducts, setHasKnownProducts] = useState<boolean | null>(null);
  const perPage = 12;

  const { data, isLoading } = useProducts({ page, per_page: perPage });

  const products = data?.data ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.ceil(total / perPage);

  useEffect(() => {
    const storedValue = window.localStorage.getItem(PRODUCTS_KNOWN_STORAGE_KEY);
    setHasKnownProducts(storedValue === "1");
  }, []);

  useEffect(() => {
    if (isLoading) {
      return;
    }

    if (total > 0) {
      window.localStorage.setItem(PRODUCTS_KNOWN_STORAGE_KEY, "1");
      setHasKnownProducts(true);
      return;
    }

    window.localStorage.removeItem(PRODUCTS_KNOWN_STORAGE_KEY);
    setHasKnownProducts(false);
  }, [isLoading, total]);

  const filtered = search
    ? products.filter((p) => p.name.toLowerCase().includes(search.toLowerCase()))
    : products;
  const showSkeleton = isLoading && hasKnownProducts === true;
  const showEmptyState = (!isLoading && total === 0) || (isLoading && hasKnownProducts === false);

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
          <MagnifyingGlassIcon className="absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search products…"
            value={search}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSearch(e.target.value)}
            className="h-9 pl-8"
          />
        </div>
      )}

      {/* Loading */}
      {showSkeleton && (
        <div className="grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <ProductSkeleton key={i} />
          ))}
        </div>
      )}

      {/* Empty state */}
      {showEmptyState && (
        <div className="card-3d flex flex-col items-center justify-center rounded-2xl p-8 text-center">
          <OpenBoxSvg className="size-36" />
          <p className="mt-3 text-sm text-muted-foreground">
            Add your first product to get started
          </p>
          <Link href="/app/products/new" className="mt-3">
            <Button variant="outline" size="sm">
              Add product
            </Button>
          </Link>
        </div>
      )}

      {/* Product grid */}
      {!isLoading && filtered.length > 0 && (
        <div className="grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-4">
          {filtered.map((product) => (
            <ProductCard key={product.id} product={product} />
          ))}
        </div>
      )}

      {/* No search results */}
      {!isLoading && total > 0 && filtered.length === 0 && (
        <p className="py-8 text-center text-sm text-muted-foreground">
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
