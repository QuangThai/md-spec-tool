import { backendClient } from './httpClient';

export interface QuotaStatus {
  session_id: string;
  used_tokens: number;
  limit_tokens: number;
  remaining_tokens: number;
  reset_at: string;
  status: 'ok' | 'exceeded';
  daily_conversions: number;
}

export interface DailyReport {
  date: string;
  session_id?: string;
  user_id?: string;
  tokens_used: number;
  conversions_count: number;
  request_count: number;
}

interface DailyReportResponse {
  reports: DailyReport[];
  period: string;
  count: number;
}

/**
 * Fetch current quota status for the user's session
 * Session ID is automatically added by backendClient interceptor
 */
export async function getQuotaStatus(): Promise<QuotaStatus> {
  const result = await backendClient.safeGet<QuotaStatus>('/api/quota/status');
  if (result.error || !result.data) {
    throw new Error(result.error || 'Failed to fetch quota status');
  }
  return result.data;
}

/**
 * Fetch daily usage reports
 * Session ID is automatically added by backendClient interceptor
 */
export async function getDailyReport(days = 7, aggregateBy = 'session'): Promise<DailyReport[]> {
  const result = await backendClient.safeGet<DailyReportResponse>('/api/quota/daily-report', {
    days: days.toString(),
    aggregate_by: aggregateBy,
  });
  if (result.error || !result.data) {
    throw new Error(result.error || 'Failed to fetch daily report');
  }
  return result.data.reports || [];
}
