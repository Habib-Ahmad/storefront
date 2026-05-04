"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { Formik, Form, Field, FieldArray } from "formik";
import * as Yup from "yup";
import {
  ArrowLeftIcon,
  CaretLeftIcon,
  CaretRightIcon,
  CheckCircleIcon,
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useAddImage, useCreateProduct } from "@/hooks/use-products";
import { ApiError } from "@/lib/api";
import { PRODUCT_CATEGORY_OPTIONS } from "@/lib/product-categories";
import { uploadImageFile } from "@/lib/media-upload";

const schema = Yup.object({
  name: Yup.string().required("Name is required"),
  description: Yup.string().trim().required("Description is required"),
  category: Yup.string().nullable(),
  variants: Yup.array()
    .of(
      Yup.object({
        sku: Yup.string().nullable(),
        price: Yup.string()
          .required("Price is required")
          .test("positive", "Must be > 0", (v) => !!v && parseFloat(v) > 0),
        cost_price: Yup.string().nullable(),
        stock_qty: Yup.string()
          .nullable()
          .test("stock-valid", "Stock can't be negative", (v) => !v || Number(v) >= 0),
      }),
    )
    .test("option-name-required", "Option name is required", function (variants) {
      if (!variants || variants.length <= 1) {
        return true;
      }
      const missingIndex = variants.findIndex((variant) => !variant.sku?.trim());
      if (missingIndex >= 0) {
        return this.createError({
          path: `variants.${missingIndex}.sku`,
          message: "Option name is required",
        });
      }
      return true;
    })
    .min(1, "At least one option"),
});

