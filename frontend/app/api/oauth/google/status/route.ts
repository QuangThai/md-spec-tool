import { NextRequest, NextResponse } from "next/server";
import {
  decryptToken,
  encryptToken,
  isTokenExpired,
  refreshAccessToken,
  OAUTH_TOKEN_COOKIE,
} from "@/lib/googleOAuth";

export const runtime = "nodejs";

export async function GET(request: NextRequest) {
  const cookieValue = request.cookies.get(OAUTH_TOKEN_COOKIE)?.value;
  if (!cookieValue) {
    return NextResponse.json({ connected: false });
  }

  const token = decryptToken(cookieValue);
  if (!token) {
    return NextResponse.json({ connected: false });
  }

  if (!isTokenExpired(token)) {
    return NextResponse.json({ connected: true, email: token.email });
  }

  if (!token.refreshToken) {
    const response = NextResponse.json({ connected: false });
    response.cookies.delete(OAUTH_TOKEN_COOKIE);
    return response;
  }

  try {
    const refreshed = await refreshAccessToken(token.refreshToken);
    const encrypted = encryptToken(refreshed);
    const response = NextResponse.json({ connected: true, email: refreshed.email });
    response.cookies.set({
      name: OAUTH_TOKEN_COOKIE,
      value: encrypted,
      httpOnly: true,
      sameSite: "lax",
      secure: process.env.NODE_ENV === "production",
      path: "/",
      maxAge: 60 * 60 * 24 * 30,
    });
    return response;
  } catch {
    const response = NextResponse.json({ connected: false });
    response.cookies.delete(OAUTH_TOKEN_COOKIE);
    return response;
  }
}
