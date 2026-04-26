"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Formik, Form, Field } from "formik";
import * as Yup from "yup";
import { getSupabase } from "@/lib/supabase";
import { signInWithGoogleOAuth } from "@/lib/oauth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { SpinnerGapIcon, GoogleLogoIcon } from "@phosphor-icons/react";
import { ShoppingBagSvg } from "@/components/illustrations";

const signupSchema = Yup.object({
  email: Yup.string().email("Invalid email").required("Email is required"),
  password: Yup.string().min(8, "At least 8 characters").required("Password is required"),
  confirm: Yup.string()
    .oneOf([Yup.ref("password")], "Passwords must match")
    .required("Please confirm your password"),
});

function isExistingAccountError(message: string) {
  return /already registered|already exists|already have an account|user already registered/i.test(
    message,
  );
}

export default function SignupPage() {
  const router = useRouter();
  const [formError, setFormError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [accountExists, setAccountExists] = useState(false);

  async function handleGoogleSignIn() {
    const supabase = getSupabase();
    if (!supabase) {
      setFormError("Auth is not configured");
      return;
    }

    await signInWithGoogleOAuth(supabase, window.location.origin, "/onboard");
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-center">
        <ShoppingBagSvg className="size-28" />
      </div>

      <div className="space-y-2 text-center">
        <h1 className="text-2xl font-bold tracking-tight">Create an account</h1>
        <p className="text-sm text-muted-foreground">Get started with Storefront</p>
      </div>

      <Formik
        initialValues={{ email: "", password: "", confirm: "" }}
        validationSchema={signupSchema}
        onSubmit={async (values, { setSubmitting }) => {
          setFormError(null);
          setAccountExists(false);
          const supabase = getSupabase();
          if (!supabase) {
            setFormError("Auth is not configured");
            return;
          }

          const { data, error } = await supabase.auth.signUp({
            email: values.email,
            password: values.password,
          });

          if (error) {
            if (isExistingAccountError(error.message)) {
              setAccountExists(true);
              setFormError("You already have an account. Sign in to continue.");
            } else {
              setFormError(error.message);
            }
            setSubmitting(false);
            return;
          }

          setSuccess(true);
          setTimeout(() => {
            router.push(data.session ? "/onboard" : "/login");
          }, 2000);
        }}
      >
        {({ isSubmitting, errors, touched }) => (
          <Form className="md:card-3d space-y-4 px-1 md:rounded-2xl md:border md:border-border/60 md:bg-background/72 md:p-6 md:shadow-lg md:shadow-black/5">
            {formError && (
              <p className="rounded-lg bg-destructive/10 px-3 py-2 text-center text-sm text-destructive">
                {formError}
              </p>
            )}

            {accountExists && (
              <p className="rounded-lg border border-primary/15 bg-primary/8 px-3 py-2 text-center text-sm text-primary">
                Sign in instead and continue setting up your store.
              </p>
            )}

            {success && (
              <p className="rounded-lg bg-primary/10 px-3 py-2 text-center text-sm text-primary">
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

            <Button type="submit" className="h-10 w-full" disabled={isSubmitting || success}>
              {isSubmitting && <SpinnerGapIcon className="size-4 animate-spin" />}
              Create account
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
        Already have an account?{" "}
        <Link href="/login" className="font-medium text-primary hover:underline">
          Sign in
        </Link>
      </p>
    </div>
  );
}
