"use client";

import { useMemo, useState } from "react";
import { Form, Formik } from "formik";
import Link from "next/link";
import * as Yup from "yup";
import {
  ArrowLeftIcon,
  CheckCircleIcon,
  ShieldCheckIcon,
  SpinnerGapIcon,
  TruckIcon,
  WarningCircleIcon,
} from "@phosphor-icons/react";
import { useSession } from "@/components/auth-provider";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useMe } from "@/hooks/use-auth";
import { useUpdateTenant } from "@/hooks/use-tenant";
import { ApiError } from "@/lib/api";
import {
  formatLogisticsAddress,
  isLogisticsAddressComplete,
  missingLogisticsAddressFields,
  parseLogisticsAddress,
} from "@/lib/logistics-address";

const logisticsSchema = Yup.object({
  name: Yup.string().trim().required("Business name is required"),
  contact_email: Yup.string().trim().email("Enter a valid email").required("Email is required"),
  contact_phone: Yup.string().trim().required("Phone number is required"),
  street_address: Yup.string().trim().required("Street address is required"),
  city: Yup.string().trim().required("City is required"),
  state: Yup.string().trim().required("State is required"),
  country: Yup.string().trim().required("Country is required"),
});

type FormValues = {
  name: string;
  contact_email: string;
  contact_phone: string;
  street_address: string;
  city: string;
  state: string;
  country: string;
};

