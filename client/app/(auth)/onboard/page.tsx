"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { Formik, Form } from "formik";
import * as Yup from "yup";
import { motion } from "framer-motion";
import {
  CheckCircleIcon,
  SignOutIcon,
  SpinnerGapIcon,
  StorefrontIcon,
} from "@phosphor-icons/react";
import { useSession } from "@/components/auth-provider";
import { ShoppingBagSvg } from "@/components/illustrations";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useMe, useSignOut } from "@/hooks/use-auth";
import { useOnboardTenant } from "@/hooks/use-tenant";
import { ApiError } from "@/lib/api";
import { getTemporaryStorefrontSlugPreview } from "@/lib/storefront";

const onboardingSchema = Yup.object({
  name: Yup.string()
    .trim()
    .required("Business name is required")
    .max(80, "Keep it under 80 characters"),
});

type FormValues = {
  name: string;
};

type SuccessStage = "preparing" | "done";

function SuccessSparkles() {
  const sparkles = [
    { left: "12%", top: "26%", delay: 0 },
    { left: "82%", top: "18%", delay: 0.2 },
    { left: "72%", top: "72%", delay: 0.4 },
    { left: "20%", top: "78%", delay: 0.6 },
    { left: "50%", top: "10%", delay: 0.8 },
  ];

  return (
    <div className="pointer-events-none absolute inset-0" aria-hidden="true">
      {sparkles.map((sparkle) => (
        <motion.span
          key={`${sparkle.left}-${sparkle.top}`}
          className="absolute block size-2 rounded-full bg-primary/35"
          style={{ left: sparkle.left, top: sparkle.top }}
          initial={{ opacity: 0, scale: 0.3, y: 4 }}
          animate={{ opacity: [0, 1, 0], scale: [0.3, 1.1, 0.6], y: [4, -10, -18] }}
          transition={{ duration: 1.6, delay: sparkle.delay, ease: "easeOut" }}
        />
      ))}
    </div>
  );
}

