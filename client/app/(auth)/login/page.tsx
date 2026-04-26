"use client";

import { Suspense, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { Formik, Form, Field } from "formik";
import * as Yup from "yup";
import { getSupabase } from "@/lib/supabase";
import { api } from "@/lib/api";
import { signInWithGoogleOAuth } from "@/lib/oauth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { SpinnerGapIcon, GoogleLogoIcon } from "@phosphor-icons/react";
import { ShoppingBagSvg } from "@/components/illustrations";

const loginSchema = Yup.object({
  email: Yup.string().email("Invalid email").required("Email is required"),
  password: Yup.string().required("Password is required"),
});

function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const redirect = searchParams.get("redirect") ?? "/app";
  const oauthError =
    searchParams.get("error") === "oauth_callback"
      ? "Google sign-in could not be completed. Please try again."
      : null;
  const [formError, setFormError] = useState<string | null>(null);

  async function handleGoogleSignIn() {
    const supabase = getSupabase();
    if (!supabase) {
      setFormError("Auth is not configured");
      return;
    }

    await signInWithGoogleOAuth(supabase, window.location.origin, redirect);
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-center">
        <ShoppingBagSvg className="size-32" />
      </div>

      <div className="space-y-2 text-center">
        <h1 className="text-2xl font-bold tracking-tight">Welcome back</h1>
        <p className="text-sm text-muted-foreground">Sign in to your store</p>
      </div>

      <Formik
        initialValues={{ email: "", password: "" }}
        validationSchema={loginSchema}
        onSubmit={async (values, { setSubmitting }) => {
          setFormError(null);
          const supabase = getSupabase();
          if (!supabase) {
            setFormError("Auth is not configured");
            return;
          }

          const { error } = await supabase.auth.signInWithPassword(values);
          if (error) {
            setFormError(error.message);
            setSubmitting(false);
            return;
          }

          try {
            const me = await api.getMe();
            router.replace(me.onboarded ? redirect : "/onboard");
          } catch {
            router.replace(redirect);
          }
        }}
      >
        {({ isSubmitting, errors, touched }) => (
          <Form className="md:card-3d space-y-4 px-1 md:rounded-2xl md:border md:border-border/60 md:bg-background/72 md:p-6 md:shadow-lg md:shadow-black/5">
            {(formError ?? oauthError) && (
              <p className="rounded-lg bg-destructive/10 px-3 py-2 text-center text-sm text-destructive">
                {formError ?? oauthError}
              </p>
            )}

            <div className="space-y-1.5">
              <Label htmlFor="email">Email</Label>
              <Field
                as={Input}
                id="email"
                name="email"
                type="email"
                placeholder="you@example.com"
                autoComplete="email"
                className="h-10"
              />
              {errors.email && touched.email && (
                <p className="text-xs text-destructive">{errors.email}</p>
              )}
            </div>

            <div className="space-y-1.5">
              <div className="flex items-center justify-between">
                <Label htmlFor="password">Password</Label>
                <Link
                  href="/forgot-password"
                  className="text-xs text-muted-foreground transition-colors hover:text-primary"
                >
                  Forgot password?
                </Link>
              </div>
              <Field
                as={PasswordInput}
                id="password"
                name="password"
                placeholder="••••••••"
                autoComplete="current-password"
                className="h-10"
              />
              {errors.password && touched.password && (
                <p className="text-xs text-destructive">{errors.password}</p>
              )}
            </div>

            <Button type="submit" className="h-10 w-full" disabled={isSubmitting}>
              {isSubmitting && <SpinnerGapIcon className="size-4 animate-spin" />}
              Sign in
            </Button>

            <div className="relative">
              <Separator />
              <span className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-card px-2 text-xs text-muted-foreground">
                or
              </span>
            </div>

            <div>
              <Button
                type="button"
                variant="outline"
                className="h-10 w-full gap-2"
                onClick={handleGoogleSignIn}
              >
                <GoogleLogoIcon className="size-4" weight="bold" />
                Google
              </Button>
            </div>
          </Form>
        )}
      </Formik>

      <p className="text-center text-sm text-muted-foreground">
        Don&apos;t have an account?{" "}
        <Link href="/signup" className="font-medium text-primary hover:underline">
          Sign up
        </Link>
      </p>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense>
      <LoginForm />
    </Suspense>
  );
}
