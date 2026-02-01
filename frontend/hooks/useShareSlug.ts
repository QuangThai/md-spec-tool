import {
  CommentResponse,
  createComment,
  getShare,
  listComments,
  ShareResponse,
  updateComment,
  updateShare,
} from "@/lib/shareApi";
import { ApiResult } from "@/lib/types";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

const queryKeys = {
  share: (key: string) => ["share", key] as const,
  comments: (key: string) => ["share", key, "comments"] as const,
};

function unwrap<T>(result: ApiResult<T>): T {
  if (result.error) {
    throw new Error(result.error);
  }
  if (!result.data) {
    throw new Error("No data returned");
  }
  return result.data;
}

async function fetchShare(key: string) {
  return unwrap(await getShare(key));
}

async function fetchComments(key: string) {
  return unwrap(await listComments(key));
}

export function useShareSlug(shareKey: string) {
  const queryClient = useQueryClient();

  const shareQuery = useQuery({
    queryKey: queryKeys.share(shareKey),
    queryFn: async () => fetchShare(shareKey),
    enabled: Boolean(shareKey),
    staleTime: 60 * 1000,
  });

  const commentsQuery = useQuery({
    queryKey: queryKeys.comments(shareKey),
    queryFn: async () => fetchComments(shareKey),
    enabled: Boolean(shareKey),
    staleTime: 30 * 1000,
  });

  const createCommentMutation = useMutation({
    mutationFn: async (payload: { author?: string; message: string }) => {
      if (!shareKey) {
        throw new Error("Missing share key");
      }
      return unwrap(await createComment(shareKey, payload));
    },
    onSuccess: (data) => {
      queryClient.setQueryData<{ items: CommentResponse[] }>(
        queryKeys.comments(shareKey),
        (prev) => ({ items: [...(prev?.items ?? []), data] })
      );
    },
  });

  const updateCommentMutation = useMutation({
    mutationFn: async (payload: { commentId: string; resolved: boolean }) => {
      if (!shareKey) {
        throw new Error("Missing share key");
      }
      return unwrap(await updateComment(shareKey, payload.commentId, payload.resolved));
    },
    onSuccess: (data) => {
      queryClient.setQueryData<{ items: CommentResponse[] }>(
        queryKeys.comments(shareKey),
        (prev) => ({
          items: (prev?.items ?? []).map((item) => (item.id === data.id ? data : item)),
        })
      );
    },
  });

  const updateShareMutation = useMutation({
    mutationFn: async (payload: { is_public?: boolean; allow_comments?: boolean }) => {
      if (!shareKey) {
        throw new Error("Missing share key");
      }
      return unwrap(await updateShare(shareKey, payload));
    },
    onSuccess: (data) => {
      queryClient.setQueryData<ShareResponse>(queryKeys.share(shareKey), data);
    },
  });

  return {
    share: shareQuery.data ?? null,
    shareLoading: shareQuery.isLoading,
    shareError: shareQuery.error instanceof Error ? shareQuery.error.message : null,
    comments: commentsQuery.data ?? { items: [] },
    commentsLoading: commentsQuery.isLoading,
    commentsError: commentsQuery.error instanceof Error ? commentsQuery.error.message : null,
    createComment: createCommentMutation.mutateAsync,
    creatingComment: createCommentMutation.isPending,
    updateComment: updateCommentMutation.mutateAsync,
    updatingComment: updateCommentMutation.isPending,
    updateShare: updateShareMutation.mutateAsync,
    updatingShare: updateShareMutation.isPending,
  };
}
