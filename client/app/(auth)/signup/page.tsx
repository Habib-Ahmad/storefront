"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Formik, Form, Field } from "formik";
import * as Yup from "yup";
import { getSupabase } from "@/lib/supabase";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { SpinnerGapIcon, GoogleLogoIcon, AppleLogoIcon } from "@phosphor-icons/react";
import { ShoppingBagSvg } from "@/components/illustrations";

const signupSchema = Yup.object({
  email: Yup.string().email("Invalid email").required("Email is required"),
  password: Yup.string().min(8, "At least 8 characters").required("Password is required"),
  confirm: Yup.string()
    .oneOf([Yup.ref("password")], "Passwords must match")
    .required("Please confirm your password"),
});

export default function SignupPage() {
  const router = useRouter();
  const [formError, setFormError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  return (
    <div className="space-y-6">
      <div className="flex justify-center">
        <ShoppingBagSvg className="size-28" />
      </div>

      <div className="space-y-2 text-center">
        <h1 className="text-2xl font-bold tracking-tight">Create an account</h1>
        <p className="text-sm text-muted-foreground">
          Get started with Storefront
        </p>
      </div>

      <Formik
        initialValues={{ email: "", password: "", confirm: "" }}
        validationSchema={signupSchema}
        onSubmit={async (values, { setSubmitting }) => {
          setFormError(null);
          const supabase = getSupabase();
          if (!supabase) {
            setFormError("Auth is not configured");
            return;
          }

          const { error } = await supabase.auth.signUp({
            email: values.email,
            password: values.password,
          });

          if (error) {
            setFormError(error.message);
            setSubmitting(false);
            return;
          }

          setSuccess(true);
          setTimeout(() => router.push("/login"), 3000);
        }}
      >
        {({ isSubmitting, errors, touched }) => (
          <Form className="card-3d rounded-2xl p-6 space-y-4">
            {formError && (
              <p className="text-sm text-destructive text-center bg-destructive/10 rounded-lg px-3 py-2">
                {formError}
              </p>
            )}

            {success && (
              <p className="text-sm text-primary text-center bg-primary/10 rounded-lg px-3 py-2">
                Check your email to confirm your account
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
              <Label htmlFor="password">Password</Label>
              <Field
                as={PasswordInput}
                id="password"
                name="password"
                placeholder="At least 8 characters"
                autoComplete="new-password"
                className="h-10"
              />
              {errors.password && touched.password && (
                <p className="text-xs text-destructive">{errors.password}</p>
              )}
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="confirm">Confirm password</Label>
              <Field
                as={PasswordInput}
                id="confirm"
                name="confirm"
                placeholder="Repeat your password"
                autoComplete="new-password"
                className="h-10"
              />
              {errors.confirm && touched.confirm && (
                <p className="text-xs text-destructive">{errors.confirm}</p>
              )}
            </div>

            <Button type="submit" className="w-full h-10" disabled={isSubmitting || success}>
              {isSubmitting && <SpinnerGapIcon className="size-4 animate-spin" />}
              Create account
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
        Already have an account?{" "}
        <Link href="/login" className="text-primary hover:underline font-medium">
          Sign in
        </Link>
      </p>
    </div>
  );
}
