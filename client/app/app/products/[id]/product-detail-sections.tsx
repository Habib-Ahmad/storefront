import {
  FloppyDiskIcon,
  PencilSimpleIcon,
  PlusIcon,
  SpinnerGapIcon,
  TagIcon,
  TrashIcon,
} from "@phosphor-icons/react";
import { Field, Form } from "formik";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import type { ProductVariant } from "@/lib/types";
import { formatCurrency } from "./product-detail-formatters";

interface ProductDetailsCardProps {
  editing: boolean;
  isSubmitting: boolean;
  errors: Record<string, string | undefined>;
  touched: Record<string, boolean | undefined>;
  onEdit: () => void;
  onCancel: () => void;
}

export function ProductDetailsCard({
  editing,
  isSubmitting,
  errors,
  touched,
  onEdit,
  onCancel,
}: ProductDetailsCardProps) {
  return (
    <Form>
      <div className="card-3d space-y-4 rounded-2xl p-5">
        <div className="flex items-center justify-between">
          <h2 className="text-base font-semibold">Details</h2>
          {!editing ? (
            <Button type="button" variant="ghost" size="sm" onClick={onEdit} className="gap-1.5">
              <PencilSimpleIcon className="size-4" />
              Edit
            </Button>
          ) : (
            <div className="flex gap-2">
              <Button type="button" variant="ghost" size="sm" onClick={onCancel}>
                Cancel
              </Button>
              <Button type="submit" size="sm" disabled={isSubmitting} className="gap-1.5">
                {isSubmitting ? (
                  <SpinnerGapIcon className="size-4 animate-spin" />
                ) : (
                  <FloppyDiskIcon className="size-4" />
                )}
                Save
              </Button>
            </div>
          )}
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="name">Name</Label>
          <Field as={Input} id="name" name="name" disabled={!editing} className="h-10" />
          {errors.name && touched.name && <p className="text-xs text-destructive">{errors.name}</p>}
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="description">Description</Label>
          <Field
            as="textarea"
            id="description"
            name="description"
            disabled={!editing}
            className="flex min-h-20 w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:opacity-60 dark:bg-input/30"
          />
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="category">Category</Label>
          <Field as={Input} id="category" name="category" disabled={!editing} className="h-10" />
        </div>

        <div className="flex items-center gap-2">
          <Field
            type="checkbox"
            id="is_available"
            name="is_available"
            disabled={!editing}
            className="size-4 accent-primary"
          />
          <Label htmlFor="is_available" className="text-sm font-normal">
            Available for sale
          </Label>
        </div>
      </div>
    </Form>
  );
}

interface VariantsCardProps {
  variants: ProductVariant[];
  onAdd: () => void;
  onEdit: (variant: ProductVariant) => void;
  onDelete: (variant: ProductVariant) => Promise<void>;
}

export function VariantsCard({ variants, onAdd, onEdit, onDelete }: VariantsCardProps) {
  return (
    <div className="card-3d space-y-4 rounded-2xl p-5">
      <div className="flex items-center justify-between">
        <h2 className="text-base font-semibold">Options</h2>
        <Button variant="outline" size="sm" className="gap-1.5" onClick={onAdd}>
          <PlusIcon className="size-4" />
          Add
        </Button>
      </div>

      {variants.length === 0 ? (
        <p className="text-sm text-muted-foreground">No options yet.</p>
      ) : (
        <div className="space-y-3">
          {variants.map((variant) => (
            <div
              key={variant.id}
              className="flex items-center justify-between rounded-lg border border-border/50 p-3"
            >
              <div className="min-w-0 space-y-0.5">
                <p className="text-sm font-medium">{variant.sku}</p>
                <div className="flex items-center gap-3 text-sm text-muted-foreground">
                  <span className="font-semibold text-foreground">
                    {formatCurrency(variant.price)}
                  </span>
                  {variant.cost_price && <span>Cost: {formatCurrency(variant.cost_price)}</span>}
                  <span>
                    {variant.stock_qty === null || variant.stock_qty === undefined
                      ? "∞ stock"
                      : variant.stock_qty === 0
                        ? "Out of stock"
                        : `${variant.stock_qty} in stock`}
                  </span>
                </div>
              </div>
              <div className="flex shrink-0 items-center gap-1">
                {variant.is_default && (
                  <Badge variant="secondary" className="mr-1 text-xs">
                    Default
                  </Badge>
                )}
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 w-7 p-0"
                  onClick={() => onEdit(variant)}
                >
                  <PencilSimpleIcon className="size-3.5" />
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 w-7 p-0 text-destructive hover:text-destructive"
                  onClick={() => void onDelete(variant)}
                >
                  <TrashIcon className="size-3.5" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

interface ImagesCardProps {
  images: Array<{
    id: string;
    url: string;
    sort_order: number;
    is_primary: boolean;
  }>;
  onAdd: () => void;
  onDelete: (imageId: string) => Promise<void>;
}

export function ImagesCard({ images, onAdd, onDelete }: ImagesCardProps) {
  return (
    <div className="card-3d space-y-4 rounded-2xl p-5">
      <div className="flex items-center justify-between">
        <h2 className="text-base font-semibold">Images</h2>
        <Button variant="outline" size="sm" className="gap-1.5" onClick={onAdd}>
          <PlusIcon className="size-4" />
          Add
        </Button>
      </div>

      {images.length === 0 ? (
        <div className="flex flex-col items-center py-6 text-muted-foreground">
          <TagIcon className="mb-2 size-10 opacity-40" />
          <p className="text-sm">No images yet</p>
        </div>
      ) : (
        <div className="grid grid-cols-3 gap-2 sm:grid-cols-4">
          {[...images]
            .sort((left, right) => left.sort_order - right.sort_order)
            .map((image) => (
              <div
                key={image.id}
                className="group relative aspect-square overflow-hidden rounded-lg border"
              >
                <img src={image.url} alt="" className="size-full object-cover" />
                {image.is_primary && (
                  <Badge className="absolute top-1 left-1 px-1.5 py-0 text-[10px]">Primary</Badge>
                )}
                <button
                  type="button"
                  className="text-destructive-foreground absolute top-1 right-1 flex size-6 items-center justify-center rounded-full bg-destructive/90 opacity-0 transition-opacity group-hover:opacity-100"
                  onClick={() => void onDelete(image.id)}
                >
                  <TrashIcon className="size-3" />
                </button>
              </div>
            ))}
        </div>
      )}
    </div>
  );
}

export function DangerZoneCard({ onDelete }: { onDelete: () => void }) {
  return (
    <div className="card-3d space-y-4 rounded-2xl p-5">
      <h2 className="text-base font-semibold text-destructive">Danger zone</h2>
      <p className="text-sm text-muted-foreground">
        Deleting a product is permanent and cannot be undone.
      </p>
      <Button variant="destructive" size="sm" onClick={onDelete}>
        Delete product
      </Button>
    </div>
  );
}

export function ProductDetailSkeleton() {
  return (
    <div className="mx-auto max-w-2xl space-y-4">
      <Skeleton className="h-8 w-48" />
      <div className="card-3d space-y-4 rounded-2xl p-5">
        <Skeleton className="h-5 w-32" />
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-20 w-full" />
        <Skeleton className="h-10 w-full" />
      </div>
      <div className="card-3d space-y-4 rounded-2xl p-5">
        <Skeleton className="h-5 w-32" />
        <Skeleton className="h-16 w-full" />
      </div>
    </div>
  );
}
