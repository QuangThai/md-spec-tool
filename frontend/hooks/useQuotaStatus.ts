import { useQuery } from '@tanstack/react-query';
import { QuotaStatus, getQuotaStatus } from '@/lib/quotaApi';

interface UseQuotaStatusOptions {
  enabled?: boolean;
  pollingInterval?: number; // milliseconds, default 30s
}

/**
 * Hook to fetch and poll quota status
 * Uses TanStack Query for caching, deduplication, and automatic polling
 * Session ID is automatically added by backendClient interceptor
 */
export function useQuotaStatus(options: UseQuotaStatusOptions = {}) {
  const { enabled = true, pollingInterval = 30000 } = options;

  const { data: quota, isLoading: loading, error, refetch } = useQuery({
    queryKey: ['quota'],
    queryFn: getQuotaStatus,
    enabled,
    refetchInterval: pollingInterval,
    staleTime: pollingInterval / 2, // Cache for half the polling interval
  });

  return {
    quota: quota ?? null,
    loading,
    error: error instanceof Error ? error.message : null,
    refetch,
  };
}
