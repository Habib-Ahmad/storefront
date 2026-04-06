"use client";

import { SpinnerGapIcon } from "@phosphor-icons/react";
import { Field, Form, Formik } from "formik";
import * as Yup from "yup";
import { Button } from "@/components/ui/button";
import {
  Dialog,
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
import { useState } from "react";

const variantSchema = Yup.object({
  sku: Yup.string().required("Option name is required"),
  price: Yup.string()
    .required("Price is required")
    .test("positive", "Must be > 0", (value) => !!value && parseFloat(value) > 0),
  cost_price: Yup.string().nullable(),
  stock_qty: Yup.string().nullable(),
});

interface VariantDialogProps {
  open: boolean;
  onClose: () => void;
  productId: string;
  variant?: ProductVariant;
}

export function VariantDialog({ open, onClose, productId, variant }: VariantDialogProps) {
  const createVariant = useCreateVariant();
  const updateVariant = useUpdateVariant();
  const [error, setError] = useState<string | null>(null);
  const isEdit = !!variant;

  const initialValues = {
    sku: variant?.sku ?? "",
    price: variant?.price ?? "",
    cost_price: variant?.cost_price ?? "",
    stock_qty: variant?.stock_qty?.toString() ?? "",
  };

  return (
    <Dialog open={open} onOpenChange={(nextOpen: boolean) => !nextOpen && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit Option" : "Add Option"}</DialogTitle>
        </DialogHeader>

        <Formik
          initialValues={initialValues}
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
          {({ isSubmitting, errors, touched, submitCount }) => {
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
                    {errors.sku && (touched.sku || tried) && (
                      <p className="text-xs text-destructive">{errors.sku}</p>
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
