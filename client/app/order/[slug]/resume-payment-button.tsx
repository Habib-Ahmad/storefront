"use client";

import { useState } from "react";
import { ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useResumeTrackedOrderPayment } from "@/hooks/use-orders";
import { ApiError } from "@/lib/api";

export function ResumePaymentButton({ slug }: { slug: string }) {
  const resumePayment = useResumeTrackedOrderPayment();
  const [error, setError] = useState<string | null>(null);

  return (
    <div className="space-y-3">
      {error ? (
        <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">{error}</p>
      ) : null}
      <Button
        type="button"
        className="inline-flex items-center justify-center gap-2 rounded-full px-5 py-3 text-sm font-medium"
        disabled={resumePayment.isPending}
        onClick={async () => {
          try {
            setError(null);
            const response = await resumePayment.mutateAsync(slug);
            window.location.href = response.authorization_url;
          } catch (err) {
            if (err instanceof ApiError) {
              setError(err.message);
            } else {
              setError("Unable to continue payment");
            }
          }
        }}
      >
        {resumePayment.isPending ? "Opening payment..." : "Continue payment"}
        <ArrowRight className="h-4 w-4" />
      </Button>
    </div>
  );
}
