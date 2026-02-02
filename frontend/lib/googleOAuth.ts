import crypto from "crypto";

const GOOGLE_OAUTH_AUTH_URL = "https://accounts.google.com/o/oauth2/v2/auth";
const GOOGLE_OAUTH_TOKEN_URL = "https://oauth2.googleapis.com/token";
const GOOGLE_SHEETS_SCOPE = "https://www.googleapis.com/auth/spreadsheets.readonly";

export const OAUTH_STATE_COOKIE = "google_oauth_state";
export const OAUTH_TOKEN_COOKIE = "google_gsheet_token";

export interface GoogleTokenPayload {
  accessToken: string;
  refreshToken?: string;
  expiresAt: number;
  email?: string;
}

function getCookieSecret(): Buffer {
  const secret = process.env.COOKIE_SECRET || "";
  if (!secret) {
    throw new Error("COOKIE_SECRET is not set");
  }
  if (secret.length >= 32) {
    return Buffer.from(secret.slice(0, 32));
  }
  return crypto.createHash("sha256").update(secret).digest();
}

export function buildOAuthRedirectUri(origin: string): string {
  return `${origin}/oauth/callback`;
}

export function buildGoogleAuthUrl(origin: string, state: string): string {
  const redirectUri = buildOAuthRedirectUri(origin);
  const url = new URL(GOOGLE_OAUTH_AUTH_URL);
  url.searchParams.set("client_id", getOAuthClientId());
  url.searchParams.set("redirect_uri", redirectUri);
  url.searchParams.set("response_type", "code");
  url.searchParams.set("scope", GOOGLE_SHEETS_SCOPE);
  url.searchParams.set("access_type", "offline");
  url.searchParams.set("include_granted_scopes", "true");
  url.searchParams.set("prompt", "consent");
  url.searchParams.set("state", state);
  return url.toString();
}

export function encryptToken(payload: GoogleTokenPayload): string {
  const iv = crypto.randomBytes(12);
  const cipher = crypto.createCipheriv("aes-256-gcm", getCookieSecret(), iv);
  const plaintext = Buffer.from(JSON.stringify(payload), "utf8");
  const ciphertext = Buffer.concat([cipher.update(plaintext), cipher.final()]);
  const tag = cipher.getAuthTag();
  return Buffer.concat([iv, ciphertext, tag]).toString("base64url");
}

export function decryptToken(value?: string): GoogleTokenPayload | null {
  if (!value) return null;
  try {
    const buffer = Buffer.from(value, "base64url");
    if (buffer.length < 12 + 16) return null;
    const iv = buffer.subarray(0, 12);
    const tag = buffer.subarray(buffer.length - 16);
    const ciphertext = buffer.subarray(12, buffer.length - 16);
    const decipher = crypto.createDecipheriv("aes-256-gcm", getCookieSecret(), iv);
    decipher.setAuthTag(tag);
    const plaintext = Buffer.concat([decipher.update(ciphertext), decipher.final()]);
    const parsed = JSON.parse(plaintext.toString("utf8")) as GoogleTokenPayload;
    if (!parsed.accessToken || !parsed.expiresAt) return null;
    return parsed;
  } catch {
    return null;
  }
}

export function isTokenExpired(payload: GoogleTokenPayload, skewMs = 60_000): boolean {
  return Date.now() + skewMs >= payload.expiresAt;
}

export async function exchangeCodeForTokens(code: string, origin: string): Promise<GoogleTokenPayload> {
  const redirectUri = buildOAuthRedirectUri(origin);
  const body = new URLSearchParams({
    code,
    client_id: getOAuthClientId(),
    client_secret: getOAuthClientSecret(),
    redirect_uri: redirectUri,
    grant_type: "authorization_code",
  });

  const response = await fetch(GOOGLE_OAUTH_TOKEN_URL, {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
    },
    body,
  });

  if (!response.ok) {
    const errorBody = await response.text();
    throw new Error(errorBody || `Token exchange failed: ${response.status}`);
  }

  const data = (await response.json()) as {
    access_token: string;
    refresh_token?: string;
    expires_in: number;
  };

  const expiresAt = Date.now() + Math.max(0, data.expires_in - 60) * 1000;
  return {
    accessToken: data.access_token,
    refreshToken: data.refresh_token,
    expiresAt,
  };
}

export async function refreshAccessToken(refreshToken: string): Promise<GoogleTokenPayload> {
  const body = new URLSearchParams({
    client_id: getOAuthClientId(),
    client_secret: getOAuthClientSecret(),
    refresh_token: refreshToken,
    grant_type: "refresh_token",
  });

  const response = await fetch(GOOGLE_OAUTH_TOKEN_URL, {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
    },
    body,
  });

  if (!response.ok) {
    const errorBody = await response.text();
    throw new Error(errorBody || `Token refresh failed: ${response.status}`);
  }

  const data = (await response.json()) as {
    access_token: string;
    expires_in: number;
  };

  const expiresAt = Date.now() + Math.max(0, data.expires_in - 60) * 1000;
  return {
    accessToken: data.access_token,
    refreshToken,
    expiresAt,
  };
}

function getOAuthClientId(): string {
  const value = process.env.GOOGLE_OAUTH_CLIENT_ID || "";
  if (!value) {
    throw new Error("GOOGLE_OAUTH_CLIENT_ID is not set");
  }
  return value;
}

function getOAuthClientSecret(): string {
  const value = process.env.GOOGLE_OAUTH_CLIENT_SECRET || "";
  if (!value) {
    throw new Error("GOOGLE_OAUTH_CLIENT_SECRET is not set");
  }
  return value;
}
