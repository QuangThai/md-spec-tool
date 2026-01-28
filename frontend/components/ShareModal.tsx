'use client';

import { useState } from 'react';
import { fetchAPI } from '@/lib/api';

interface ShareModalProps {
  specId: string;
  isOpen: boolean;
  onClose: () => void;
  onShareSuccess: () => void;
}

export function ShareModal({ specId, isOpen, onClose, onShareSuccess }: ShareModalProps) {
  const [shares, setShares] = useState<any[]>([]);
  const [sharedWithUserId, setSharedWithUserId] = useState('');
  const [permissionLevel, setPermissionLevel] = useState('view');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  if (!isOpen) return null;

  const loadShares = async () => {
    try {
      const data = await fetchAPI<any>(`/spec/${specId}/shares`, { method: 'GET' });
      if (data.data?.shares) {
        setShares(data.data.shares);
      }
    } catch (err) {
      setError('Failed to load shares');
    }
  };

  const handleShare = async () => {
    if (!sharedWithUserId.trim()) {
      setError('Please enter a user ID');
      return;
    }

    setLoading(true);
    setError('');

    try {
      const result = await fetchAPI(`/spec/${specId}/share`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          shared_with_user_id: sharedWithUserId,
          permission_level: permissionLevel,
        }),
      });

      if (result.error) {
        setError(result.error);
        return;
      }

      setSharedWithUserId('');
      setPermissionLevel('view');
      await loadShares();
      onShareSuccess();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to share spec');
    } finally {
      setLoading(false);
    }
  };

  const handleRemoveShare = async (userId: string) => {
    setLoading(true);
    try {
      const result = await fetchAPI(`/spec/${specId}/share/${userId}`, {
        method: 'DELETE',
      });
      if (result.error) {
        setError(result.error);
        return;
      }
      await loadShares();
    } catch (err) {
      setError('Failed to remove share');
    } finally {
      setLoading(false);
    }
  };

  const handleUpdatePermission = async (userId: string, newPermission: string) => {
    setLoading(true);
    try {
      const result = await fetchAPI(`/spec/${specId}/share/${userId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ permission_level: newPermission }),
      });
      if (result.error) {
        setError(result.error);
        return;
      }
      await loadShares();
    } catch (err) {
      setError('Failed to update permission');
    } finally {
      setLoading(false);
    }
  };

  const handleOpenChange = (open: boolean) => {
    if (open) {
      loadShares();
    }
    if (!open) {
      onClose();
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 w-96 max-h-96 overflow-y-auto">
        <h2 className="text-xl font-bold mb-4">Share Document</h2>

        {error && <div className="text-red-500 text-sm mb-4">{error}</div>}

        <div className="space-y-4 mb-6">
          <div>
            <label className="block text-sm font-medium mb-2">User ID</label>
            <input
              type="text"
              value={sharedWithUserId}
              onChange={(e) => setSharedWithUserId(e.target.value)}
              placeholder="Enter user ID"
              className="w-full px-3 py-2 border border-gray-300 rounded"
              disabled={loading}
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Permission</label>
            <select
              value={permissionLevel}
              onChange={(e) => setPermissionLevel(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded"
              disabled={loading}
            >
              <option value="view">View Only</option>
              <option value="edit">Can Edit</option>
            </select>
          </div>

          <button
            onClick={handleShare}
            disabled={loading}
            className="w-full bg-blue-500 text-white py-2 rounded hover:bg-blue-600 disabled:bg-gray-400"
          >
            {loading ? 'Sharing...' : 'Share'}
          </button>
        </div>

        <div className="border-t pt-4">
          <h3 className="font-semibold mb-3">Shared With</h3>
          {shares.length === 0 ? (
            <p className="text-gray-500 text-sm">Not shared with anyone yet</p>
          ) : (
            <div className="space-y-2">
              {shares.map((share) => (
                <div key={share.id} className="flex items-center justify-between bg-gray-100 p-2 rounded">
                  <div>
                    <p className="text-sm font-medium">{share.shared_with_username}</p>
                    <p className="text-xs text-gray-600">{share.permission_level}</p>
                  </div>
                  <select
                    value={share.permission_level}
                    onChange={(e) => handleUpdatePermission(share.shared_with_user_id, e.target.value)}
                    disabled={loading}
                    className="text-xs px-2 py-1 border rounded"
                  >
                    <option value="view">View</option>
                    <option value="edit">Edit</option>
                  </select>
                  <button
                    onClick={() => handleRemoveShare(share.shared_with_user_id)}
                    disabled={loading}
                    className="text-red-500 hover:text-red-700 text-sm"
                  >
                    Remove
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        <button
          onClick={() => handleOpenChange(false)}
          className="w-full mt-4 bg-gray-300 text-gray-700 py-2 rounded hover:bg-gray-400"
        >
          Close
        </button>
      </div>
    </div>
  );
}
