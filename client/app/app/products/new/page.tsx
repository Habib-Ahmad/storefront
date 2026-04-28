"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { Formik, Form, Field, FieldArray } from "formik";
import * as Yup from "yup";
import {
  ArrowLeftIcon,
  ImageIcon,
  PlusIcon,
  SpinnerGapIcon,
  TrashIcon,
} from "@phosphor-icons/react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { useAddImage, useCreateProduct } from "@/hooks/use-products";
import { ApiError } from "@/lib/api";
import { uploadImageFile } from "@/lib/media-upload";

const schema = Yup.object({
  name: Yup.string().required("Name is required"),
  description: Yup.string().nullable(),
  category: Yup.string().nullable(),
  is_available: Yup.boolean().required(),
  variants: Yup.array()
    .of(
      Yup.object({
        sku: Yup.string().required("Option name is required"),
        price: Yup.string()
          .required("Price is required")
          .test("positive", "Must be > 0", (v) => !!v && parseFloat(v) > 0),
        cost_price: Yup.string().nullable(),
        stock_qty: Yup.number().nullable().min(0, "Can't be negative"),
      }),
    )
    .min(1, "At least one option"),
});

type FormValues = {
  name: string;
  description: string;
  category: string;
  is_available: boolean;
  variants: {
    sku: string;
    price: string;
    cost_price: string;
    stock_qty: string;
  }[];
};

const emptyVariant = { sku: "", price: "", cost_price: "", stock_qty: "" };

type PendingImage = {
  id: string;
  file: File;
  preview: string;
  isPrimary: boolean;
};