type FormValues = {
  name: string;
  description: string;
  category: string;
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
  const [successState, setSuccessState] = useState<{
    productId: string;
    productName: string;
    imageUploadFailed: boolean;
  } | null>(null);
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

  function clearPendingImages() {
    for (const image of pendingImagesRef.current) {
      URL.revokeObjectURL(image.preview);
    }
    setPendingImages([]);
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

  function movePendingImage(id: string, direction: "left" | "right") {
    setPendingImages((current) => {
      const index = current.findIndex((image) => image.id === id);
      if (index < 0) {
        return current;
      }
      const nextIndex = direction === "left" ? index - 1 : index + 1;
      if (nextIndex < 0 || nextIndex >= current.length) {
        return current;
      }
      const next = [...current];
      [next[index], next[nextIndex]] = [next[nextIndex], next[index]];
      return next;
    });
  }

  const initialValues: FormValues = {
    name: "",
    description: "",
    category: "",
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
              description: values.description.trim(),
              category: values.category || null,
              is_available: true,
              variants: values.variants.map((v) => ({
                sku: values.variants.length === 1 ? v.sku.trim() || "Default" : v.sku.trim(),
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

            setSuccessState({
              productId: product.id,
              productName: values.name,
              imageUploadFailed,
            });
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
            <>
              <Form className="space-y-6">
                {formError && (
                  <p className="rounded-lg bg-destructive/10 px-3 py-2 text-center text-sm text-destructive">
                    {formError}
                  </p>
                )}

                {/* Product details */}
                <div className="card-3d space-y-4 rounded-2xl p-5">
                  <h2 className="text-base font-semibold">Basic information</h2>

                  <div className="space-y-1.5">
                    <Label htmlFor="name">Name</Label>
                    <Field
                      as={Input}
                      id="name"
                      name="name"
                      placeholder="What should this product be called?"
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
                      placeholder="Describe the product clearly so customers understand the material, size, features, or use case."
                      className="flex min-h-24 w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30"
                    />
                    {errors.description && (touched.description || tried) && (
                      <p className="text-xs text-destructive">{errors.description}</p>
                    )}
                  </div>

                  <div className="space-y-1.5">
                    <Label htmlFor="category">Category</Label>
                    <Field
                      as="select"
                      id="category"
                      name="category"
                      className="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30"
                    >
                      <option value="">Select a category</option>
                      {PRODUCT_CATEGORY_OPTIONS.map((category) => (
                        <option key={category} value={category}>
                          {category}
                        </option>
                      ))}
                    </Field>
                  </div>
                </div>

                {/* Options (variants) */}
                <div className="card-3d space-y-4 rounded-2xl p-5">
                  <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                    <div>
                      <h2 className="text-base font-semibold">Options, pricing, and stock</h2>
                      <p className="mt-1 text-sm text-muted-foreground">
                        Add every version a customer can choose from, such as size, color, or pack
                        type. If there is only one version, keep a single default option.
                      </p>
                    </div>
                  </div>

                  <FieldArray name="variants">
                    {({ push, remove }) => (
                      <div className="space-y-4">
                        {values.variants.map((_, i) => {
                          const ve = errors.variants?.[i] as Record<string, string> | undefined;
                          const vt = touched.variants?.[i] as Record<string, boolean> | undefined;
                          const isSingleOption = values.variants.length === 1;
                          const isDefaultOption = i === 0;
                          return (
                            <div key={i}>
                              {i > 0 && <Separator className="mb-4" />}
                              <div className="mb-3 flex items-center justify-between">
                                <div>
                                  <p className="text-sm font-medium text-muted-foreground">
                                    {isSingleOption ? "Default option" : `Option ${i + 1}`}
                                  </p>
                                  {isSingleOption && (
                                    <p className="text-xs text-muted-foreground">
                                      No option name needed unless you add more choices.
                                    </p>
                                  )}
                                  {!isSingleOption && isDefaultOption && (
                                    <p className="text-xs text-muted-foreground">
                                      Give this first option a clear customer-facing name too.
                                    </p>
                                  )}
                                </div>
                                {i > 0 && (
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

                              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                                {!isSingleOption && (
                                  <div className="space-y-1.5">
                                    <Label htmlFor={`variants.${i}.sku`}>Option name</Label>
                                    <Field
                                      as={Input}
                                      id={`variants.${i}.sku`}
                                      name={`variants.${i}.sku`}
                                      placeholder={
                                        isDefaultOption ? "e.g. Standard" : "e.g. Small, Red, 1L"
                                      }
                                      className="h-10"
                                    />
                                    {ve?.sku && (vt?.sku || tried) && (
                                      <p className="text-xs text-destructive">{ve.sku}</p>
                                    )}
                                  </div>
                                )}
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
                                    placeholder="Leave blank for unlimited"
                                    className="h-10"
                                  />
                                  <p className="text-xs text-muted-foreground">
                                    Leave blank for unlimited stock.
                                  </p>
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
                          Add size, color, or other option
                        </Button>
                      </div>
                    )}
                  </FieldArray>
                </div>

                <div className="card-3d space-y-4 rounded-2xl p-5">
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                    <div>
                      <h2 className="text-base font-semibold">Photos</h2>
                      <p className="mt-1 text-sm text-muted-foreground">
                        Add the photos customers should notice first. You can reorder them here and
                        choose the main photo.
                      </p>
                    </div>
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
                      <p className="text-sm">Add product photos before you publish</p>
                    </div>
                  ) : (
                    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
                      {pendingImages.map((image, index) => (
                        <div key={image.id} className="space-y-2 rounded-xl border p-2">
                          <div className="relative aspect-square overflow-hidden rounded-lg bg-muted">
                            <img src={image.preview} alt="" className="size-full object-cover" />
                            {image.isPrimary && (
                              <Badge className="absolute top-2 left-2 px-1.5 py-0 text-[10px]">
                                Primary
                              </Badge>
                            )}
                          </div>
                          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                            <Button
                              type="button"
                              variant="outline"
                              size="sm"
                              className="col-span-2"
                              onClick={() => setPrimaryImage(image.id)}
                              disabled={image.isPrimary}
                            >
                              Set as main photo
                            </Button>
                            <Button
                              type="button"
                              variant="outline"
                              size="sm"
                              disabled={index === 0}
                              onClick={() => movePendingImage(image.id, "left")}
                              aria-label="Move photo earlier"
                            >
                              <CaretLeftIcon className="size-4" />
                            </Button>
                            <Button
                              type="button"
                              variant="outline"
                              size="sm"
                              disabled={index === pendingImages.length - 1}
                              onClick={() => movePendingImage(image.id, "right")}
                              aria-label="Move photo later"
                            >
                              <CaretRightIcon className="size-4" />
                            </Button>
                            <Button
                              type="button"
                              variant="ghost"
                              size="sm"
                              className="col-span-2 text-destructive hover:text-destructive"
                              onClick={() => removePendingImage(image.id)}
                            >
                              <TrashIcon className="size-4" />
                              Remove
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>

                {/* Submit */}
                <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
                  <Link href="/app/products">
                    <Button type="button" variant="outline" className="w-full sm:w-auto">
                      Cancel
                    </Button>
                  </Link>
                  <Button type="submit" disabled={isSubmitting} className="w-full sm:w-auto">
                    {isSubmitting && <SpinnerGapIcon className="size-4 animate-spin" />}
                    Create product
                  </Button>
                </div>
              </Form>

              <Dialog
                open={!!successState}
                onOpenChange={(open: boolean) => !open && setSuccessState(null)}
              >
                <DialogContent>
                  <DialogHeader>
                    <div className="flex items-center gap-3">
                      <div className="flex size-10 items-center justify-center rounded-full bg-primary/10 text-primary">
                        <CheckCircleIcon className="size-5" weight="fill" />
                      </div>
                      <div>
                        <DialogTitle>Product created</DialogTitle>
                        <DialogDescription>
                          {successState?.productName} is ready for you to review.
                        </DialogDescription>
                      </div>
                    </div>
                  </DialogHeader>
                  {successState?.imageUploadFailed && (
                    <p className="rounded-lg bg-amber-50 px-3 py-2 text-sm text-amber-900">
                      The product was created, but at least one image failed to upload. You can fix
                      that from the product page.
                    </p>
                  )}
                  <DialogFooter>
                    <Button
                      variant="outline"
                      onClick={() => {
                        setSuccessState(null);
                        clearPendingImages();
                      }}
                    >
                      Stay here
                    </Button>
                    <Button
                      variant="outline"
                      onClick={() => {
                        clearPendingImages();
                        setFormError(null);
                        setSuccessState(null);
                      }}
                    >
                      Create another
                    </Button>
                    <Button
                      onClick={() => {
                        if (!successState) return;
                        router.push(`/app/products/${successState.productId}`);
                      }}
                    >
                      View product
                    </Button>
                  </DialogFooter>
                </DialogContent>
              </Dialog>
            </>
          );
        }}
      </Formik>
    </div>
  );
}
