"use client";

import { useMemo, useState } from "react";
import { Formik, Form } from "formik";
import * as Yup from "yup";
import {
  CheckCircleIcon,
  GlobeHemisphereWestIcon,
  SpinnerGapIcon,
  StorefrontIcon,
} from "@phosphor-icons/react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useMe } from "@/hooks/use-auth";
import { useUpdateStorefront } from "@/hooks/use-tenant";
import { ApiError } from "@/lib/api";
import { normalizeStorefrontSlug } from "@/lib/storefront";

const storefrontSchema = Yup.object({
  slug: Yup.string()
    .trim()
    .required("Storefront slug is required")
    .min(3, "Use at least 3 characters")
    .max(50, "Keep it under 50 characters")
    .matches(/^[a-z0-9]+(-[a-z0-9]+)*$/, "Use lowercase letters, numbers, and hyphens only"),
});

type FormValues = {
  slug: string;
};

export default function StorefrontPage() {
  const { data: me, error, isError, isLoading, refetch } = useMe();
  const updateStorefront = useUpdateStorefront();
  const [formError, setFormError] = useState<string | null>(null);
  const tenant = me?.onboarded ? me.tenant : undefined;

  const initialValues = useMemo<FormValues>(
    () => ({
      slug: tenant?.slug ?? "",
    }),
    [tenant],
  );

  if (isLoading) {
    return (
      <div className="card-3d flex min-h-80 flex-col items-center justify-center gap-3 rounded-2xl p-8 text-center">
        <SpinnerGapIcon className="size-5 animate-spin text-primary" />
        <p className="text-sm text-muted-foreground">Loading storefront controls</p>
      </div>
    );
  }

  if (!tenant) {
    const message =
      isError && error instanceof Error
        ? error.message
        : "Storefront details are unavailable right now.";

    return (
      <div className="card-3d flex min-h-80 flex-col items-center justify-center gap-4 rounded-2xl p-8 text-center">
        <div className="space-y-2">
          <h1 className="text-xl font-semibold">We couldn&apos;t load your storefront</h1>
          <p className="text-sm text-muted-foreground">{message}</p>
        </div>
        <Button type="button" variant="outline" onClick={() => void refetch()}>
          Retry
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h1 className="text-2xl font-bold">Storefront</h1>
        <p className="text-sm text-muted-foreground">
          Review the temporary link we created at onboarding, claim your customer-facing slug, and
          publish when you&apos;re ready.
        </p>
      </div>

      <div className="grid gap-4 lg:grid-cols-[1.15fr_0.85fr]">
        <Formik
          initialValues={initialValues}
          enableReinitialize
          validationSchema={storefrontSchema}
          onSubmit={async (values, { setErrors, setSubmitting }) => {
            setFormError(null);
            try {
              await updateStorefront.mutateAsync({
                slug: normalizeStorefrontSlug(values.slug),
                storefront_published: tenant.storefront_published,
              });
            } catch (err) {
              if (err instanceof ApiError) {
                setFormError(err.message);
                if (err.fields) {
                  setErrors({ slug: err.fields.slug });
                }
              } else {
                setFormError("Something went wrong while saving your storefront");
              }
              setSubmitting(false);
            }
          }}
        >
          {({ errors, touched, values, isSubmitting, submitCount, setFieldValue }) => {
            const tried = submitCount > 0;
            const normalizedSlug = normalizeStorefrontSlug(values.slug);
            const preview = normalizedSlug || "your-store";
            const canTogglePublish = normalizedSlug.length >= 3;

            return (
              <Form className="card-3d space-y-5 rounded-2xl p-6">
                {formError && (
                  <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
                    {formError}
                  </p>
                )}

                <div className="space-y-1.5">
                  <Label htmlFor="slug">Public storefront slug</Label>
                  <Input
                    id="slug"
                    name="slug"
                    value={values.slug}
                    onChange={(event) => {
                      void setFieldValue("slug", normalizeStorefrontSlug(event.target.value));
                    }}
                    placeholder="amina-fashion-house"
                    autoCapitalize="none"
                    autoCorrect="off"
                    spellCheck={false}
                    className="h-11"
                  />
                  {errors.slug && (touched.slug || tried) && (
                    <p className="text-xs text-destructive">{errors.slug}</p>
                  )}
                  <p className="text-xs text-muted-foreground">
                    Use lowercase letters, numbers, and hyphens only. Some words, like app, api, and
                    track, are not available.
                  </p>
                  <p className="text-xs text-muted-foreground">
                    If you change your storefront link, the new one works right away and the old one
                    stops working.
                  </p>
                </div>

                <div className="overflow-hidden rounded-2xl border border-primary/12 bg-linear-to-br from-primary/10 via-primary/4 to-transparent">
                  <div className="flex items-center gap-2 border-b border-primary/10 px-4 py-2.5 text-xs font-semibold tracking-[0.18em] text-primary/80 uppercase">
                    <GlobeHemisphereWestIcon className="size-4" weight="bold" />
                    Public preview
                  </div>
                  <div className="space-y-2 px-4 py-4">
                    <p className="text-sm text-muted-foreground">Customers will visit</p>
                    <div className="rounded-xl border border-primary/12 bg-background/80 px-3 py-3 text-base font-semibold tracking-tight text-foreground shadow-sm shadow-primary/5">
                      <span className="text-muted-foreground">storefront.com/</span>
                      <span>{preview}</span>
                    </div>
                  </div>
                </div>

                <div className="flex flex-wrap gap-3">
                  <Button type="submit" className="h-11" disabled={isSubmitting}>
                    {isSubmitting && <SpinnerGapIcon className="size-4 animate-spin" />}
                    Save slug
                  </Button>
                  <Button
                    type="button"
                    variant={tenant.storefront_published ? "outline" : "default"}
                    className="h-11"
                    disabled={isSubmitting || !canTogglePublish}
                    onClick={async () => {
                      setFormError(null);
                      try {
                        await updateStorefront.mutateAsync({
                          slug: normalizedSlug,
                          storefront_published: !tenant.storefront_published,
                        });
                      } catch (err) {
                        if (err instanceof ApiError) {
                          setFormError(err.message);
                        } else {
                          setFormError("Something went wrong while updating storefront status");
                        }
                      }
                    }}
                  >
                    {tenant.storefront_published ? "Unpublish storefront" : "Publish storefront"}
                  </Button>
                </div>
              </Form>
            );
          }}
        </Formik>

        <div className="space-y-4">
          <div className="card-3d rounded-2xl p-6">
            <div className="flex items-start gap-3">
              <div className="mt-0.5 flex size-10 items-center justify-center rounded-full bg-primary/10 text-primary">
                <StorefrontIcon className="size-5" weight="fill" />
              </div>
              <div className="space-y-2">
                <h2 className="text-lg font-semibold">Publishing status</h2>
                <div className="flex items-center gap-2 text-sm">
                  <CheckCircleIcon
                    className={
                      tenant.storefront_published
                        ? "size-4 text-primary"
                        : "size-4 text-muted-foreground"
                    }
                    weight="fill"
                  />
                  <span className="text-muted-foreground">
                    {tenant.storefront_published
                      ? "Your storefront is live and ready for public traffic."
                      : "Your storefront is still private. Customers cannot access it yet."}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <div className="card-3d rounded-2xl p-6">
            <h2 className="text-lg font-semibold">
              {tenant.storefront_published ? "Current public link" : "Reserved storefront link"}
            </h2>
            <p className="mt-2 text-sm text-muted-foreground">{`storefront.com/${tenant.slug}`}</p>
            {!tenant.storefront_published && (
              <p className="mt-2 text-sm text-muted-foreground">
                This reserved link stays private until you publish the storefront.
              </p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
