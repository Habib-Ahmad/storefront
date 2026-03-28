"use client";

import { useState, useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import { Formik, Form, Field, FieldArray } from "formik";
import * as Yup from "yup";
import {
  ArrowLeftIcon,
  PlusIcon,
  TrashIcon,
  SpinnerGapIcon,
  PencilSimpleIcon,
  FloppyDiskIcon,
  TagIcon,
  UploadSimpleIcon,
} from "@phosphor-icons/react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  useProduct,
  useUpdateProduct,
  useDeleteProduct,
  useCreateVariant,
  useUpdateVariant,
  useDeleteVariant,
  useAddImage,
  useDeleteImage,
} from "@/hooks/use-products";
import { ApiError, api } from "@/lib/api";
import { cn } from "@/lib/utils";
import type { ProductVariant, CreateVariantRequest } from "@/lib/contracts";

const productSchema = Yup.object({
  name: Yup.string().required("Name is required"),
  description: Yup.string().nullable(),
  category: Yup.string().nullable(),
  is_available: Yup.boolean().required(),
});

const variantSchema = Yup.object({
  sku: Yup.string().required("Option name is required"),
  price: Yup.string()
    .required("Price is required")
    .test("positive", "Must be > 0", (v) => !!v && parseFloat(v) > 0),
  cost_price: Yup.string().nullable(),
  stock_qty: Yup.string().nullable(),
});

const imageSchema = Yup.object({
  url: Yup.string().url("Must be a valid URL").required("URL is required"),
  is_primary: Yup.boolean().required(),
});

function formatCurrency(val: string) {
  return new Intl.NumberFormat("en-NG", {
    style: "currency",
    currency: "NGN",
    minimumFractionDigits: 0,
  }).format(parseFloat(val));
}

// ── Variant Dialog ─────────────────────────────────────

