"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Formik, Form, Field } from "formik";
import * as Yup from "yup";
import { getSupabase } from "@/lib/supabase";
import { Button } from "@/components/ui/button";
import { PasswordInput } from "@/components/ui/password-input";
import { Label } from "@/components/ui/label";
import { SpinnerGapIcon } from "@phosphor-icons/react";
import { KeySvg } from "@/components/illustrations";

const resetSchema = Yup.object({
  password: Yup.string().min(8, "At least 8 characters").required("Password is required"),
  confirm: Yup.string()
    .oneOf([Yup.ref("password")], "Passwords must match")
    .required("Please confirm your password"),
});

export default function ResetPasswordPage() {
  const router = useRouter();
  const [ready, setReady] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    const supabase = getSupabase();
    if (!supabase) return;

    supabase.auth.onAuthStateChange((event) => {
      if (event === "PASSWORD_RECOVERY") {
        setReady(true);
      }
    });
  }, []);

  if (!ready) {
    return (
      <div className="space-y-6">
        <div className="flex justify-center">
          <KeySvg className="size-28" />
        </div>
        <div className="space-y-2 text-center">
          <h1 className="text-2xl font-bold tracking-tight">Reset password</h1>
          <p className="text-sm text-muted-foreground">
            Verifying your reset link…
          </p>
        </div>
        <div className="flex justify-center py-8">
          <SpinnerGapIcon className="size-6 animate-spin text-muted-foreground" />
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-center">
        <KeySvg className="size-28" />
      </div>

      <div className="space-y-2 text-center">
        <h1 className="text-2xl font-bold tracking-tight">Set new password</h1>
        <p className="text-sm text-muted-foreground">
          Choose a new password for your account
        </p>
      </div>

      <Formik
        initialValues={{ password: "", confirm: "" }}
        validationSchema={resetSchema}
        onSubmit={async (values, { setSubmitting }) => {
          setFormError(null);
          const supabase = getSupabase();
          if (!supabase) {
            setFormError("Auth is not configured");
            return;
          }

          const { error } = await supabase.auth.updateUser({ password: values.password });
          setSubmitting(false);

          if (error) {
            setFormError(error.message);
            return;
          }

          setSuccess(true);
          setTimeout(() => router.replace("/app"), 2000);
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
                Password updated — redirecting…
              </p>
            )}

            <div className="space-y-1.5">
              <Label htmlFor="password">New password</Label>
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
              <Label htmlFor="confirm">Confirm new password</Label>
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
              Update password
            </Button>
          </Form>
        )}
      </Formik>
    </div>
  );
}
