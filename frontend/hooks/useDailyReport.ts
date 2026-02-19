import { useQuery } from '@tanstack/react-query';
import { DailyReport, getDailyReport } from '@/lib/quotaApi';

interface UseDailyReportOptions {
  enabled?: boolean;
  days?: number;
  aggregateBy?: string;
}

export function useDailyReport(options: UseDailyReportOptions = {}) {
  const { enabled = true, days = 7, aggregateBy = 'session' } = options;

  const { data: reports, isLoading: loading, error } = useQuery({
    queryKey: ['daily-report', days, aggregateBy],
    queryFn: () => getDailyReport(days, aggregateBy),
    enabled,
    staleTime: 5 * 60 * 1000,
  });

  return {
    reports: reports ?? [],
    loading,
    error: error instanceof Error ? error.message : null,
  };
}
