import { type NextRequest, NextResponse } from "next/server";

// Routes that require authentication
const PROTECTED_PREFIX = "/app";

export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Only protect /app/* routes
  if (pathname.startsWith(PROTECTED_PREFIX)) {
    // Check for Supabase auth tokens in cookies
    const hasSession =
      request.cookies.has("sb-access-token") ||
      request.cookies.has("sb-refresh-token") ||
      // Supabase SSR uses this pattern
      Array.from(request.cookies.getAll()).some((c) =>
        c.name.includes("-auth-token"),
      );

    if (!hasSession) {
      const loginUrl = new URL("/login", request.url);
      loginUrl.searchParams.set("redirect", pathname);
      return NextResponse.redirect(loginUrl);
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/app/:path*"],
};