export default function LogisticsSettingsPage() {
  const { session } = useSession();
  const { data: me, error, isError, isLoading, refetch } = useMe();
  const updateTenant = useUpdateTenant();
  const [formError, setFormError] = useState<string | null>(null);
  const tenant = me?.onboarded ? me.tenant : undefined;
  const isAdmin = me?.onboarded && me.role === "admin";
  const loginEmail = session?.user.email ?? "";
  const address = parseLogisticsAddress(tenant?.address);

  const initialValues = useMemo<FormValues>(
    () => ({
      name: tenant?.name ?? "",
      contact_email: tenant?.contact_email ?? loginEmail,
      contact_phone: tenant?.contact_phone ?? "",
      street_address: address.streetAddress,
      city: address.city,
      state: address.state,
      country: address.country,
    }),
    [address.city, address.country, address.state, address.streetAddress, loginEmail, tenant],
  );

  const missingFields = [
    ...(tenant?.contact_email?.trim() ? [] : ["logistics email"]),
    ...(tenant?.contact_phone?.trim() ? [] : ["pickup phone"]),
    ...missingLogisticsAddressFields(address),
  ];
  const logisticsReady = missingFields.length === 0;

  if (isLoading) {
    return (
      <div className="card-3d flex min-h-80 flex-col items-center justify-center gap-3 rounded-2xl p-8 text-center">
        <SpinnerGapIcon className="size-5 animate-spin text-primary" />
        <p className="text-sm text-muted-foreground">Loading logistics setup</p>
      </div>
    );
  }

  if (!tenant) {
    const message =
      isError && error instanceof Error
        ? error.message
        : "Logistics settings are unavailable right now.";

    return (
      <div className="card-3d flex min-h-80 flex-col items-center justify-center gap-4 rounded-2xl p-8 text-center">
        <div className="space-y-2">
          <h1 className="text-xl font-semibold">We couldn&apos;t load logistics setup</h1>
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
      <div className="flex items-center gap-3 text-sm text-muted-foreground">
        <Link href="/app/settings" className="inline-flex items-center gap-2 hover:text-foreground">
          <ArrowLeftIcon className="size-4" />
          Back to settings
        </Link>
      </div>

      <div className="space-y-2">
        <h1 className="text-2xl font-bold">Logistics Setup</h1>
        <p className="text-sm text-muted-foreground">
          Add a clear pickup address with city, state, and country. Delivery activation happens
          automatically after this profile is complete and saved by an admin.
        </p>
      </div>

      <div className="grid gap-4 xl:grid-cols-[1.15fr_0.85fr]">
        <Formik
          initialValues={initialValues}
          enableReinitialize
          validationSchema={logisticsSchema}
          onSubmit={async (values, { setSubmitting }) => {
            setFormError(null);
            try {
              await updateTenant.mutateAsync({
                name: values.name.trim(),
                contact_email: values.contact_email.trim(),
                contact_phone: values.contact_phone.trim(),
                address: formatLogisticsAddress({
                  streetAddress: values.street_address,
                  city: values.city,
                  state: values.state,
                  country: values.country,
                }),
              });
            } catch (err) {
              if (err instanceof ApiError) {
                setFormError(err.message);
              } else {
                setFormError("Something went wrong while saving your logistics profile");
              }
              setSubmitting(false);
            }
          }}
        >
          {({ errors, touched, values, handleChange, isSubmitting }) => {
            const currentAddress = {
              streetAddress: values.street_address,
              city: values.city,
              state: values.state,
              country: values.country,
            };
            const formReady = isLogisticsAddressComplete(currentAddress);

            return (
              <Form className="card-3d space-y-5 rounded-2xl p-6">
                <div className="flex items-start justify-between gap-4">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <h2 className="text-lg font-semibold">Pickup profile</h2>
                      <Badge
                        variant={
                          tenant.active_modules.logistics
                            ? "default"
                            : logisticsReady
                              ? "secondary"
                              : "outline"
                        }
                        className="text-xs"
                      >
                        {tenant.active_modules.logistics
                          ? "Active"
                          : logisticsReady || formReady
                            ? "Ready on save"
                            : "Needs setup"}
                      </Badge>
                    </div>
                    <p className="text-sm text-muted-foreground">
                      Only admins can edit this. The pickup address must be specific enough for
                      courier validation.
                    </p>
                  </div>
                  <div className="mt-0.5 flex size-10 items-center justify-center rounded-full bg-primary/10 text-primary">
                    <ShieldCheckIcon className="size-5" weight="fill" />
                  </div>
                </div>

                {!isAdmin ? (
                  <div className="rounded-xl border border-border/60 bg-muted/40 p-4 text-sm text-muted-foreground">
                    Only store admins can update delivery and logistics setup.
                  </div>
                ) : (
                  <>
                    {formError ? (
                      <div className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
                        {formError}
                      </div>
                    ) : null}

                    <div className="grid gap-4 sm:grid-cols-2">
                      <div className="space-y-2 sm:col-span-2">
                        <Label htmlFor="name">Business name</Label>
                        <Input id="name" name="name" value={values.name} onChange={handleChange} />
                        {errors.name && touched.name ? (
                          <p className="text-xs text-destructive">{errors.name}</p>
                        ) : null}
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="contact_email">Logistics email</Label>
                        <Input
                          id="contact_email"
                          name="contact_email"
                          type="email"
                          value={values.contact_email}
                          onChange={handleChange}
                          placeholder={loginEmail || "owner@yourstore.com"}
                        />
                        <p className="text-xs text-muted-foreground">
                          Defaults to the signed-in admin email.
                        </p>
                        {errors.contact_email && touched.contact_email ? (
                          <p className="text-xs text-destructive">{errors.contact_email}</p>
                        ) : null}
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="contact_phone">Pickup phone</Label>
                        <Input
                          id="contact_phone"
                          name="contact_phone"
                          value={values.contact_phone}
                          onChange={handleChange}
                          placeholder="08012345678"
                        />
                        {errors.contact_phone && touched.contact_phone ? (
                          <p className="text-xs text-destructive">{errors.contact_phone}</p>
                        ) : null}
                      </div>

                      <div className="space-y-2 sm:col-span-2">
                        <Label htmlFor="street_address">Street address</Label>
                        <Input
                          id="street_address"
                          name="street_address"
                          value={values.street_address}
                          onChange={handleChange}
                          placeholder="16 Owerri Street, War College, Gwarinpa"
                        />
                        {errors.street_address && touched.street_address ? (
                          <p className="text-xs text-destructive">{errors.street_address}</p>
                        ) : null}
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="city">City</Label>
                        <Input
                          id="city"
                          name="city"
                          value={values.city}
                          onChange={handleChange}
                          placeholder="Abuja"
                        />
                        {errors.city && touched.city ? (
                          <p className="text-xs text-destructive">{errors.city}</p>
                        ) : null}
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="state">State</Label>
                        <Input
                          id="state"
                          name="state"
                          value={values.state}
                          onChange={handleChange}
                          placeholder="FCT"
                        />
                        {errors.state && touched.state ? (
                          <p className="text-xs text-destructive">{errors.state}</p>
                        ) : null}
                      </div>

                      <div className="space-y-2 sm:col-span-2">
                        <Label htmlFor="country">Country</Label>
                        <Input
                          id="country"
                          name="country"
                          value={values.country}
                          onChange={handleChange}
                          placeholder="Nigeria"
                        />
                        {errors.country && touched.country ? (
                          <p className="text-xs text-destructive">{errors.country}</p>
                        ) : null}
                      </div>

                      <div className="flex items-center gap-3 sm:col-span-2">
                        <Button type="submit" disabled={isSubmitting}>
                          {isSubmitting ? <SpinnerGapIcon className="size-4 animate-spin" /> : null}
                          Save logistics setup
                        </Button>
                        {tenant.active_modules.logistics ? (
                          <span className="inline-flex items-center gap-1 text-sm text-primary">
                            <CheckCircleIcon className="size-4" weight="fill" />
                            Delivery is enabled.
                          </span>
                        ) : null}
                      </div>
                    </div>
                  </>
                )}
              </Form>
            );
          }}
        </Formik>

        <div className="space-y-4">
          <div className="card-3d rounded-2xl p-6">
            <div className="flex items-start gap-3">
              <div className="mt-0.5 flex size-10 items-center justify-center rounded-full bg-primary/10 text-primary">
                <TruckIcon className="size-5" weight="fill" />
              </div>
              <div className="space-y-2">
                <h2 className="text-lg font-semibold">Activation checklist</h2>
                <p className="text-sm text-muted-foreground">
                  Quotes become available after the storefront is published and this pickup profile
                  is complete.
                </p>
              </div>
            </div>

            <div className="mt-5 space-y-3 text-sm">
              {[
                { label: "Logistics email", ready: Boolean(tenant.contact_email?.trim()) },
                { label: "Pickup phone", ready: Boolean(tenant.contact_phone?.trim()) },
                {
                  label: "Structured pickup address",
                  ready: isLogisticsAddressComplete(address),
                },
                { label: "Storefront published", ready: tenant.storefront_published },
              ].map((item) => (
                <div
                  key={item.label}
                  className="flex items-center justify-between rounded-xl border border-border/60 px-3 py-2"
                >
                  <span>{item.label}</span>
                  <Badge variant={item.ready ? "default" : "outline"} className="text-xs">
                    {item.ready ? "Done" : "Missing"}
                  </Badge>
                </div>
              ))}
            </div>

            {missingFields.length > 0 ? (
              <div className="mt-4 rounded-xl border border-border/60 bg-muted/40 p-4 text-sm text-muted-foreground">
                <p className="font-medium text-foreground">Still needed</p>
                <p className="mt-1">{missingFields.join(", ")}</p>
              </div>
            ) : null}

            {!tenant.storefront_published ? (
              <div className="mt-4 flex items-start gap-2 rounded-xl border border-border/60 bg-muted/40 p-3 text-sm text-muted-foreground">
                <WarningCircleIcon className="mt-0.5 size-4 shrink-0" />
                <span>
                  Delivery quotes stay unavailable publicly until the storefront is published.
                </span>
              </div>
            ) : null}
          </div>
        </div>
      </div>
    </div>
  );
}
