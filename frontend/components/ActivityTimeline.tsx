'use client';

import { useEffect, useState } from 'react';
import { fetchAPI } from '@/lib/api';

interface ActivityLog {
  id: string;
  username: string;
  action: string;
  resource_type: string;
  details: any;
  created_at: string;
}

interface ActivityTimelineProps {
  specId: string;
}

export function ActivityTimeline({ specId }: ActivityTimelineProps) {
  const [logs, setLogs] = useState<ActivityLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);

  useEffect(() => {
    loadActivity();
  }, [specId]);

  const loadActivity = async (loadOffset = 0) => {
    setLoading(true);
    try {
      const data = await fetchAPI<any>(`/spec/${specId}/activity?limit=20&offset=${loadOffset}`, {
        method: 'GET',
      });
      if (data.data) {
        setLogs(data.data.logs || []);
        setTotal(data.data.total || 0);
      }
    } catch (err) {
      console.error('Failed to load activity:', err);
    } finally {
      setLoading(false);
    }
  };

  const getActionIcon = (action: string) => {
    switch (action) {
      case 'created':
        return '✨';
      case 'updated':
        return '✏️';
      case 'shared':
        return '📤';
      case 'commented':
        return '💬';
      case 'deleted':
        return '🗑️';
      default:
        return '📌';
    }
  };

  const getActionLabel = (action: string) => {
    switch (action) {
      case 'created':
        return 'Created';
      case 'updated':
        return 'Updated';
      case 'shared':
        return 'Shared';
      case 'commented':
        return 'Commented';
      case 'deleted':
        return 'Deleted';
      default:
        return 'Activity';
    }
  };

  return (
    <div className="mt-6 border-t pt-6">
      <h3 className="text-lg font-bold mb-4">Activity Timeline</h3>

      {loading && <p className="text-gray-600">Loading activity...</p>}

      {logs.length === 0 && !loading && (
        <p className="text-gray-500 text-sm">No activity yet</p>
      )}

      <div className="space-y-4">
        {logs.map((log, index) => (
          <div key={log.id} className="flex gap-4">
            {/* Timeline line */}
            <div className="flex flex-col items-center">
              <div className="w-10 h-10 rounded-full bg-blue-100 flex items-center justify-center text-lg">
                {getActionIcon(log.action)}
              </div>
              {index < logs.length - 1 && (
                <div className="w-1 h-8 bg-gray-300 my-2"></div>
              )}
            </div>

            {/* Activity details */}
            <div className="flex-1 pt-2">
              <div className="flex justify-between items-start">
                <div>
                  <p className="font-semibold text-sm">
                    {getActionLabel(log.action)} by {log.username}
                  </p>
                  {log.details && (
                    <p className="text-sm text-gray-600 mt-1">
                      {log.details.field && (
                        <>
                          {log.details.field}
                          {log.details.old_value && ` changed from "${log.details.old_value}"`}
                          {log.details.new_value && ` to "${log.details.new_value}"`}
                        </>
                      )}
                    </p>
                  )}
                </div>
                <span className="text-xs text-gray-500">
                  {new Date(log.created_at).toLocaleDateString()} at{' '}
                  {new Date(log.created_at).toLocaleTimeString([], {
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </span>
              </div>
            </div>
          </div>
        ))}
      </div>

      {logs.length < total && (
        <button
          onClick={() => {
            const newOffset = offset + 20;
            setOffset(newOffset);
            loadActivity(newOffset);
          }}
          disabled={loading}
          className="mt-4 text-sm text-blue-500 hover:text-blue-700 disabled:text-gray-400"
        >
          Load more activity
        </button>
      )}
    </div>
  );
}
