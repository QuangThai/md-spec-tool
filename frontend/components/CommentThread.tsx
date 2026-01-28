'use client';

import { useState } from 'react';
import { fetchAPI } from '@/lib/api';

interface CommentThreadProps {
  comment: any;
  specId: string;
  currentUserId: string;
  onDelete: () => void;
  onReplyAdded: () => void;
}

export function CommentThread({ comment, specId, currentUserId, onDelete, onReplyAdded }: CommentThreadProps) {
  const [showReplyForm, setShowReplyForm] = useState(false);
  const [replyContent, setReplyContent] = useState('');
  const [loading, setLoading] = useState(false);

  const handleAddReply = async () => {
    if (!replyContent.trim()) return;

    setLoading(true);
    try {
      await fetchAPI(`/spec/${specId}/comments/${comment.id}/reply`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content: replyContent }),
      });

      setReplyContent('');
      setShowReplyForm(false);
      onReplyAdded();
    } catch (err) {
      console.error('Failed to add reply:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    setLoading(true);
    try {
      await fetchAPI(`/spec/${specId}/comments/${comment.id}`, {
        method: 'DELETE',
      });
      onDelete();
    } catch (err) {
      console.error('Failed to delete comment:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-gray-50 p-4 rounded border border-gray-200 mb-4">
      {/* Main Comment */}
      <div>
        <div className="flex justify-between items-start mb-2">
          <div>
            <p className="font-semibold text-sm">{comment.username}</p>
            <p className="text-xs text-gray-600">
              {new Date(comment.created_at).toLocaleDateString()} at{' '}
              {new Date(comment.created_at).toLocaleTimeString()}
            </p>
          </div>
          {comment.user_id === currentUserId && (
            <button
              onClick={handleDelete}
              disabled={loading}
              className="text-red-500 hover:text-red-700 text-xs"
            >
              Delete
            </button>
          )}
        </div>

        <p className="text-sm text-gray-800 mb-3">{comment.content}</p>

        <button
          onClick={() => setShowReplyForm(!showReplyForm)}
          className="text-blue-500 hover:text-blue-700 text-xs"
          disabled={loading}
        >
          {showReplyForm ? 'Cancel' : 'Reply'}
        </button>
      </div>

      {/* Reply Form */}
      {showReplyForm && (
        <div className="mt-3 ml-4 p-3 bg-white rounded border border-gray-300">
          <textarea
            value={replyContent}
            onChange={(e) => setReplyContent(e.target.value)}
            placeholder="Add a reply..."
            className="w-full px-2 py-2 border border-gray-300 rounded text-sm mb-2"
            rows={2}
            disabled={loading}
          />
          <button
            onClick={handleAddReply}
            disabled={loading || !replyContent.trim()}
            className="text-sm px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:bg-gray-400"
          >
            {loading ? 'Adding...' : 'Add Reply'}
          </button>
        </div>
      )}

      {/* Replies */}
      {comment.replies && comment.replies.length > 0 && (
        <div className="mt-4 ml-4 space-y-3 border-l-2 border-gray-300 pl-4">
          {comment.replies.map((reply: any) => (
            <div key={reply.id} className="bg-white p-3 rounded border border-gray-200">
              <div className="flex justify-between items-start mb-2">
                <div>
                  <p className="font-semibold text-sm">{reply.username}</p>
                  <p className="text-xs text-gray-600">
                    {new Date(reply.created_at).toLocaleDateString()} at{' '}
                    {new Date(reply.created_at).toLocaleTimeString()}
                  </p>
                </div>
                {reply.user_id === currentUserId && (
                  <button
                    onClick={async () => {
                      try {
                        await fetchAPI(`/spec/${specId}/comments/${reply.id}`, {
                          method: 'DELETE',
                        });
                        onReplyAdded();
                      } catch (err) {
                        console.error('Failed to delete reply:', err);
                      }
                    }}
                    className="text-red-500 hover:text-red-700 text-xs"
                  >
                    Delete
                  </button>
                )}
              </div>
              <p className="text-sm text-gray-800">{reply.content}</p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