export default function OnboardPage() {
  const router = useRouter();
  const { session, loading: sessionLoading } = useSession();
  const { data: me, isLoading: meLoading } = useMe();
  const signOut = useSignOut();
  const onboardTenant = useOnboardTenant();
  const [formError, setFormError] = useState<string | null>(null);
  const [successName, setSuccessName] = useState<string | null>(null);
  const [successSlug, setSuccessSlug] = useState<string | null>(null);
  const [successStage, setSuccessStage] = useState<SuccessStage>("preparing");
  const [showIncompleteSetupBanner, setShowIncompleteSetupBanner] = useState(false);

  useEffect(() => {
    const bannerReason = window.sessionStorage.getItem("storefront:onboarding-banner");
    if (bannerReason === "app-guard") {
      setShowIncompleteSetupBanner(true);
      window.sessionStorage.removeItem("storefront:onboarding-banner");
    }
  }, []);

  useEffect(() => {
    if (!successName && me?.onboarded) {
      router.replace("/app");
    }
  }, [me?.onboarded, router, successName]);

  useEffect(() => {
    if (!successName) return;
    setSuccessStage("preparing");
    const stageTimer = window.setTimeout(() => {
      setSuccessStage("done");
    }, 2800);
    const redirectTimer = window.setTimeout(() => {
      router.replace("/app/storefront");
    }, 4600);

    return () => {
      window.clearTimeout(stageTimer);
      window.clearTimeout(redirectTimer);
    };
  }, [router, successName]);

  const ownerEmail = session?.user.email ?? "";

  const initialValues = useMemo<FormValues>(
    () => ({
      name: "",
    }),
    [],
  );

  if (sessionLoading || meLoading) {
    return (
      <div className="space-y-6 text-center">
        <div className="flex justify-center">
          <ShoppingBagSvg className="size-28" />
        </div>
        <div className="card-3d space-y-3 rounded-2xl p-6">
          <SpinnerGapIcon className="mx-auto size-5 animate-spin text-primary" />
          <div className="space-y-1">
            <h1 className="text-xl font-semibold">Setting things up</h1>
            <p className="text-sm text-muted-foreground">
              Checking your account before we create your store
            </p>
          </div>
        </div>
      </div>
    );
  }

  if (!session) {
    return null;
  }

  if (successName) {
    return (
      <motion.div
        className="relative space-y-6"
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.35, ease: "easeOut" }}
      >
        <div className="flex justify-center">
          {successStage === "preparing" ? (
            <motion.div
              className="relative flex size-32 items-center justify-center rounded-full border border-primary/16 bg-primary/6 shadow-lg shadow-primary/8"
              initial={{ scale: 0.94, opacity: 0.7 }}
              animate={{ scale: 1, opacity: 1 }}
              transition={{ duration: 0.5, ease: "easeOut" }}
            >
              <motion.div
                className="absolute inset-0 rounded-full border border-primary/14"
                animate={{ scale: [0.88, 1.08], opacity: [0.45, 0] }}
                transition={{ duration: 1.5, repeat: Infinity, ease: "easeOut" }}
              />
              <motion.div
                className="absolute inset-3 rounded-full border border-primary/12"
                animate={{ scale: [0.96, 1.04], opacity: [0.5, 0.18, 0.5] }}
                transition={{ duration: 1.8, repeat: Infinity, ease: "easeInOut" }}
              />
              <SpinnerGapIcon className="size-12 animate-spin text-primary" />
            </motion.div>
          ) : (
            <motion.div
              className="relative flex size-32 items-center justify-center rounded-full border border-primary/20 bg-primary/8 shadow-lg shadow-primary/10"
              initial={{ scale: 0.86, opacity: 0.7 }}
              animate={{ scale: [0.86, 1.02, 1], opacity: 1 }}
              transition={{ duration: 0.8, ease: "easeOut" }}
            >
              <motion.div
                className="absolute inset-0 rounded-full border border-primary/20"
                initial={{ scale: 0.8, opacity: 0 }}
                animate={{ scale: [0.8, 1.08, 1.16], opacity: [0, 0.4, 0] }}
                transition={{ duration: 1.4, repeat: 1, ease: "easeOut" }}
              />
              <SuccessSparkles />
              <motion.div
                initial={{ scale: 0.7, rotate: -8 }}
                animate={{ scale: [0.7, 1.08, 1], rotate: [-8, 4, 0] }}
                transition={{ duration: 0.9, ease: "easeOut", delay: 0.1 }}
              >
                <CheckCircleIcon className="size-18 text-primary" weight="fill" />
              </motion.div>
            </motion.div>
          )}
        </div>

        <div className="card-3d relative space-y-4 rounded-2xl p-6 text-center">
          {successStage === "preparing" ? (
            <>
              <div className="space-y-1.5">
                <h1 className="text-2xl font-bold tracking-tight">Setting up your store</h1>
                <p className="text-sm text-muted-foreground">
                  Creating {successName}, reserving your temporary storefront link, and getting
                  everything ready.
                </p>
              </div>
              <div className="space-y-2 rounded-xl border border-primary/15 bg-primary/6 px-3 py-3 text-sm text-primary">
                <p>This usually takes a moment</p>
                {successSlug && (
                  <p className="font-medium tracking-tight text-foreground">
                    storefront.com/{successSlug}
                  </p>
                )}
              </div>
            </>
          ) : (
            <>
              <div className="space-y-1.5">
                <h1 className="text-2xl font-bold tracking-tight">Success. Redirecting you now.</h1>
                <p className="text-sm text-muted-foreground">
                  {successName} is ready. Taking you to Storefront so you can review the draft link
                  and publish when you&apos;re ready.
                </p>
              </div>
              <div className="space-y-2 rounded-xl border border-primary/15 bg-primary/6 px-3 py-3 text-sm text-primary">
                <p>Setup complete</p>
                {successSlug && (
                  <p className="font-medium tracking-tight text-foreground">
                    storefront.com/{successSlug}
                  </p>
                )}
              </div>
            </>
          )}
          <div className="overflow-hidden rounded-full bg-primary/10">
            <motion.div
              className="h-1.5 rounded-full bg-primary"
              initial={{ width: "0%" }}
              animate={{ width: "100%" }}
              transition={{ duration: 4.3, ease: "easeInOut" }}
            />
          </div>
        </div>
      </motion.div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-end">
        <Button type="button" variant="ghost" size="sm" className="gap-1.5" onClick={signOut}>
          <SignOutIcon className="size-4" />
          Sign out
        </Button>
      </div>

      <div className="flex justify-center">
        <ShoppingBagSvg className="size-28" />
      </div>

      <div className="space-y-2 text-center">
        <h1 className="text-2xl font-bold tracking-tight">Create your workspace</h1>
        <p className="text-sm text-muted-foreground">
          We&apos;ll create a private temporary storefront link now so your store identity is
          visible from day one. You can rename and publish it later.
        </p>
      </div>

      <Formik
        initialValues={initialValues}
        validationSchema={onboardingSchema}
        onSubmit={async (values, { setErrors, setSubmitting }) => {
          setFormError(null);

          try {
            const tenant = await onboardTenant.mutateAsync({
              name: values.name.trim(),
              admin_email: ownerEmail,
            });
            setSuccessName(tenant.name);
            setSuccessSlug(tenant.slug);
          } catch (err) {
            if (err instanceof ApiError) {
              setFormError(err.message);
              if (err.fields) {
                setErrors({
                  name: err.fields.name,
                });
              }
            } else {
              setFormError("Something went wrong while creating your store");
            }
            setSubmitting(false);
          }
        }}
      >
        {({ errors, touched, values, isSubmitting, submitCount, handleChange }) => {
          const tried = submitCount > 0;
          const temporarySlugPreview = getTemporaryStorefrontSlugPreview(values.name);

          return (
            <Form className="md:card-3d space-y-5 md:rounded-2xl md:p-7">
              {showIncompleteSetupBanner && (
                <p className="rounded-lg border border-primary/15 bg-primary/8 px-3 py-2 text-center text-sm text-primary">
                  Finish setting up your workspace to continue.
                </p>
              )}

              {formError && (
                <p className="rounded-lg bg-destructive/10 px-3 py-2 text-center text-sm text-destructive">
                  {formError}
                </p>
              )}

              <div className="rounded-xl border border-border/60 bg-background/50 p-4">
                <div className="flex items-start gap-3">
                  <div className="mt-0.5 flex size-9 items-center justify-center rounded-full bg-primary/10 text-primary">
                    <StorefrontIcon className="size-4" weight="fill" />
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm font-medium">Owner account</p>
                    <p className="text-sm text-muted-foreground">{ownerEmail}</p>
                  </div>
                </div>
              </div>

              <div className="space-y-1.5">
                <Label htmlFor="name">Business name</Label>
                <Input
                  id="name"
                  name="name"
                  value={values.name}
                  onChange={handleChange}
                  placeholder="e.g. Amina Fashion House"
                  autoComplete="organization"
                  className="h-11"
                />
                {errors.name && (touched.name || tried) && (
                  <p className="text-xs text-destructive">{errors.name}</p>
                )}
              </div>

              <div className="overflow-hidden rounded-2xl border border-primary/12 bg-linear-to-br from-primary/10 via-primary/4 to-transparent">
                <div className="flex items-center gap-2 border-b border-primary/10 px-4 py-2.5 text-xs font-semibold tracking-[0.18em] text-primary/80 uppercase">
                  <StorefrontIcon className="size-4" weight="fill" />
                  Temporary storefront link
                </div>
                <div className="space-y-2 px-4 py-4">
                  <div className="rounded-xl border border-primary/12 bg-background/80 px-3 py-3 text-base font-semibold tracking-tight text-foreground shadow-sm shadow-primary/5">
                    <span className="text-muted-foreground">storefront.com/</span>
                    <span>{temporarySlugPreview}</span>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    This temporary link stays private. You can clean it up, claim your final slug,
                    and publish from the Storefront tab after onboarding.
                  </p>
                </div>
              </div>

              <div className="rounded-xl border border-border/60 bg-background/50 px-3 py-3 text-sm text-muted-foreground">
                We may add a short suffix if this temporary link is already taken, but you&apos;ll
                still be able to change it before launch.
              </div>

              <Button type="submit" className="h-11 w-full" disabled={isSubmitting || !ownerEmail}>
                {isSubmitting && <SpinnerGapIcon className="size-4 animate-spin" />}
                Create workspace
              </Button>
            </Form>
          );
        }}
      </Formik>
    </div>
  );
}
