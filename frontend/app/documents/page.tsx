'use client';

import { useEffect, useState } from 'react';
import { fetchAPI } from '@/lib/api';
import Link from 'next/link';
import { useAuthStore } from '@/lib/store';
import { useRouter } from 'next/navigation';

interface Spec {
  id: string;
  title: string;
  content: string;
  version: number;
  created_at: string;
  updated_at: string;
}

export default function DocumentsPage() {
  const [docs, setDocs] = useState<Spec[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [searching, setSearching] = useState(false);
  const { isLoggedIn, token } = useAuthStore();
  const router = useRouter();

  useEffect(() => {
    if (!isLoggedIn) {
      router.push('/auth/login');
      return;
    }
    loadDocs();
  }, [isLoggedIn, router]);

  const loadDocs = async () => {
    setLoading(true);
    setError('');
    const result = await fetchAPI<Spec[] | { specs: Spec[] }>('/spec', {
      headers: { Authorization: `Bearer ${token}` },
    });

    if (result.error) {
      setError(result.error);
    } else if (result.data) {
      const specList = Array.isArray(result.data) ? result.data : result.data.specs || [];
      setDocs(specList);
    }
    setLoading(false);
  };

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!searchQuery.trim()) {
      loadDocs();
      return;
    }

    setSearching(true);
    const result = await fetchAPI<{ specs: Spec[] }>('/spec/search', {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify({ query: searchQuery }),
    });

    if (result.data?.specs) {
      setDocs(result.data.specs);
    }
    setSearching(false);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this document?')) return;

    const result = await fetchAPI(`/spec/${id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${token}` },
    });

    if (result.error) {
      alert('Failed to delete: ' + result.error);
    } else {
      setDocs(docs.filter(d => d.id !== id));
    }
  };

  const handleDownload = async (id: string, title: string) => {
    const response = await fetch(
      `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/spec/${id}/download`,
      {
        headers: { Authorization: `Bearer ${token}` },
      }
    );

    if (response.ok) {
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${title}.md`;
      a.click();
      window.URL.revokeObjectURL(url);
    }
  };

  if (!isLoggedIn) {
    return <p>Redirecting to login...</p>;
  }

  if (loading) return <p className="text-center mt-8">Loading documents...</p>;

  return (
    <div className="space-y-6">
      <div className="surface p-6">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <p className="pill">Documents</p>
            <h2 className="mt-3 text-3xl font-semibold text-slate-900">My Library</h2>
            <p className="mt-2 text-sm text-slate-500">Track every version of your specs in one place.</p>
          </div>
          <Link href="/converter" className="btn-primary">
            + New Document
          </Link>
        </div>
      </div>

      {error && (
        <div className="rounded-2xl border border-rose-200 bg-rose-50 px-5 py-4 text-sm text-rose-700">
          {error}
        </div>
      )}

      <form onSubmit={handleSearch} className="surface flex flex-col gap-3 p-5 sm:flex-row sm:items-center">
        <input
          type="text"
          placeholder="Search documents by title..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="input flex-1"
        />
        <div className="flex gap-2">
          <button type="submit" disabled={searching} className="btn-primary">
            {searching ? 'Searching...' : 'Search'}
          </button>
          {searchQuery && (
            <button
              type="button"
              onClick={() => {
                setSearchQuery('');
                loadDocs();
              }}
              className="btn-secondary"
            >
              Clear
            </button>
          )}
        </div>
      </form>

      {docs.length === 0 ? (
        <div className="surface p-10 text-center">
          <p className="text-slate-500">
            {searchQuery ? 'No documents found matching your search.' : 'No documents yet.'}
          </p>
          <Link href="/converter" className="btn-secondary mt-5">
            Create your first document
          </Link>
        </div>
      ) : (
        <div className="space-y-4">
          {docs.map((doc) => (
            <div key={doc.id} className="surface p-5">
              <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                <div className="flex-1">
                  <Link href={`/documents/${doc.id}`}>
                    <h3 className="text-lg font-semibold text-slate-900 hover:text-slate-700">
                      {doc.title}
                    </h3>
                  </Link>
                  <div className="mt-2 flex flex-wrap gap-4 text-xs text-slate-500">
                    <span>Version {doc.version}</span>
                    <span>Updated {new Date(doc.updated_at).toLocaleDateString()}</span>
                    <span>Created {new Date(doc.created_at).toLocaleDateString()}</span>
                  </div>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Link href={`/documents/${doc.id}`} className="btn-secondary">
                    View
                  </Link>
                  <button onClick={() => handleDownload(doc.id, doc.title)} className="btn-secondary">
                    Download
                  </button>
                  <button onClick={() => handleDelete(doc.id)} className="btn-ghost text-rose-600 hover:text-rose-700">
                    Delete
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
