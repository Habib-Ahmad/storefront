"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Formik } from "formik";
import * as Yup from "yup";
import { ArrowLeftIcon, SpinnerGapIcon } from "@phosphor-icons/react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  useProduct,
  useUpdateProduct,
  useDeleteProduct,
  useDeleteVariant,
  useAddImage,
  useDeleteImage,
} from "@/hooks/use-products";
import { ApiError } from "@/lib/api";
import type { ProductVariant } from "@/lib/types";
import { ImageDialog } from "./image-dialog";
import {
  DangerZoneCard,
  ImagesCard,
  ProductDetailSkeleton,
  ProductDetailsCard,
  VariantsCard,
} from "./product-detail-sections";
import { VariantDialog } from "./variant-dialog";

const productSchema = Yup.object({
  name: Yup.string().required("Name is required"),
  description: Yup.string().nullable(),
  category: Yup.string().nullable(),
  is_available: Yup.boolean().required(),
});

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
    return <ProductDetailSkeleton />;
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
          <ProductDetailsCard
            editing={editing}
            isSubmitting={isSubmitting}
            errors={errors as Record<string, string | undefined>}
            touched={touched as Record<string, boolean | undefined>}
            onEdit={() => setEditing(true)}
            onCancel={() => {
              resetForm();
              setEditing(false);
            }}
          />
        )}
      </Formik>

      <VariantsCard
        variants={variants}
        onAdd={() => setVariantDialog({ open: true })}
        onEdit={(variant) => setVariantDialog({ open: true, variant })}
        onDelete={async (variant) => {
          await deleteVariant.mutateAsync({ productId: product.id, variantId: variant.id });
        }}
      />

      <ImagesCard
        images={images}
        onAdd={() => setImageDialogOpen(true)}
        onDelete={async (imageId) => {
          try {
            await deleteImageMut.mutateAsync({ productId: product.id, imageId });
          } catch (err) {
            if (err instanceof ApiError) setFormError(err.message);
          }
        }}
      />

      <DangerZoneCard onDelete={() => setDeleteConfirm(true)} />

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
        productId={product.id}
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
