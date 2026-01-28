'use client';

import { useEffect, useState } from 'react';
import { fetchAPI } from '@/lib/api';
import { CommentThread } from './CommentThread';

interface Comment {
  id: string;
  spec_id: string;
  user_id: string;
  username: string;
  content: string;
  created_at: string;
  updated_at: string;
}

interface CommentsSectionProps {
  specId: string;
  currentUserId: string;
}

export function CommentsSection({ specId, currentUserId }: CommentsSectionProps) {
  const [comments, setComments] = useState<Comment[]>([]);
  const [newComment, setNewComment] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editContent, setEditContent] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    loadComments();
  }, [specId]);

  const loadComments = async () => {
    try {
      const data = await fetchAPI<any>(`/spec/${specId}/comments`, { method: 'GET' });
      if (data.data?.comments) {
        setComments(data.data.comments);
      }
    } catch (err) {
      setError('Failed to load comments');
    }
  };

  const handleAddComment = async () => {
    if (!newComment.trim()) return;

    setLoading(true);
    setError('');

    try {
      const result = await fetchAPI(`/spec/${specId}/comments`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content: newComment }),
      });

      if (result.error) {
        setError(result.error);
        return;
      }

      setNewComment('');
      await loadComments();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add comment');
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateComment = async (commentId: string) => {
    if (!editContent.trim()) return;

    setLoading(true);
    try {
      const result = await fetchAPI(`/spec/${specId}/comments/${commentId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content: editContent }),
      });

      if (result.error) {
        setError(result.error);
        return;
      }

      setEditingId(null);
      setEditContent('');
      await loadComments();
    } catch (err) {
      setError('Failed to update comment');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteComment = async (commentId: string) => {
    setLoading(true);
    try {
      const result = await fetchAPI(`/spec/${specId}/comments/${commentId}`, {
        method: 'DELETE',
      });
      if (result.error) {
        setError(result.error);
        return;
      }
      await loadComments();
    } catch (err) {
      setError('Failed to delete comment');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mt-6 border-t pt-6">
      <h3 className="text-lg font-bold mb-4">Comments ({comments.length})</h3>

      {error && <div className="text-red-500 text-sm mb-4">{error}</div>}

      {/* Add Comment */}
      <div className="mb-6">
        <textarea
          value={newComment}
          onChange={(e) => setNewComment(e.target.value)}
          placeholder="Add a comment..."
          className="w-full px-3 py-2 border border-gray-300 rounded mb-2"
          rows={3}
          disabled={loading}
        />
        <button
          onClick={handleAddComment}
          disabled={loading || !newComment.trim()}
          className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 disabled:bg-gray-400"
        >
          {loading ? 'Adding...' : 'Add Comment'}
        </button>
      </div>

      {/* Comments List with Threading */}
      <div className="space-y-4">
        {comments.length === 0 ? (
          <p className="text-gray-500 text-sm">No comments yet. Be the first to comment!</p>
        ) : (
          comments.map((comment) => (
            <CommentThread
              key={comment.id}
              comment={comment}
              specId={specId}
              currentUserId={currentUserId}
              onDelete={loadComments}
              onReplyAdded={loadComments}
            />
          ))
        )}
      </div>
    </div>
  );
}
