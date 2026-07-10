import { type NextRequest, NextResponse } from "next/server";

export function proxy(request: NextRequest) {
  const hasSession = Boolean(request.cookies.get("cf_session")?.value);
  const { pathname, search } = request.nextUrl;
  if (pathname.startsWith("/app") && !hasSession) {
    const target = request.nextUrl.clone();
    target.pathname = "/sign-in";
    target.search = "";
    target.searchParams.set("next", pathname + search);
    return NextResponse.redirect(target);
  }
  if ((pathname === "/sign-in" || pathname === "/sign-up") && hasSession) {
    const target = request.nextUrl.clone();
    target.pathname = "/app";
    target.search = "";
    return NextResponse.redirect(target);
  }
  return NextResponse.next();
}

export const config = { matcher: ["/app/:path*", "/sign-in", "/sign-up"] };