function VariantDialog({
  open,
  onClose,
  productId,
  variant,
}: {
  open: boolean;
  onClose: () => void;
  productId: string;
  variant?: ProductVariant;
}) {
  const createVariant = useCreateVariant();
  const updateVariant = useUpdateVariant();
  const [error, setError] = useState<string | null>(null);

  const isEdit = !!variant;

  const initial = {
    sku: variant?.sku ?? "",
    price: variant?.price ?? "",
    cost_price: variant?.cost_price ?? "",
    stock_qty: variant?.stock_qty?.toString() ?? "",
  };

  return (
    <Dialog open={open} onOpenChange={(o: boolean) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit Option" : "Add Option"}</DialogTitle>
        </DialogHeader>

        <Formik
          initialValues={initial}
          validationSchema={variantSchema}
          enableReinitialize
          onSubmit={async (values) => {
            setError(null);
            const data: CreateVariantRequest = {
              sku: values.sku,
              price: values.price,
              cost_price: values.cost_price || null,
              stock_qty: values.stock_qty ? parseInt(values.stock_qty, 10) : null,
            };
            try {
              if (isEdit && variant) {
                await updateVariant.mutateAsync({ productId, variantId: variant.id, data });
              } else {
                await createVariant.mutateAsync({ productId, data });
              }
              onClose();
            } catch (err) {
              if (err instanceof ApiError) setError(err.message);
            }
          }}
        >
          {({ isSubmitting, errors: e, touched: t, submitCount }) => {
            const tried = submitCount > 0;
            return (
              <Form className="space-y-4">
                {error && (
                  <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
                    {error}
                  </p>
                )}
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-1.5">
                    <Label htmlFor="v-sku">Option name</Label>
                    <Field
                      as={Input}
                      id="v-sku"
                      name="sku"
                      placeholder="e.g. Default, Small, Red"
                      className="h-10"
                    />
                    {e.sku && (t.sku || tried) && (
                      <p className="text-xs text-destructive">{e.sku}</p>
                    )}
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="v-price">Price (₦)</Label>
                    <Field
                      as={Input}
                      id="v-price"
                      name="price"
                      type="number"
                      placeholder="0.00"
                      className="h-10"
                    />
                    {e.price && (t.price || tried) && (
                      <p className="text-xs text-destructive">{e.price}</p>
                    )}
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="v-cost">Cost price (₦)</Label>
                    <Field
                      as={Input}
                      id="v-cost"
                      name="cost_price"
                      type="number"
                      placeholder="Optional"
                      className="h-10"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="v-stock">Stock</Label>
                    <Field
                      as={Input}
                      id="v-stock"
                      name="stock_qty"
                      type="number"
                      placeholder="∞"
                      className="h-10"
                    />
                  </div>
                </div>

                <DialogFooter>
                  <Button type="button" variant="outline" onClick={onClose}>
                    Cancel
                  </Button>
                  <Button type="submit" disabled={isSubmitting}>
                    {isSubmitting && <SpinnerGapIcon className="size-4 animate-spin" />}
                    {isEdit ? "Save" : "Add"}
                  </Button>
                </DialogFooter>
              </Form>
            );
          }}
        </Formik>
      </DialogContent>
    </Dialog>
  );
}

// ── Main Page ──────────────────────────────────────────

export default function ProductDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const { data, isLoading } = useProduct(id);
  const updateProduct = useUpdateProduct();
  const deleteProduct = useDeleteProduct();
  const deleteVariant = useDeleteVariant();
  const addImage = useAddImage();
  const deleteImageMut = useDeleteImage();

  const [editing, setEditing] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [variantDialog, setVariantDialog] = useState<{
    open: boolean;
    variant?: ProductVariant;
  }>({ open: false });
  const [deleteConfirm, setDeleteConfirm] = useState(false);
  const [imageDialogOpen, setImageDialogOpen] = useState(false);

  if (isLoading) {
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

  if (!data) {
    return (
      <div className="mx-auto max-w-2xl space-y-4">
        <Link href="/app/products">
          <Button variant="ghost" size="sm" className="gap-1">
            <ArrowLeftIcon className="size-4" /> Back
          </Button>
        </Link>
        <p className="text-muted-foreground">Product not found.</p>
      </div>
    );
  }

  const { product, variants: rawVariants, images: rawImages } = data;
  const variants = rawVariants ?? [];
  const images = rawImages ?? [];

  return (
    <div className="mx-auto max-w-2xl space-y-4">
      {/* Header */}
      <div>
        <div className="flex items-center justify-between">
          <Link href="/app/products">
            <Button variant="ghost" size="sm" className="-ml-2 gap-1">
              <ArrowLeftIcon className="size-4" />
              Back
            </Button>
          </Link>
          <Badge variant={product.is_available ? "default" : "secondary"}>
            {product.is_available ? "Active" : "Draft"}
          </Badge>
        </div>
        <h1 className="mt-1 truncate text-2xl font-bold">{product.name}</h1>
      </div>

      {formError && (
        <p className="rounded-lg bg-destructive/10 px-3 py-2 text-center text-sm text-destructive">
          {formError}
        </p>
      )}

      {/* Product details card */}
      <Formik
        initialValues={{
          name: product.name,
          description: product.description ?? "",
          category: product.category ?? "",
          is_available: product.is_available,
        }}
        validationSchema={productSchema}
        enableReinitialize
        onSubmit={async (values) => {
          setFormError(null);
          try {
            await updateProduct.mutateAsync({
              id: product.id,
              data: {
                name: values.name,
                description: values.description || null,
                category: values.category || null,
                is_available: values.is_available,
              },
            });
            setEditing(false);
          } catch (err) {
            if (err instanceof ApiError) setFormError(err.message);
          }
        }}
      >
        {({ isSubmitting, errors, touched, resetForm }) => (
          <Form>
            <div className="card-3d space-y-4 rounded-2xl p-5">
              <div className="flex items-center justify-between">
                <h2 className="text-base font-semibold">Details</h2>
                {!editing ? (
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => setEditing(true)}
                    className="gap-1.5"
                  >
                    <PencilSimpleIcon className="size-4" />
                    Edit
                  </Button>
                ) : (
                  <div className="flex gap-2">
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => {
                        resetForm();
                        setEditing(false);
                      }}
                    >
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
                {errors.name && touched.name && (
                  <p className="text-xs text-destructive">{errors.name}</p>
                )}
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
                <Field
                  as={Input}
                  id="category"
                  name="category"
                  disabled={!editing}
                  className="h-10"
                />
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
        )}
      </Formik>

      {/* Variants card */}
      <div className="card-3d space-y-4 rounded-2xl p-5">
        <div className="flex items-center justify-between">
          <h2 className="text-base font-semibold">Options</h2>
          <Button
            variant="outline"
            size="sm"
            className="gap-1.5"
            onClick={() => setVariantDialog({ open: true })}
          >
            <PlusIcon className="size-4" />
            Add
          </Button>
        </div>

        {variants.length === 0 ? (
          <p className="text-sm text-muted-foreground">No options yet.</p>
        ) : (
          <div className="space-y-3">
            {variants.map((v) => (
              <div
                key={v.id}
                className="flex items-center justify-between rounded-lg border border-border/50 p-3"
              >
                <div className="min-w-0 space-y-0.5">
                  <p className="text-sm font-medium">{v.sku}</p>
                  <div className="flex items-center gap-3 text-sm text-muted-foreground">
                    <span className="font-semibold text-foreground">{formatCurrency(v.price)}</span>
                    {v.cost_price && <span>Cost: {formatCurrency(v.cost_price)}</span>}
                    <span>
                      {v.stock_qty === null || v.stock_qty === undefined
                        ? "∞ stock"
                        : v.stock_qty === 0
                          ? "Out of stock"
                          : `${v.stock_qty} in stock`}
                    </span>
                  </div>
                </div>
                <div className="flex shrink-0 items-center gap-1">
                  {v.is_default && (
                    <Badge variant="secondary" className="mr-1 text-xs">
                      Default
                    </Badge>
                  )}
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-7 w-7 p-0"
                    onClick={() => setVariantDialog({ open: true, variant: v })}
                  >
                    <PencilSimpleIcon className="size-3.5" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-7 w-7 p-0 text-destructive hover:text-destructive"
                    onClick={async () => {
                      await deleteVariant.mutateAsync({ productId: product.id, variantId: v.id });
                    }}
                  >
                    <TrashIcon className="size-3.5" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Images card */}
      <div className="card-3d space-y-4 rounded-2xl p-5">
        <div className="flex items-center justify-between">
          <h2 className="text-base font-semibold">Images</h2>
          <Button
            variant="outline"
            size="sm"
            className="gap-1.5"
            onClick={() => setImageDialogOpen(true)}
          >
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
              .sort((a, b) => a.sort_order - b.sort_order)
              .map((img) => (
                <div
                  key={img.id}
                  className="group relative aspect-square overflow-hidden rounded-lg border"
                >
                  <img src={img.url} alt="" className="size-full object-cover" />
                  {img.is_primary && (
                    <Badge className="absolute top-1 left-1 px-1.5 py-0 text-[10px]">Primary</Badge>
                  )}
                  <button
                    type="button"
                    className="text-destructive-foreground absolute top-1 right-1 flex size-6 items-center justify-center rounded-full bg-destructive/90 opacity-0 transition-opacity group-hover:opacity-100"
                    onClick={async () => {
                      try {
                        await deleteImageMut.mutateAsync({
                          productId: product.id,
                          imageId: img.id,
                        });
                      } catch (err) {
                        if (err instanceof ApiError) setFormError(err.message);
                      }
                    }}
                  >
                    <TrashIcon className="size-3" />
                  </button>
                </div>
              ))}
          </div>
        )}
      </div>

      {/* Danger zone */}
      <div className="card-3d space-y-4 rounded-2xl p-5">
        <h2 className="text-base font-semibold text-destructive">Danger zone</h2>
        <p className="text-sm text-muted-foreground">
          Deleting a product is permanent and cannot be undone.
        </p>
        <Button variant="destructive" size="sm" onClick={() => setDeleteConfirm(true)}>
          Delete product
        </Button>
      </div>

      {/* Variant dialog */}
      <VariantDialog
        open={variantDialog.open}
        onClose={() => setVariantDialog({ open: false })}
        productId={product.id}
        variant={variantDialog.variant}
      />

      {/* Delete confirm dialog */}
      <Dialog open={deleteConfirm} onOpenChange={(o: boolean) => setDeleteConfirm(o)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete {product.name}?</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            This will permanently remove this product and all its variants.
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteConfirm(false)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={async () => {
                await deleteProduct.mutateAsync(product.id);
                router.replace("/app/products");
              }}
            >
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Image add dialog */}
      <ImageDialog
        open={imageDialogOpen}
        onClose={() => setImageDialogOpen(false)}
        nextSortOrder={images.length}
        onSubmit={async (imgData) => {
          setFormError(null);
          try {
            await addImage.mutateAsync({ productId: product.id, data: imgData });
            setImageDialogOpen(false);
          } catch (err) {
            if (err instanceof ApiError) throw err;
          }
        }}
      />
    </div>
  );
}

// ── Image Dialog ───────────────────────────────────────

function ImageDialog({
  open,
  onClose,
  nextSortOrder,
  onSubmit,
}: {
  open: boolean;
  onClose: () => void;
  nextSortOrder: number;
  onSubmit: (data: { url: string; sort_order: number; is_primary: boolean }) => Promise<void>;
}) {
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const [isPrimary, setIsPrimary] = useState(nextSortOrder === 0);
  const [uploading, setUploading] = useState(false);
  const [dialogError, setDialogError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) {
      setFile(null);
      if (preview) URL.revokeObjectURL(preview);
      setPreview(null);
      setDialogError(null);
    }
    setIsPrimary(nextSortOrder === 0);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, nextSortOrder]);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0];
    if (!f) return;
    if (preview) URL.revokeObjectURL(preview);
    setFile(f);
    setPreview(URL.createObjectURL(f));
  };

  const handleSubmit = async () => {
    if (!file) return;
    setUploading(true);
    setDialogError(null);
    try {
      const { upload_url } = await api.getUploadUrl();
      const form = new FormData();
      form.append("file", file);
      const cfRes = await fetch(upload_url, { method: "POST", body: form });
      const cfData = await cfRes.json();
      if (!cfData.success) throw new Error("Cloudflare rejected the upload");
      const url: string = cfData.result.variants[0];
      await onSubmit({ url, sort_order: nextSortOrder, is_primary: isPrimary });
    } catch (err) {
      if (err instanceof ApiError) setDialogError(err.message);
      else if (err instanceof Error) setDialogError(err.message);
    } finally {
      setUploading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={(o: boolean) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add image</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          {dialogError && (
            <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {dialogError}
            </p>
          )}
          <label className="block cursor-pointer">
            <div
              className={cn(
                "flex aspect-video flex-col items-center justify-center gap-2 rounded-xl border-2 border-dashed text-muted-foreground transition-colors",
                preview
                  ? "overflow-hidden border-transparent p-0"
                  : "border-border p-6 hover:border-primary/50",
              )}
            >
              {preview ? (
                <img src={preview} alt="" className="size-full object-cover" />
              ) : (
                <>
                  <UploadSimpleIcon className="size-8 opacity-40" />
                  <p className="text-sm">Tap to choose a photo</p>
                </>
              )}
            </div>
            <input type="file" accept="image/*" className="sr-only" onChange={handleFileChange} />
          </label>
          {preview && (
            <button
              type="button"
              className="text-xs text-muted-foreground underline"
              onClick={() => {
                URL.revokeObjectURL(preview!);
                setFile(null);
                setPreview(null);
              }}
            >
              Choose a different image
            </button>
          )}
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="img-primary"
              className="size-4 accent-primary"
              checked={isPrimary}
              onChange={(e) => setIsPrimary(e.target.checked)}
            />
            <Label htmlFor="img-primary" className="text-sm font-normal">
              Set as primary image
            </Label>
          </div>
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={!file || uploading}>
            {uploading && <SpinnerGapIcon className="size-4 animate-spin" />}
            {uploading ? "Uploading…" : "Add image"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
