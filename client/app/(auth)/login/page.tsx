"use client";

import { Suspense, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { Formik, Form, Field } from "formik";
import * as Yup from "yup";
import { getSupabase } from "@/lib/supabase";
import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { SpinnerGapIcon, GoogleLogoIcon, AppleLogoIcon } from "@phosphor-icons/react";
import { ShoppingBagSvg } from "@/components/illustrations";

const loginSchema = Yup.object({
  email: Yup.string().email("Invalid email").required("Email is required"),
  password: Yup.string().required("Password is required"),
});

function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const redirect = searchParams.get("redirect") ?? "/app";
  const [formError, setFormError] = useState<string | null>(null);

  return (
    <div className="space-y-6">
      <div className="flex justify-center">
        <ShoppingBagSvg className="size-32" />
      </div>

      <div className="space-y-2 text-center">
        <h1 className="text-2xl font-bold tracking-tight">Welcome back</h1>
        <p className="text-sm text-muted-foreground">
          Sign in to your store
        </p>
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
          <Form className="card-3d rounded-2xl p-6 space-y-4">
            {formError && (
              <p className="text-sm text-destructive text-center bg-destructive/10 rounded-lg px-3 py-2">
                {formError}
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
                  className="text-xs text-muted-foreground hover:text-primary transition-colors"
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

            <Button type="submit" className="w-full h-10" disabled={isSubmitting}>
              {isSubmitting && <SpinnerGapIcon className="size-4 animate-spin" />}
              Sign in
            </Button>

            <div className="relative">
              <Separator />
              <span className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 bg-card px-2 text-xs text-muted-foreground">
                or
              </span>
            </div>

            <div className="grid grid-cols-2 gap-3">
              <Button
                type="button"
                variant="outline"
                className="h-10 gap-2"
                onClick={() => {
                  const supabase = getSupabase();
                  supabase?.auth.signInWithOAuth({ provider: "google", options: { redirectTo: `${window.location.origin}/app` } });
                }}
              >
                <GoogleLogoIcon className="size-4" weight="bold" />
                Google
              </Button>
              <Button
                type="button"
                variant="outline"
                className="h-10 gap-2"
                onClick={() => {
                  const supabase = getSupabase();
                  supabase?.auth.signInWithOAuth({ provider: "apple", options: { redirectTo: `${window.location.origin}/app` } });
                }}
              >
                <AppleLogoIcon className="size-4" weight="fill" />
                Apple
              </Button>
            </div>
          </Form>
        )}
      </Formik>

      <p className="text-center text-sm text-muted-foreground">
        Don&apos;t have an account?{" "}
        <Link href="/signup" className="text-primary hover:underline font-medium">
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
