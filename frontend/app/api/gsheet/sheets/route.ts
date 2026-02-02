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
  if (!cookieValue) {
    console.log("[gsheet/sheets] No OAuth cookie found");
    return {} as { token?: string; cookieValue?: string };
  }

  const payload = decryptToken(cookieValue);
  if (!payload) {
    console.log("[gsheet/sheets] Failed to decrypt token cookie");
    return {} as { token?: string; cookieValue?: string };
  }

  if (!isTokenExpired(payload)) {
    console.log("[gsheet/sheets] Using valid access token");
    return { token: payload.accessToken };
  }

  console.log("[gsheet/sheets] Token expired, attempting refresh");
  if (!payload.refreshToken) {
    console.log("[gsheet/sheets] No refresh token available");
    return {} as { token?: string; cookieValue?: string };
  }

  try {
    const refreshed = await refreshAccessToken(payload.refreshToken);
    console.log("[gsheet/sheets] Token refreshed successfully");
    return {
      token: refreshed.accessToken,
      cookieValue: encryptToken(refreshed),
    };
  } catch (err) {
    console.log("[gsheet/sheets] Token refresh failed:", err);
    return {} as { token?: string; cookieValue?: string };
  }
}

export async function POST(request: NextRequest) {
  const body = await request.text();
  const { token, cookieValue } = await resolveAccessToken(request);

  console.log("[gsheet/sheets] Token present:", !!token);
  console.log("[gsheet/sheets] Calling backend:", `${API_URL}/api/mdflow/gsheet/sheets`);

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
  console.log("[gsheet/sheets] Backend status:", backendResponse.status);
  
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
