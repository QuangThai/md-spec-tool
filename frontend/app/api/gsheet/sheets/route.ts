import { NextRequest, NextResponse } from "next/server";
import {
  decryptToken,
  encryptToken,
  isTokenExpired,
  refreshAccessToken,
  OAUTH_TOKEN_COOKIE,
} from "@/lib/googleOAuth";

export const runtime = "nodejs";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

async function resolveAccessToken(request: NextRequest) {
  const cookieValue = request.cookies.get(OAUTH_TOKEN_COOKIE)?.value;
  if (!cookieValue) return {} as { token?: string; cookieValue?: string };

  const payload = decryptToken(cookieValue);
  if (!payload) return {} as { token?: string; cookieValue?: string };

  if (!isTokenExpired(payload)) {
    return { token: payload.accessToken };
  }

  if (!payload.refreshToken) {
    return {} as { token?: string; cookieValue?: string };
  }

  try {
    const refreshed = await refreshAccessToken(payload.refreshToken);
    return {
      token: refreshed.accessToken,
      cookieValue: encryptToken(refreshed),
    };
  } catch {
    return {} as { token?: string; cookieValue?: string };
  }
}

export async function POST(request: NextRequest) {
  const body = await request.text();
  const { token, cookieValue } = await resolveAccessToken(request);

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const backendResponse = await fetch(`${API_URL}/api/mdflow/gsheet/sheets`, {
    method: "POST",
    headers,
    body,
  });

  const text = await backendResponse.text();
  let data: unknown = null;
  try {
    data = text ? JSON.parse(text) : null;
  } catch {
    data = { error: text };
  }

  const response = NextResponse.json(data, { status: backendResponse.status });
  if (cookieValue) {
    response.cookies.set({
      name: OAUTH_TOKEN_COOKIE,
      value: cookieValue,
      httpOnly: true,
      sameSite: "lax",
      secure: process.env.NODE_ENV === "production",
      path: "/",
      maxAge: 60 * 60 * 24 * 30,
    });
  }

  return response;
}
