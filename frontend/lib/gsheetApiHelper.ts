import { NextRequest, NextResponse } from "next/server";
import {
  decryptToken,
  encryptToken,
  isTokenExpired,
  refreshAccessToken,
  OAUTH_TOKEN_COOKIE,
  type GoogleTokenPayload,
} from "@/lib/googleOAuth";

const API_URL = process.env.NEXT_PUBLIC_API_URL;
if (!API_URL) {
  throw new Error("NEXT_PUBLIC_API_URL is not set");
}

interface TokenResolutionResult {
  token?: string;
  newCookieValue?: string;
}

/**
 * Resolves the access token from the OAuth cookie.
 * If the token is expired, attempts to refresh it using the refresh token.
 * 
 * @returns Object containing the valid access token and optionally a new cookie value
 *          if the token was refreshed.
 */
export async function resolveAccessToken(
  request: NextRequest
): Promise<TokenResolutionResult> {
  const cookieValue = request.cookies.get(OAUTH_TOKEN_COOKIE)?.value;
  if (!cookieValue) {
    return {};
  }

  const payload = decryptToken(cookieValue);
  if (!payload) {
    return {};
  }

  if (!isTokenExpired(payload)) {
    return { token: payload.accessToken };
  }

  if (!payload.refreshToken) {
    return {};
  }

  try {
    const refreshed = await refreshAccessToken(payload.refreshToken);
    return {
      token: refreshed.accessToken,
      newCookieValue: encryptToken(refreshed),
    };
  } catch {
    return {};
  }
}

/**
 * Sets the OAuth cookie on the response if a new cookie value was generated
 * during token refresh.
 */
export function setRefreshedCookie(
  response: NextResponse,
  newCookieValue?: string
): void {
  if (newCookieValue) {
    response.cookies.set({
      name: OAUTH_TOKEN_COOKIE,
      value: newCookieValue,
      httpOnly: true,
      sameSite: "lax",
      secure: process.env.NODE_ENV === "production",
      path: "/",
      maxAge: 60 * 60 * 24 * 30, // 30 days
    });
  }
}

interface ProxyOptions {
  /** The backend API path to proxy to (e.g., "/api/mdflow/gsheet/convert") */
  backendPath: string;
}

/**
 * Proxies a POST request to the Go backend with OAuth token injection.
 * 
 * This function:
 * 1. Resolves the access token from the OAuth cookie
 * 2. Forwards the request body to the backend with Authorization header
 * 3. Returns the backend response, setting a refreshed cookie if needed
 * 
 * @param request - The incoming Next.js request
 * @param options - Configuration options including the backend path
 * @returns NextResponse with the backend response data
 */
export async function proxyToBackend(
  request: NextRequest,
  options: ProxyOptions
): Promise<NextResponse> {
  // async-parallel: Run body read and token resolution in parallel
  const [body, tokenResult] = await Promise.all([
    request.text(),
    resolveAccessToken(request),
  ]);
  
  const { token, newCookieValue } = tokenResult;

  const byokHeader = request.headers.get("x-openai-api-key")?.trim();

  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...(token && { Authorization: `Bearer ${token}` }),
    ...(byokHeader && { "X-OpenAI-API-Key": byokHeader }),
  };

  const backendResponse = await fetch(`${API_URL}${options.backendPath}`, {
    method: "POST",
    headers,
    body,
  });

  // server-serialization: Stream response directly when possible
  const text = await backendResponse.text();
  let data: unknown = null;
  try {
    data = text ? JSON.parse(text) : null;
  } catch {
    data = { error: text };
  }

  const response = NextResponse.json(data, { status: backendResponse.status });
  setRefreshedCookie(response, newCookieValue);

  return response;
}
