import { nextApiClient } from './httpClient';
import { ApiResult } from './types';

export interface GoogleAuthStatus {
  connected: boolean;
  email?: string;
}

export async function getGoogleAuthStatus(): Promise<ApiResult<GoogleAuthStatus>> {
  return nextApiClient.safeGet<GoogleAuthStatus>('/api/oauth/google/status');
}

export async function logoutGoogle(): Promise<ApiResult<void>> {
  return nextApiClient.safePost<void>('/api/oauth/google/logout');
}

export function getGoogleLoginUrl(): string {
  return '/api/oauth/google/login';
}
