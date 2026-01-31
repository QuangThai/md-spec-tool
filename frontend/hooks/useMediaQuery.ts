"use client";

import { useSyncExternalStore } from "react";

/**
 * SSR-safe useMediaQuery hook (React 18+).
 *
 * Best practices:
 * - Uses useSyncExternalStore to avoid hydration mismatch and tearing.
 * - getServerSnapshot returns initialValue so server and first client paint match.
 * - Pass initialValue when you need a specific SSR/default (e.g. true for "desktop" to avoid sidebar flash).
 *
 * @param query - Media query string (e.g. "(min-width: 1024px)")
 * @param options.initialValue - Value used during SSR and before first client update. Default false.
 * @returns whether the media query matches
 */
export function useMediaQuery(
  query: string,
  options?: { initialValue?: boolean }
): boolean {
  const initial = options?.initialValue ?? false;

  const subscribe = (onStoreChange: () => void) => {
    if (typeof window === "undefined") return () => {};
    const mq = window.matchMedia(query);
    mq.addEventListener("change", onStoreChange);
    return () => mq.removeEventListener("change", onStoreChange);
  };

  const getSnapshot = (): boolean => {
    if (typeof window === "undefined") return initial;
    return window.matchMedia(query).matches;
  };

  const getServerSnapshot = (): boolean => initial;

  return useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
}
