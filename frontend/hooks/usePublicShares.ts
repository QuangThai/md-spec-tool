import { listPublicShares } from "@/lib/shareApi";
import { useQuery } from "@tanstack/react-query";

const queryKey = ["shares", "public"] as const;

async function fetchPublicShares() {
  const result = await listPublicShares();
  if (result.error) {
    throw new Error(result.error);
  }
  return result.data ?? { items: [] };
}

export function usePublicShares() {
  const { data, isLoading, error } = useQuery({
    queryKey,
    queryFn: fetchPublicShares,
    staleTime: 60 * 1000,
  });

  return {
    items: data?.items ?? [],
    loading: isLoading,
    error: error instanceof Error ? error.message : null,
  };
}
