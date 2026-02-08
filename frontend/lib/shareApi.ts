import { backendClient } from './httpClient';
import { ApiResult } from './types';

export type SharePermission = 'view' | 'comment';

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

export async function createShare(payload: SharePayload): Promise<ApiResult<ShareResponse>> {
  return backendClient.safePost<ShareResponse>('/api/share', payload);
}

export async function getShare(key: string): Promise<ApiResult<ShareResponse>> {
  return backendClient.safeGet<ShareResponse>(`/api/share/${encodeURIComponent(key)}`);
}

export async function listPublicShares(): Promise<ApiResult<{ items: ShareSummary[] }>> {
  return backendClient.safeGet<{ items: ShareSummary[] }>('/api/share/public');
}

export async function listComments(key: string): Promise<ApiResult<{ items: CommentResponse[] }>> {
  return backendClient.safeGet<{ items: CommentResponse[] }>(
    `/api/share/${encodeURIComponent(key)}/comments`
  );
}

export async function createComment(
  key: string,
  payload: { author?: string; message: string }
): Promise<ApiResult<CommentResponse>> {
  return backendClient.safePost<CommentResponse>(
    `/api/share/${encodeURIComponent(key)}/comments`,
    payload
  );
}

export async function updateComment(
  key: string,
  commentId: string,
  resolved: boolean
): Promise<ApiResult<CommentResponse>> {
  return backendClient.safePatch<CommentResponse>(
    `/api/share/${encodeURIComponent(key)}/comments/${encodeURIComponent(commentId)}`,
    { resolved }
  );
}

export async function updateShare(
  key: string,
  payload: { is_public?: boolean; allow_comments?: boolean }
): Promise<ApiResult<ShareResponse>> {
  return backendClient.safePatch<ShareResponse>(
    `/api/share/${encodeURIComponent(key)}`,
    payload
  );
}
