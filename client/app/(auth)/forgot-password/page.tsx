"use client";

import { useState } from "react";
import Link from "next/link";
import { Formik, Form, Field } from "formik";
import * as Yup from "yup";
import { getSupabase } from "@/lib/supabase";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ArrowLeftIcon, SpinnerGapIcon } from "@phosphor-icons/react";
import { MailSentSvg } from "@/components/illustrations";

const forgotSchema = Yup.object({
  email: Yup.string().email("Invalid email").required("Email is required"),
});

export default function ForgotPasswordPage() {
  const [sent, setSent] = useState(false);
  const [sentEmail, setSentEmail] = useState("");
  const [formError, setFormError] = useState<string | null>(null);

  if (sent) {
    return (
      <div className="space-y-6">
        <div className="flex justify-center">
          <MailSentSvg className="size-28" />
        </div>

        <div className="space-y-2 text-center">
          <h1 className="text-2xl font-bold tracking-tight">Check your email</h1>
          <p className="text-sm text-muted-foreground">
            We sent a password reset link to{" "}
            <span className="font-medium text-foreground">{sentEmail}</span>
          </p>
        </div>

        <div className="card-3d space-y-4 rounded-2xl p-6">
          <p className="text-center text-sm text-muted-foreground">
            Didn&apos;t get the email? Check your spam folder or try again.
          </p>
          <Button variant="outline" className="h-10 w-full" onClick={() => setSent(false)}>
            Try again
          </Button>
        </div>

        <p className="text-center">
          <Link
            href="/login"
            className="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-primary"
          >
            <ArrowLeftIcon className="size-3.5" />
            Back to sign in
          </Link>
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-center">
        <MailSentSvg className="size-28" />
      </div>

      <div className="space-y-2 text-center">
        <h1 className="text-2xl font-bold tracking-tight">Forgot password?</h1>
        <p className="text-sm text-muted-foreground">
          Enter your email and we&apos;ll send you a reset link
        </p>
      </div>

      <Formik
        initialValues={{ email: "" }}
        validationSchema={forgotSchema}
        onSubmit={async (values, { setSubmitting }) => {
          setFormError(null);
          const supabase = getSupabase();
          if (!supabase) {
            setFormError("Auth is not configured");
            return;
          }

          const { error } = await supabase.auth.resetPasswordForEmail(values.email, {
            redirectTo: `${window.location.origin}/reset-password`,
          });

          setSubmitting(false);

          if (error) {
            setFormError(error.message);
            return;
          }

          setSentEmail(values.email);
          setSent(true);
        }}
      >
        {({ isSubmitting, errors, touched }) => (
          <Form className="card-3d space-y-4 rounded-2xl p-6">
            {formError && (
              <p className="rounded-lg bg-destructive/10 px-3 py-2 text-center text-sm text-destructive">
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

            <Button type="submit" className="h-10 w-full" disabled={isSubmitting}>
              {isSubmitting && <SpinnerGapIcon className="size-4 animate-spin" />}
              Send reset link
            </Button>
          </Form>
        )}
      </Formik>

      <p className="text-center">
        <Link
          href="/login"
          className="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-primary"
        >
          <ArrowLeftIcon className="size-3.5" />
          Back to sign in
        </Link>
      </p>
    </div>
  );
}
