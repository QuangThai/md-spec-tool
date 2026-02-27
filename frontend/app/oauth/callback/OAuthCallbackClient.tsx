"use client";

import { useMutation } from "@tanstack/react-query";
import { useSearchParams } from "next/navigation";
import { useEffect, useReducer, useRef } from "react";

type CallbackState = {
  message: string;
};

type CallbackAction =
  | { type: "set_message"; payload: string }
  | { type: "oauth_error"; payload: string }
  | { type: "oauth_success" };

function reducer(state: CallbackState, action: CallbackAction): CallbackState {
  switch (action.type) {
    case "set_message":
      return { ...state, message: action.payload };
    case "oauth_error":
      return { ...state, message: action.payload };
    case "oauth_success":
      return { ...state, message: "Connected. You can close this window." };
    default:
      return state;
  }
}

function OAuthCallbackContent() {
  const params = useSearchParams();
  const [state, dispatch] = useReducer(reducer, {
    message: "Connecting Google account…",
  });
  const hasTriggered = useRef(false);

  const exchangeMutation = useMutation({
    mutationFn: async ({
      code,
      state: stateParam,
    }: {
      code: string;
      state: string;
    }) => {
      const response = await fetch("/api/oauth/google/callback", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ code, state: stateParam }),
      });
      if (!response.ok) {
        const data = await response.json().catch(() => ({}));
        throw new Error(data.error || "OAuth exchange failed");
      }
    },
    onSuccess: () => {
      dispatch({ type: "oauth_success" });
      window.opener?.postMessage(
        { type: "google-oauth-success" },
        window.location.origin
      );
      window.close();
    },
    onError: (err: Error) => {
      const errorMessage = err.message || "OAuth exchange failed";
      dispatch({ type: "oauth_error", payload: errorMessage });
      window.opener?.postMessage(
        { type: "google-oauth-error", error: errorMessage },
        window.location.origin
      );
    },
  });

  useEffect(() => {
    const code = params.get("code");
    const stateParam = params.get("state");
    const error = params.get("error");

    if (error) {
      const errorMessage = `Google OAuth error: ${error}`;
      dispatch({ type: "oauth_error", payload: errorMessage });
      window.opener?.postMessage(
        { type: "google-oauth-error", error: errorMessage },
        window.location.origin
      );
      return;
    }

    if (!code || !stateParam) {
      const errorMessage = "Missing OAuth code or state";
      dispatch({ type: "oauth_error", payload: errorMessage });
      window.opener?.postMessage(
        { type: "google-oauth-error", error: errorMessage },
        window.location.origin
      );
      return;
    }

    if (!hasTriggered.current) {
      hasTriggered.current = true;
      exchangeMutation.mutate({ code, state: stateParam });
    }
    // Intentionally run only when params change; mutate is stable from useMutation
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params]);

  return (
    <div className="text-sm text-white/80">{state.message}</div>
  );
}

export default OAuthCallbackContent;
