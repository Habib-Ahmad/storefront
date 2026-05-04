"use client";

import { SpinnerGapIcon } from "@phosphor-icons/react";
import { Field, Form, Formik } from "formik";
import * as Yup from "yup";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogDescription,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useCreateVariant, useUpdateVariant } from "@/hooks/use-products";
import { ApiError } from "@/lib/api";
import type { CreateVariantRequest, ProductVariant } from "@/lib/types";

interface VariantDialogProps {
  open: boolean;
  onClose: () => void;
  productId: string;
  variant?: ProductVariant;
  variantCount: number;
}

export function VariantDialog({
  open,
  onClose,
  productId,
  variant,
  variantCount,
}: VariantDialogProps) {
  const createVariant = useCreateVariant();
  const updateVariant = useUpdateVariant();
  const [error, setError] = useState<string | null>(null);
  const isEdit = !!variant;
  const isDefaultVariant = !!variant?.is_default;
  const requiresName = !isDefaultVariant || variantCount > 1 || !isEdit;
  const variantSchema = Yup.object({
    sku: requiresName
      ? Yup.string().trim().required("Option name is required")
      : Yup.string().nullable(),
    price: Yup.string()
      .required("Price is required")
      .test("positive", "Must be > 0", (value) => !!value && parseFloat(value) > 0),
    cost_price: Yup.string().nullable(),
    stock_qty: Yup.string()
      .nullable()
      .test("stock-valid", "Stock can't be negative", (value) => !value || Number(value) >= 0),
  });

  const initialValues = {
    sku: variant?.sku ?? "",
    price: variant?.price ?? "",
    cost_price: variant?.cost_price ?? "",
    stock_qty: variant?.stock_qty?.toString() ?? "",
  };

  return (
    <Dialog open={open} onOpenChange={(nextOpen: boolean) => !nextOpen && onClose()}>
      <DialogContent className="sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit Option" : "Add Option"}</DialogTitle>
          <DialogDescription>
            {requiresName
              ? "Use a customer-facing name that clearly distinguishes this option from the others."
              : "This product only has one option, so you only need to manage price and stock here."}
          </DialogDescription>
        </DialogHeader>

        <Formik
          initialValues={initialValues}
          validationSchema={variantSchema}
          enableReinitialize
          onSubmit={async (values) => {
            setError(null);
            const data: CreateVariantRequest = {
              sku: requiresName ? values.sku.trim() : (variant?.sku ?? "Default"),
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
          {({ isSubmitting, errors, touched, submitCount }) => {
            const tried = submitCount > 0;
            return (
              <Form className="space-y-4">
                {error && (
                  <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
                    {error}
                  </p>
                )}
                <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                  {requiresName && (
                    <div className="space-y-1.5">
                      <Label htmlFor="v-sku">Option name</Label>
                      <Field
                        as={Input}
                        id="v-sku"
                        name="sku"
                        placeholder={isDefaultVariant ? "e.g. Standard" : "e.g. Small, Red, 1L"}
                        className="h-10"
                      />
                      {errors.sku && (touched.sku || tried) && (
                        <p className="text-xs text-destructive">{errors.sku}</p>
                      )}
                    </div>
                  )}
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
                    {errors.price && (touched.price || tried) && (
                      <p className="text-xs text-destructive">{errors.price}</p>
                    )}
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="v-cost">Cost price (₦)</Label>
                    <Field
                      as={Input}
                      id="v-cost"
                      name="cost_price"
                      type="number"
                      placeholder="Optional internal cost"
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
                      placeholder="Leave blank for unlimited"
                      className="h-10"
                    />
                    <p className="text-xs text-muted-foreground">
                      Leave blank for unlimited stock.
                    </p>
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
