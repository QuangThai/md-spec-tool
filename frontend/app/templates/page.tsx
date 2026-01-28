'use client';

import { useEffect, useState } from 'react';
import { fetchAPI } from '@/lib/api';
import { useAuthStore } from '@/lib/store';
import { useRouter } from 'next/navigation';

interface Template {
  id: string;
  name: string;
  description?: string;
}

export default function TemplatesPage() {
  const [templates, setTemplates] = useState<Template[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedTemplate, setSelectedTemplate] = useState<Template | null>(null);
  const { isLoggedIn, token } = useAuthStore();
  const router = useRouter();

  useEffect(() => {
    if (!isLoggedIn) {
      router.push('/auth/login');
      return;
    }

    const loadTemplates = async () => {
      const result = await fetchAPI<{ templates: Template[] }>('/templates', {
        headers: { Authorization: `Bearer ${token}` },
      });

      if (result.error) {
        setError(result.error);
      } else if (result.data?.templates) {
        setTemplates(result.data.templates);
      }
      setLoading(false);
    };

    loadTemplates();
  }, [isLoggedIn, router, token]);

  if (loading) return <div className="text-center mt-8">Loading templates...</div>;
  if (error) return <div className="text-red-500 text-center mt-8">{error}</div>;

  return (
    <div className="space-y-6">
      <div className="surface p-6">
        <p className="pill">Templates</p>
        <h2 className="mt-3 text-3xl font-semibold text-slate-900">Markdown Templates</h2>
        <p className="mt-2 text-sm text-slate-500">Reusable formats that keep every spec consistent.</p>
      </div>

      {templates.length === 0 ? (
        <div className="surface p-10 text-center">
          <p className="text-slate-500">No templates available yet.</p>
          <p className="mt-2 text-sm text-slate-400">Templates will appear here once created by administrators.</p>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {templates.map((tmpl) => (
            <button
              key={tmpl.id}
              className="surface flex h-full flex-col gap-3 p-5 text-left transition hover:-translate-y-1"
              onClick={() => setSelectedTemplate(tmpl)}
            >
              <div>
                <h3 className="text-lg font-semibold text-slate-900">{tmpl.name}</h3>
                <p className="mt-2 text-sm text-slate-600">{tmpl.description || 'No description available'}</p>
              </div>
              <span className="text-sm font-semibold text-slate-700">View details →</span>
            </button>
          ))}
        </div>
      )}

      {selectedTemplate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4">
          <div className="surface w-full max-w-md p-6">
            <h3 className="text-2xl font-semibold text-slate-900">{selectedTemplate.name}</h3>
            <p className="mt-3 text-sm text-slate-600">
              {selectedTemplate.description || 'No description'}
            </p>
            <div className="mt-6 flex gap-2">
              <button onClick={() => setSelectedTemplate(null)} className="btn-secondary flex-1">
                Close
              </button>
              <button
                onClick={() => {
                  setSelectedTemplate(null);
                }}
                className="btn-primary flex-1"
              >
                Use Template
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
