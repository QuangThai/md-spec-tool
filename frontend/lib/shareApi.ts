import { ApiResult } from "@/lib/types";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export type SharePermission = "view" | "comment";

export interface SharePayload {
  title?: string;
  template?: string;
  mdflow: string;
  slug?: string;
  is_public?: boolean;
  allow_comments?: boolean;
  permission?: SharePermission;
}

export interface ShareResponse {
  token: string;
  slug: string;
  title: string;
  template: string;
  mdflow: string;
  is_public: boolean;
  allow_comments: boolean;
  permission: SharePermission;
  created_at: string;
}

export interface CommentResponse {
  id: string;
  author: string;
  message: string;
  resolved: boolean;
  created_at: string;
}

export interface ShareSummary {
  slug: string;
  title: string;
  template: string;
  created_at: string;
}

async function apiCall<T>(url: string, options?: RequestInit): Promise<ApiResult<T>> {
  try {
    const response = await fetch(url, options);

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      return { error: errorData.error || `HTTP ${response.status}` };
    }

    const data = await response.json();
    return { data };
  } catch (error) {
    return { error: error instanceof Error ? error.message : "Network error" };
  }
}

export async function createShare(payload: SharePayload): Promise<ApiResult<ShareResponse>> {
  return apiCall<ShareResponse>(`${API_URL}/api/share`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

export async function getShare(key: string): Promise<ApiResult<ShareResponse>> {
  return apiCall<ShareResponse>(`${API_URL}/api/share/${encodeURIComponent(key)}`);
}

export async function listPublicShares(): Promise<ApiResult<{ items: ShareSummary[] }>> {
  return apiCall<{ items: ShareSummary[] }>(`${API_URL}/api/share/public`);
}

export async function listComments(key: string): Promise<ApiResult<{ items: CommentResponse[] }>> {
  return apiCall<{ items: CommentResponse[] }>(
    `${API_URL}/api/share/${encodeURIComponent(key)}/comments`
  );
}

export async function createComment(
  key: string,
  payload: { author?: string; message: string }
): Promise<ApiResult<CommentResponse>> {
  return apiCall<CommentResponse>(`${API_URL}/api/share/${encodeURIComponent(key)}/comments`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

export async function updateComment(
  key: string,
  commentId: string,
  resolved: boolean
): Promise<ApiResult<CommentResponse>> {
  return apiCall<CommentResponse>(
    `${API_URL}/api/share/${encodeURIComponent(key)}/comments/${encodeURIComponent(commentId)}`,
    {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ resolved }),
    }
  );
}

export async function updateShare(
  key: string,
  payload: { is_public?: boolean; allow_comments?: boolean }
): Promise<ApiResult<ShareResponse>> {
  return apiCall<ShareResponse>(`${API_URL}/api/share/${encodeURIComponent(key)}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}
