"use client";

import { useRouter } from "next/navigation";
import Link from "next/link";
import { Formik, Form, Field } from "formik";
import * as Yup from "yup";
import { getSupabase } from "@/lib/supabase";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
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
          const supabase = getSupabase();
          if (!supabase) {
            toast.error("Auth is not configured");
            return;
          }

          const { error } = await supabase.auth.signUp({
            email: values.email,
            password: values.password,
          });

          if (error) {
            toast.error(error.message);
            setSubmitting(false);
            return;
          }

          toast.success("Check your email to confirm your account");
          router.push("/login");
        }}
      >
        {({ isSubmitting, errors, touched }) => (
          <Form className="card-3d rounded-2xl p-6 space-y-4">
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
              <Label htmlFor="confirm">Confirm password</Label>
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
              Create account
            </Button>
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
