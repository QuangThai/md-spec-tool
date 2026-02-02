import { NextRequest, NextResponse } from "next/server";
import {
  encryptToken,
  exchangeCodeForTokens,
  OAUTH_STATE_COOKIE,
  OAUTH_TOKEN_COOKIE,
} from "@/lib/googleOAuth";

export const runtime = "nodejs";

function getOrigin(request: NextRequest): string {
  const proto = request.headers.get("x-forwarded-proto") || "http";
  const host =
    request.headers.get("x-forwarded-host") || request.headers.get("host");
  return `${proto}://${host}`;
}

export async function POST(request: NextRequest) {
  const body = await request.json().catch(() => null);
  const code = body?.code as string | undefined;
  const state = body?.state as string | undefined;

  if (!code || !state) {
    const response = NextResponse.json(
      { error: "Missing code or state" },
      { status: 400 }
    );
    response.cookies.delete(OAUTH_STATE_COOKIE);
    return response;
  }

  const cookieState = request.cookies.get(OAUTH_STATE_COOKIE)?.value;
  if (!cookieState || cookieState !== state) {
    const response = NextResponse.json(
      { error: "Invalid OAuth state" },
      { status: 400 }
    );
    response.cookies.delete(OAUTH_STATE_COOKIE);
    return response;
  }

  try {
    const origin = getOrigin(request);
    const tokenPayload = await exchangeCodeForTokens(code, origin);

    // Extract user email from Google userinfo endpoint
    try {
      const userInfoRes = await fetch("https://www.googleapis.com/oauth2/v1/userinfo", {
        headers: { Authorization: `Bearer ${tokenPayload.accessToken}` },
      });
      if (userInfoRes.ok) {
        const userInfo = (await userInfoRes.json()) as { email?: string };
        if (userInfo.email) {
          tokenPayload.email = userInfo.email;
        }
      }
    } catch {
      // Continue without email - not critical
    }

    const encrypted = encryptToken(tokenPayload);
    const response = NextResponse.json({ connected: true, email: tokenPayload.email });
    response.cookies.set({
      name: OAUTH_TOKEN_COOKIE,
      value: encrypted,
      httpOnly: true,
      sameSite: "lax",
      secure: process.env.NODE_ENV === "production",
      path: "/",
      maxAge: 60 * 60 * 24 * 30,
    });
    response.cookies.delete(OAUTH_STATE_COOKIE);
    return response;
  } catch (error) {
    const response = NextResponse.json(
      {
        error: error instanceof Error ? error.message : "OAuth exchange failed",
      },
      { status: 400 }
    );
    response.cookies.delete(OAUTH_STATE_COOKIE);
    return response;
  }
}