export default function NewProductPage() {
  const router = useRouter();
  const createProduct = useCreateProduct();
  const addImage = useAddImage();
  const [formError, setFormError] = useState<string | null>(null);
  const [pendingImages, setPendingImages] = useState<PendingImage[]>([]);
  const pendingImagesRef = useRef<PendingImage[]>([]);

  useEffect(() => {
    pendingImagesRef.current = pendingImages;
  }, [pendingImages]);

  useEffect(() => {
    return () => {
      for (const image of pendingImagesRef.current) {
        URL.revokeObjectURL(image.preview);
      }
    };
  }, []);

  function normalizePrimary(images: PendingImage[]) {
    if (images.length === 0) {
      return images;
    }
    if (images.some((image) => image.isPrimary)) {
      return images;
    }
    return images.map((image, index) => ({ ...image, isPrimary: index === 0 }));
  }

  function handleImageSelection(event: React.ChangeEvent<HTMLInputElement>) {
    const files = Array.from(event.target.files ?? []);
    if (files.length === 0) {
      return;
    }

    setPendingImages((current) => {
      const next = [
        ...current,
        ...files.map((file, index) => ({
          id: crypto.randomUUID(),
          file,
          preview: URL.createObjectURL(file),
          isPrimary: current.length === 0 && index === 0,
        })),
      ];
      return normalizePrimary(next);
    });
    event.target.value = "";
  }

  function removePendingImage(id: string) {
    setPendingImages((current) => {
      const image = current.find((entry) => entry.id === id);
      if (image) {
        URL.revokeObjectURL(image.preview);
      }
      return normalizePrimary(current.filter((entry) => entry.id !== id));
    });
  }

  function setPrimaryImage(id: string) {
    setPendingImages((current) =>
      current.map((image) => ({
        ...image,
        isPrimary: image.id === id,
      })),
    );
  }

  const initialValues: FormValues = {
    name: "",
    description: "",
    category: "",
    is_available: true,
    variants: [{ ...emptyVariant }],
  };

  return (
    <div className="mx-auto max-w-2xl space-y-4">
      <div>
        <Link href="/app/products">
          <Button variant="ghost" size="sm" className="-ml-2 gap-1">
            <ArrowLeftIcon className="size-4" />
            Back
          </Button>
        </Link>
        <h1 className="mt-1 text-2xl font-bold">New Product</h1>
      </div>

      <Formik
        initialValues={initialValues}
        validationSchema={schema}
        onSubmit={async (values) => {
          setFormError(null);
          try {
            const product = await createProduct.mutateAsync({
              name: values.name,
              description: values.description || null,
              category: values.category || null,
              is_available: values.is_available,
              variants: values.variants.map((v) => ({
                sku: v.sku,
                price: v.price,
                cost_price: v.cost_price || null,
                stock_qty: v.stock_qty ? parseInt(v.stock_qty, 10) : null,
              })),
            });

            let imageUploadFailed = false;
            for (const [index, image] of pendingImages.entries()) {
              try {
                const uploadedURL = await uploadImageFile(image.file, product.id);
                await addImage.mutateAsync({
                  productId: product.id,
                  data: {
                    url: uploadedURL,
                    sort_order: index,
                    is_primary: image.isPrimary,
                  },
                });
              } catch {
                imageUploadFailed = true;
                break;
              }
            }

            router.push(
              imageUploadFailed
                ? `/app/products/${product.id}?image_upload=failed`
                : `/app/products/${product.id}`,
            );
          } catch (err) {
            if (err instanceof ApiError) {
              setFormError(err.message);
            }
          }
        }}
      >
        {({ isSubmitting, errors, touched, values, submitCount }) => {
          const tried = submitCount > 0;
          return (
            <Form className="space-y-6">
              {formError && (
                <p className="rounded-lg bg-destructive/10 px-3 py-2 text-center text-sm text-destructive">
                  {formError}
                </p>
              )}

              {/* Product details */}
              <div className="card-3d space-y-4 rounded-2xl p-5">
                <h2 className="text-base font-semibold">Details</h2>

                <div className="space-y-1.5">
                  <Label htmlFor="name">Name</Label>
                  <Field
                    as={Input}
                    id="name"
                    name="name"
                    placeholder="Product name"
                    className="h-10"
                  />
                  {errors.name && (touched.name || tried) && (
                    <p className="text-xs text-destructive">{errors.name}</p>
                  )}
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="description">Description</Label>
                  <Field
                    as="textarea"
                    id="description"
                    name="description"
                    placeholder="Optional product description"
                    className="flex min-h-20 w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30"
                  />
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="category">Category</Label>
                  <Field
                    as={Input}
                    id="category"
                    name="category"
                    placeholder="e.g. Clothing, Electronics"
                    className="h-10"
                  />
                </div>

                <div className="flex items-center gap-2">
                  <Field
                    type="checkbox"
                    id="is_available"
                    name="is_available"
                    className="size-4 accent-primary"
                  />
                  <Label htmlFor="is_available" className="text-sm font-normal">
                    Available for sale
                  </Label>
                </div>
              </div>

              {/* Options (variants) */}
              <div className="card-3d space-y-4 rounded-2xl p-5">
                <div className="flex items-center justify-between">
                  <h2 className="text-base font-semibold">Pricing & Inventory</h2>
                </div>

                <FieldArray name="variants">
                  {({ push, remove }) => (
                    <div className="space-y-4">
                      {values.variants.map((_, i) => {
                        const ve = errors.variants?.[i] as Record<string, string> | undefined;
                        const vt = touched.variants?.[i] as Record<string, boolean> | undefined;
                        return (
                          <div key={i}>
                            {i > 0 && <Separator className="mb-4" />}
                            <div className="mb-3 flex items-center justify-between">
                              <p className="text-sm font-medium text-muted-foreground">
                                {values.variants.length > 1 ? `Option ${i + 1}` : "Default option"}
                              </p>
                              {values.variants.length > 1 && (
                                <Button
                                  type="button"
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => remove(i)}
                                  className="h-7 px-2 text-destructive hover:text-destructive"
                                >
                                  <TrashIcon className="size-4" />
                                </Button>
                              )}
                            </div>

                            <div className="grid grid-cols-2 gap-3">
                              <div className="space-y-1.5">
                                <Label htmlFor={`variants.${i}.sku`}>Option name</Label>
                                <Field
                                  as={Input}
                                  id={`variants.${i}.sku`}
                                  name={`variants.${i}.sku`}
                                  placeholder="e.g. Default, Small, Red"
                                  className="h-10"
                                />
                                {ve?.sku && (vt?.sku || tried) && (
                                  <p className="text-xs text-destructive">{ve.sku}</p>
                                )}
                              </div>
                              <div className="space-y-1.5">
                                <Label htmlFor={`variants.${i}.price`}>Price (₦)</Label>
                                <Field
                                  as={Input}
                                  id={`variants.${i}.price`}
                                  name={`variants.${i}.price`}
                                  type="number"
                                  placeholder="0.00"
                                  className="h-10"
                                />
                                {ve?.price && (vt?.price || tried) && (
                                  <p className="text-xs text-destructive">{ve.price}</p>
                                )}
                              </div>
                              <div className="space-y-1.5">
                                <Label htmlFor={`variants.${i}.cost_price`}>Cost price (₦)</Label>
                                <Field
                                  as={Input}
                                  id={`variants.${i}.cost_price`}
                                  name={`variants.${i}.cost_price`}
                                  type="number"
                                  placeholder="Optional"
                                  className="h-10"
                                />
                              </div>
                              <div className="space-y-1.5">
                                <Label htmlFor={`variants.${i}.stock_qty`}>Stock</Label>
                                <Field
                                  as={Input}
                                  id={`variants.${i}.stock_qty`}
                                  name={`variants.${i}.stock_qty`}
                                  type="number"
                                  placeholder="∞"
                                  className="h-10"
                                />
                              </div>
                            </div>
                          </div>
                        );
                      })}

                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => push({ ...emptyVariant })}
                        className="gap-1.5"
                      >
                        <PlusIcon className="size-4" />
                        Add option
                      </Button>
                    </div>
                  )}
                </FieldArray>
              </div>

              <div className="card-3d space-y-4 rounded-2xl p-5">
                <div className="flex items-center justify-between">
                  <h2 className="text-base font-semibold">Images</h2>
                  <label className="inline-flex cursor-pointer items-center gap-1.5 rounded-[min(var(--radius-md),12px)] border border-border bg-background px-2.5 py-1 text-[0.8rem] font-medium transition-colors hover:bg-muted hover:text-foreground">
                    <Input
                      type="file"
                      accept="image/*"
                      multiple
                      className="hidden"
                      onChange={handleImageSelection}
                    />
                    <PlusIcon className="size-4" />
                    <span>Add images</span>
                  </label>
                </div>

                {pendingImages.length === 0 ? (
                  <div className="flex flex-col items-center py-6 text-muted-foreground">
                    <ImageIcon className="mb-2 size-10 opacity-40" />
                    <p className="text-sm">Upload product images during creation</p>
                  </div>
                ) : (
                  <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
                    {pendingImages.map((image) => (
                      <div key={image.id} className="space-y-2 rounded-xl border p-2">
                        <div className="relative aspect-square overflow-hidden rounded-lg bg-muted">
                          <img src={image.preview} alt="" className="size-full object-cover" />
                          {image.isPrimary && (
                            <Badge className="absolute top-2 left-2 px-1.5 py-0 text-[10px]">
                              Primary
                            </Badge>
                          )}
                        </div>
                        <div className="flex gap-2">
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            className="flex-1"
                            onClick={() => setPrimaryImage(image.id)}
                          >
                            Set primary
                          </Button>
                          <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            className="text-destructive hover:text-destructive"
                            onClick={() => removePendingImage(image.id)}
                          >
                            <TrashIcon className="size-4" />
                          </Button>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Submit */}
              <div className="flex justify-end gap-3">
                <Link href="/app/products">
                  <Button type="button" variant="outline">
                    Cancel
                  </Button>
                </Link>
                <Button type="submit" disabled={isSubmitting}>
                  {isSubmitting && <SpinnerGapIcon className="size-4 animate-spin" />}
                  Create product
                </Button>
              </div>
            </Form>
          );
        }}
      </Formik>
    </div>
  );
}
