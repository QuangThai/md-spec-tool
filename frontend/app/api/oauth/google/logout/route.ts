import { NextResponse } from "next/server";
import { OAUTH_TOKEN_COOKIE } from "@/lib/googleOAuth";

export const runtime = "nodejs";

export async function POST() {
  const response = NextResponse.json({ connected: false });
  response.cookies.delete(OAUTH_TOKEN_COOKIE);
  return response;
}
