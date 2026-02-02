import { NextRequest, NextResponse } from "next/server";
import crypto from "crypto";
import { buildGoogleAuthUrl, OAUTH_STATE_COOKIE } from "@/lib/googleOAuth";

export const runtime = "nodejs";

function getOrigin(request: NextRequest): string {
  const proto = request.headers.get("x-forwarded-proto") || "http";
  const host =
    request.headers.get("x-forwarded-host") || request.headers.get("host");
  return `${proto}://${host}`;
}

export async function GET(request: NextRequest) {
  const state = crypto.randomBytes(16).toString("hex");
  const origin = getOrigin(request);
  const authUrl = buildGoogleAuthUrl(origin, state);

  const response = NextResponse.redirect(authUrl);
  response.cookies.set({
    name: OAUTH_STATE_COOKIE,
    value: state,
    httpOnly: true,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/",
    maxAge: 60 * 10,
  });

  return response;
}
