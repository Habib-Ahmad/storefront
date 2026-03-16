"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Formik, Form, Field } from "formik";
import * as Yup from "yup";
import { getSupabase } from "@/lib/supabase";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
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
          <Loader2 className="size-6 animate-spin text-muted-foreground" />
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
          const supabase = getSupabase();
          if (!supabase) {
            toast.error("Auth is not configured");
            return;
          }

          const { error } = await supabase.auth.updateUser({ password: values.password });
          setSubmitting(false);

          if (error) {
            toast.error(error.message);
            return;
          }

          toast.success("Password updated successfully");
          router.replace("/app");
        }}
      >
        {({ isSubmitting, errors, touched }) => (
          <Form className="card-3d rounded-2xl p-6 space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="password">New password</Label>
              <Field
                as={Input}
                id="password"
                name="password"
                type="password"
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
                as={Input}
                id="confirm"
                name="confirm"
                type="password"
                placeholder="Repeat your password"
                autoComplete="new-password"
                className="h-10"
              />
              {errors.confirm && touched.confirm && (
                <p className="text-xs text-destructive">{errors.confirm}</p>
              )}
            </div>

            <Button type="submit" className="w-full h-10" disabled={isSubmitting}>
              {isSubmitting && <Loader2 className="animate-spin" />}
              Update password
            </Button>
          </Form>
        )}
      </Formik>
    </div>
  );
}
