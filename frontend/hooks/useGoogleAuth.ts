"use client";

import { getGoogleLoginUrl, getGoogleAuthStatus, logoutGoogle } from '@/lib/oauthApi';
import { useCallback, useEffect, useRef, useState } from 'react';

type GoogleAuthState = {
  connected: boolean;
  loading: boolean;
  error: string | null;
  email?: string;
};

export function useGoogleAuth() {
  const [state, setState] = useState<GoogleAuthState>({
    connected: false,
    loading: true,
    error: null,
    email: undefined,
  });
  const popupRef = useRef<Window | null>(null);

  const refreshStatus = useCallback(async () => {
    setState((prev) => ({ ...prev, loading: true, error: null }));
    try {
      const result = await getGoogleAuthStatus();
      if (result.error) {
        setState({
          connected: false,
          loading: false,
          error: result.error,
          email: undefined,
        });
        return;
      }
      setState({
        connected: Boolean(result.data?.connected),
        loading: false,
        error: null,
        email: result.data?.email,
      });
    } catch (error) {
      setState({
        connected: false,
        loading: false,
        error: error instanceof Error ? error.message : 'Failed to check status',
        email: undefined,
      });
    }
  }, []);

  const login = useCallback(() => {
    const width = 520;
    const height = 640;
    const left = window.screenX + (window.outerWidth - width) / 2;
    const top = window.screenY + (window.outerHeight - height) / 2;

    popupRef.current = window.open(
      getGoogleLoginUrl(),
      'google-oauth',
      `width=${width},height=${height},left=${left},top=${top},resizable,scrollbars`
    );

    if (!popupRef.current) {
      setState({
        connected: false,
        loading: false,
        error: 'Failed to open popup. Popups may be blocked by your browser.',
        email: undefined,
      });
    }
  }, []);

  const logout = useCallback(async () => {
    setState((prev) => ({ ...prev, loading: true, error: null }));
    const result = await logoutGoogle();

    if (result.error) {
      setState((prev) => ({
        ...prev,
        loading: false,
        error: result.error || null,
      }));
      return;
    }

    setState({ connected: false, loading: false, error: null, email: undefined });
  }, []);

  useEffect(() => {
    void refreshStatus();
  }, [refreshStatus]);

  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      if (event.origin !== window.location.origin) return;
      if (event.data?.type === 'google-oauth-success') {
        void refreshStatus();
      }
      if (event.data?.type === 'google-oauth-error') {
        setState({
          connected: false,
          loading: false,
          error: event.data.error || 'Google OAuth failed',
        });
      }
    };

    window.addEventListener('message', handleMessage);
    return () => window.removeEventListener('message', handleMessage);
  }, [refreshStatus]);

  return {
    ...state,
    login,
    logout,
    refreshStatus,
  };
}
