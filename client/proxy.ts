import { type NextRequest, NextResponse } from "next/server";

export function proxy(_request: NextRequest) {
  // Route protection disabled during development.
  // Will be re-enabled when auth flow is finalized.
  return NextResponse.next();
}

export const config = {
  matcher: [],
};
