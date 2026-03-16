"use client";

import Link from "next/link";
import { PlusIcon } from "@phosphor-icons/react";
import { Button } from "@/components/ui/button";
import { OpenBoxSvg } from "@/components/illustrations";

export default function ProductsPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Products</h1>
        <Link href="/app/products/new">
          <Button size="sm" className="gap-1.5">
            <PlusIcon className="size-4" weight="bold" />
            Add Product
          </Button>
        </Link>
      </div>
      <div className="card-3d rounded-2xl p-8 flex flex-col items-center justify-center text-center">
        <OpenBoxSvg className="size-36" />
        <p className="text-sm text-muted-foreground mt-3">
          Add your first product to get started
        </p>
        <Link href="/app/products/new" className="mt-3">
          <Button variant="outline" size="sm">Add product</Button>
        </Link>
      </div>
    </div>
  );
}
