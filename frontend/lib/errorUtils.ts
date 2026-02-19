export type ErrorKind =
  | "rate_limit"
  | "timeout"
  | "network"
  | "unauthorized"
  | "forbidden"
  | "not_found"
  | "validation"
  | "request_too_large"
  | "server"
  | "service_unavailable"
  | "unknown";

/**
 * API Error Codes (from backend ERROR-CONTRACT)
 * Used by backend in error responses
 */
export enum ErrorCode {
  BadRequest = "BAD_REQUEST",
  Unauthorized = "UNAUTHORIZED",
  Forbidden = "FORBIDDEN",
  NotFound = "NOT_FOUND",
  RequestTooLarge = "REQUEST_TOO_LARGE",
  RateLimitExceeded = "RATE_LIMIT_EXCEEDED",
  InternalError = "INTERNAL_ERROR",
  ServiceUnavailable = "SERVICE_UNAVAILABLE",
}

export interface UserFacingError {
  kind: ErrorKind;
  title: string;
  message: string;
  retryable: boolean;
  requestId?: string;
  code?: ErrorCode;
}

/**
 * Extract request ID from error message or error object
 */
function extractRequestId(raw: string | { requestId?: string }): string | undefined {
  // Handle object with requestId field
  if (typeof raw === "object" && raw !== null && "requestId" in raw) {
    return (raw as any).requestId;
  }

  // Extract from string
  if (typeof raw === "string") {
    const match = raw.match(/request[_ -]?id[:=]\s*([a-zA-Z0-9._:-]+)/i);
    return match?.[1];
  }

  return undefined;
}

/**
 * Map backend error codes to user-facing errors
 * Follows backend error taxonomy: docs/ERROR-CONTRACT.md
 */
export function mapErrorCodeToUserFacing(code: string | undefined): Partial<UserFacingError> {
  switch (code) {
    case ErrorCode.RateLimitExceeded:
      return {
        kind: "rate_limit",
        title: "Rate Limited",
        message: "Too many requests. Please wait and retry.",
        retryable: true,
        code: ErrorCode.RateLimitExceeded,
      };

    case ErrorCode.Unauthorized:
      return {
        kind: "unauthorized",
        title: "Unauthorized",
        message: "Authentication required or session expired.",
        retryable: false,
        code: ErrorCode.Unauthorized,
      };

    case ErrorCode.Forbidden:
      return {
        kind: "forbidden",
        title: "Access Denied",
        message: "You don't have permission for this action.",
        retryable: false,
        code: ErrorCode.Forbidden,
      };

    case ErrorCode.NotFound:
      return {
        kind: "not_found",
        title: "Not Found",
        message: "The requested resource was not found.",
        retryable: false,
        code: ErrorCode.NotFound,
      };

    case ErrorCode.RequestTooLarge:
      return {
        kind: "request_too_large",
        title: "File Too Large",
        message: "The file exceeds the maximum allowed size.",
        retryable: false,
        code: ErrorCode.RequestTooLarge,
      };

    case ErrorCode.BadRequest:
      return {
        kind: "validation",
        title: "Invalid Input",
        message: "Please check your input and try again.",
        retryable: false,
        code: ErrorCode.BadRequest,
      };

    case ErrorCode.ServiceUnavailable:
      return {
        kind: "service_unavailable",
        title: "Service Unavailable",
        message: "Server is temporarily unavailable. Please retry shortly.",
        retryable: true,
        code: ErrorCode.ServiceUnavailable,
      };

    case ErrorCode.InternalError:
      return {
        kind: "server",
        title: "Server Error",
        message: "An unexpected error occurred. Please try again.",
        retryable: true,
        code: ErrorCode.InternalError,
      };

    default:
      return {};
  }
}

/**
 * Map generic error strings to user-facing errors (fallback for network errors, timeouts, etc.)
 */
export function mapErrorStringToUserFacing(rawError: string): UserFacingError {
  const normalized = (rawError || "").toLowerCase();
  const requestId = extractRequestId(rawError);

  // Rate limit (pattern match as fallback)
  if (
    normalized.includes("rate_limit_exceeded") ||
    normalized.includes("rate limit exceeded") ||
    normalized.includes("429")
  ) {
    return {
      kind: "rate_limit",
      title: "Rate Limited",
      message: "Too many requests. Please wait and retry.",
      retryable: true,
      requestId,
    };
  }

  // Timeout
  if (normalized.includes("timeout") || normalized.includes("request timeout")) {
    return {
      kind: "timeout",
      title: "Request Timed Out",
      message: "The request took too long. Please retry.",
      retryable: true,
      requestId,
    };
  }

  // Network error
  if (normalized.includes("network_error") || normalized.includes("network")) {
    return {
      kind: "network",
      title: "Network Error",
      message: "Cannot reach the server. Check your connection and retry.",
      retryable: true,
      requestId,
    };
  }

  // 401/Unauthorized (pattern match)
  if (normalized.includes("unauthorized") || normalized.includes("401")) {
    return {
      kind: "unauthorized",
      title: "Unauthorized",
      message: "Authentication required or session expired.",
      retryable: false,
      requestId,
    };
  }

  // 403/Forbidden (pattern match)
  if (normalized.includes("forbidden") || normalized.includes("403")) {
    return {
      kind: "forbidden",
      title: "Access Denied",
      message: "You don't have permission for this action.",
      retryable: false,
      requestId,
    };
  }

  // 404/Not Found (pattern match)
  if (
    normalized.includes("not_found") ||
    normalized.includes("not found") ||
    normalized.includes("404")
  ) {
    return {
      kind: "not_found",
      title: "Not Found",
      message: "The requested resource was not found.",
      retryable: false,
      requestId,
    };
  }

  // 413/Request Too Large (pattern match)
  if (normalized.includes("request_too_large") || normalized.includes("413")) {
    return {
      kind: "request_too_large",
      title: "File Too Large",
      message: "The file exceeds the maximum allowed size.",
      retryable: false,
      requestId,
    };
  }

  // Validation errors
  if (
    normalized.includes("bad_request") ||
    normalized.includes("validation") ||
    normalized.includes("invalid")
  ) {
    return {
      kind: "validation",
      title: "Invalid Input",
      message: "Please check your input and try again.",
      retryable: false,
      requestId,
    };
  }

  // 5xx/Server errors (pattern match)
  if (
    normalized.includes("internal_error") ||
    normalized.includes("500") ||
    normalized.includes("502") ||
    normalized.includes("service_unavailable") ||
    normalized.includes("503") ||
    normalized.includes("504")
  ) {
    return {
      kind: "server",
      title: "Server Error",
      message: "An unexpected error occurred. Please try again.",
      retryable: true,
      requestId,
    };
  }

  return {
    kind: "unknown",
    title: "Request Failed",
    message: rawError || "An unexpected error occurred.",
    retryable: true,
    requestId,
  };
}

/**
 * Primary error mapping function (handles both error codes and fallback strings)
 * @param errorInput - Either error code string or raw error message
 * @param requestId - Optional request ID to include
 */
export function mapErrorToUserFacing(errorInput: string, requestId?: string): UserFacingError {
  // First, try to map as error code
  const partial = mapErrorCodeToUserFacing(errorInput);
  if (partial.kind) {
    return { ...partial, requestId } as UserFacingError;
  }

  // Fall back to string pattern matching
  const result = mapErrorStringToUserFacing(errorInput);
  if (requestId && !result.requestId) {
    result.requestId = requestId;
  }
  return result;
}
