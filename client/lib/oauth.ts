import type { SupabaseClient } from "@supabase/supabase-js";

export function buildOAuthCallbackUrl(origin: string, next: string) {
  const callbackUrl = new URL("/auth/callback", origin);
  callbackUrl.searchParams.set("next", next);
  return callbackUrl.toString();
}

export function signInWithGoogleOAuth(supabase: SupabaseClient, origin: string, next: string) {
  return supabase.auth.signInWithOAuth({
    provider: "google",
    options: { redirectTo: buildOAuthCallbackUrl(origin, next) },
  });
}
