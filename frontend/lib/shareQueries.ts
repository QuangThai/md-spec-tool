import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { queryKeys, unwrap } from './queryUtils';
import {
  CommentResponse,
  createComment,
  createShare,
  getShare,
  listComments,
  listPublicShares,
  SharePayload,
  ShareResponse,
  ShareSummary,
  updateComment,
  updateShare,
} from './shareApi';

export function useShareQuery(key: string, enabled = true) {
  return useQuery({
    queryKey: queryKeys.share.detail(key),
    queryFn: async () => unwrap(await getShare(key)),
    enabled: enabled && Boolean(key),
    staleTime: 5 * 60 * 1000,
  });
}

export function usePublicSharesQuery() {
  return useQuery({
    queryKey: queryKeys.share.publicList,
    queryFn: async () => {
      const result = await listPublicShares();
      return unwrap(result).items;
    },
    staleTime: 2 * 60 * 1000,
  });
}

export function useCommentsQuery(key: string, enabled = true) {
  return useQuery({
    queryKey: queryKeys.share.comments(key),
    queryFn: async () => {
      const result = await listComments(key);
      return unwrap(result).items;
    },
    enabled: enabled && Boolean(key),
    staleTime: 30 * 1000,
  });
}

export function useCreateShareMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: SharePayload): Promise<ShareResponse> => {
      return unwrap(await createShare(payload));
    },
    onSuccess: (data) => {
      queryClient.setQueryData(queryKeys.share.detail(data.slug), data);
      queryClient.invalidateQueries({ queryKey: queryKeys.share.publicList });
    },
  });
}

export function useUpdateShareMutation(key: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: { is_public?: boolean; allow_comments?: boolean }): Promise<ShareResponse> => {
      return unwrap(await updateShare(key, payload));
    },
    onSuccess: (data) => {
      queryClient.setQueryData(queryKeys.share.detail(key), data);
      queryClient.invalidateQueries({ queryKey: queryKeys.share.publicList });
    },
  });
}

export function useCreateCommentMutation(shareKey: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: { author?: string; message: string }): Promise<CommentResponse> => {
      return unwrap(await createComment(shareKey, payload));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.share.comments(shareKey) });
    },
  });
}

export function useUpdateCommentMutation(shareKey: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: { commentId: string; resolved: boolean }): Promise<CommentResponse> => {
      return unwrap(await updateComment(shareKey, payload.commentId, payload.resolved));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.share.comments(shareKey) });
    },
  });
}
