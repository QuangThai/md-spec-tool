'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { fetchAPI } from '@/lib/api';
import { useAuthStore } from '@/lib/store';
import Link from 'next/link';

interface Spec {
  id: string;
  title: string;
  content: string;
  version: number;
  created_at: string;
  updated_at: string;
}

interface Version {
  version: number;
  updated_at: string;
}

export default function DocumentDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { token, isLoggedIn } = useAuthStore();
  
  const [spec, setSpec] = useState<Spec | null>(null);
  const [versions, setVersions] = useState<Version[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const docId = params.id as string;

  useEffect(() => {
    if (!isLoggedIn) {
      router.push('/auth/login');
      return;
    }
    loadSpec();
  }, [isLoggedIn, router, docId]);

  const loadSpec = async () => {
    setLoading(true);
    const result = await fetchAPI<Spec>(`/spec/${docId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });

    if (result.error) {
      setError(result.error);
    } else if (result.data) {
      setSpec(result.data);
      loadVersions();
    }
    setLoading(false);
  };

  const loadVersions = async () => {
    const result = await fetchAPI<{ versions: Version[] }>(`/spec/${docId}/versions`, {
      headers: { Authorization: `Bearer ${token}` },
    });

    if (result.data?.versions) {
      setVersions(result.data.versions);
    }
  };

  const handleDownload = async () => {
    if (!spec) return;
    const response = await fetch(
      `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/spec/${docId}/download`,
      {
        headers: { Authorization: `Bearer ${token}` },
      }
    );

    if (response.ok) {
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${spec.title}.md`;
      a.click();
      window.URL.revokeObjectURL(url);
    }
  };

  if (!isLoggedIn) {
    return <p>Redirecting to login...</p>;
  }

  if (loading) return <div className="text-center mt-8">Loading document...</div>;

  if (error) {
    return (
      <div className="space-y-4">
        <div className="rounded-2xl border border-rose-200 bg-rose-50 px-5 py-4 text-sm text-rose-700">
          {error}
        </div>
        <Link href="/documents" className="btn-ghost">
          {'<- Back to Documents'}
        </Link>
      </div>
    );
  }

  if (!spec) {
    return (
      <div className="space-y-4">
        <p className="text-slate-600">Document not found.</p>
        <Link href="/documents" className="btn-ghost">
          {'<- Back to Documents'}
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <Link href="/documents" className="btn-ghost">
        {'<- Back to Documents'}
      </Link>

      <div className="surface p-6">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <h1 className="text-3xl font-semibold text-slate-900">{spec.title}</h1>
            <div className="mt-2 flex flex-wrap gap-4 text-xs text-slate-500">
              <span>Version {spec.version}</span>
              <span>Created {new Date(spec.created_at).toLocaleDateString()}</span>
              <span>Updated {new Date(spec.updated_at).toLocaleDateString()}</span>
            </div>
          </div>
          <button onClick={handleDownload} className="btn-secondary">
            Download Markdown
          </button>
        </div>

        {versions.length > 0 && (
          <div className="mt-6 rounded-xl border border-slate-200/70 bg-slate-50/80 p-4">
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">Version History</p>
            <div className="mt-3 flex flex-wrap gap-2">
              {versions.map((v) => (
                <button
                  key={v.version}
                  disabled={v.version === spec.version}
                  className={`rounded-full px-3 py-1 text-xs font-semibold transition ${
                    v.version === spec.version
                      ? 'bg-slate-900 text-white'
                      : 'border border-slate-200/70 bg-white text-slate-600 hover:bg-slate-100'
                  }`}
                >
                  v{v.version}
                </button>
              ))}
            </div>
          </div>
        )}
      </div>

      <div className="surface p-6">
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-semibold text-slate-900">Content</h2>
          <button
            onClick={() => {
              navigator.clipboard.writeText(spec.content);
              alert('Copied to clipboard!');
            }}
            className="btn-ghost"
          >
            Copy
          </button>
        </div>
        <div className="mt-4 max-h-96 overflow-y-auto rounded-xl border border-slate-200/70 bg-slate-50 p-4">
          <pre className="text-sm whitespace-pre-wrap font-mono text-slate-700">{spec.content}</pre>
        </div>
      </div>
    </div>
  );
}
