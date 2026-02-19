export type TelemetryStatus = "success" | "error" | "cancel";

export interface TelemetryEvent {
  event_name: string;
  event_time: string;
  session_id: string;
  status: TelemetryStatus;
  input_source?: "paste" | "xlsx" | "gsheet" | "tsv" | "share";
  template_type?: "spec" | "table";
  duration_ms?: number;
  error_code?: string;
  warning_count?: number;
  confidence_score?: number;
  needs_review?: boolean;
  total_rows?: number;
  [key: string]: unknown;
}

const SESSION_STORAGE_KEY = "mdflow_session_id";
const TELEMETRY_ENDPOINT =
  (process.env.NEXT_PUBLIC_API_URL || "").replace(/\/$/, "") +
  "/api/telemetry/events";

function generateSessionID(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
}

export function getOrCreateSessionID(): string {
  if (typeof window === "undefined") {
    return "server";
  }

  // Use localStorage so all tabs share the same session ID for quota tracking
  const existing = window.localStorage.getItem(SESSION_STORAGE_KEY);
  if (existing) {
    return existing;
  }

  const next = generateSessionID();
  window.localStorage.setItem(SESSION_STORAGE_KEY, next);
  return next;
}

export function emitTelemetryEvent(
  eventName: string,
  data: Omit<TelemetryEvent, "event_name" | "event_time" | "session_id" | "status"> & {
    status?: TelemetryStatus;
  }
): void {
  if (typeof window === "undefined") {
    return;
  }

  const event: TelemetryEvent = {
    event_name: eventName,
    event_time: new Date().toISOString(),
    session_id: getOrCreateSessionID(),
    status: data.status ?? "success",
    ...data,
  };

  window.dispatchEvent(
    new CustomEvent("mdflow-telemetry", {
      detail: event,
    })
  );

  const scoped = window as Window & {
    __MDFLOW_TELEMETRY__?: TelemetryEvent[];
  };
  if (!scoped.__MDFLOW_TELEMETRY__) {
    scoped.__MDFLOW_TELEMETRY__ = [];
  }
  scoped.__MDFLOW_TELEMETRY__.push(event);

  if (process.env.NODE_ENV !== "production") {
    // Keep dev telemetry visible for quick validation during MVP rollout.
    console.info("[mdflow-telemetry]", event);
  }

  // Fire-and-forget transport for MVP dashboard ingestion.
  const payload = JSON.stringify({ events: [event] });
  if (navigator.sendBeacon) {
    const blob = new Blob([payload], { type: "application/json" });
    navigator.sendBeacon(TELEMETRY_ENDPOINT, blob);
    return;
  }

  void fetch(TELEMETRY_ENDPOINT, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: payload,
    keepalive: true,
  }).catch(() => {
    // Telemetry must never block user interactions.
  });
}
