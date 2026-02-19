import { ApiError, ApiResult } from './types';

export type RequestInterceptor = (config: RequestConfig) => RequestConfig | Promise<RequestConfig>;
export type ResponseInterceptor = (response: Response, config: RequestConfig) => Response | Promise<Response>;
export type ErrorInterceptor = (error: ApiError, config: RequestConfig) => ApiError | Promise<ApiError>;

export interface RequestConfig extends Omit<RequestInit, 'body'> {
  url: string;
  params?: Record<string, string | number | boolean | undefined>;
  body?: unknown;
  timeout?: number;
}

export interface HttpClientConfig {
  baseURL: string;
  defaultHeaders?: Record<string, string>;
  credentials?: RequestCredentials;
  timeout?: number;
}

export class HttpClient {
  private baseURL: string;
  private defaultHeaders: Record<string, string>;
  private credentials: RequestCredentials;
  private timeout: number;
  private requestInterceptors: RequestInterceptor[] = [];
  private responseInterceptors: ResponseInterceptor[] = [];
  private errorInterceptors: ErrorInterceptor[] = [];

  constructor(config: HttpClientConfig) {
    this.baseURL = config.baseURL.replace(/\/$/, '');
    this.defaultHeaders = config.defaultHeaders || {};
    this.credentials = config.credentials || 'same-origin';
    this.timeout = config.timeout || 30000;
  }

  addRequestInterceptor(interceptor: RequestInterceptor): () => void {
    this.requestInterceptors.push(interceptor);
    return () => {
      const index = this.requestInterceptors.indexOf(interceptor);
      if (index !== -1) this.requestInterceptors.splice(index, 1);
    };
  }

  addResponseInterceptor(interceptor: ResponseInterceptor): () => void {
    this.responseInterceptors.push(interceptor);
    return () => {
      const index = this.responseInterceptors.indexOf(interceptor);
      if (index !== -1) this.responseInterceptors.splice(index, 1);
    };
  }

  addErrorInterceptor(interceptor: ErrorInterceptor): () => void {
    this.errorInterceptors.push(interceptor);
    return () => {
      const index = this.errorInterceptors.indexOf(interceptor);
      if (index !== -1) this.errorInterceptors.splice(index, 1);
    };
  }

  private buildURL(path: string, params?: Record<string, string | number | boolean | undefined>): string {
    const url = path.startsWith('http') ? path : `${this.baseURL}${path.startsWith('/') ? path : `/${path}`}`;
    
    if (!params) return url;

    const searchParams = new URLSearchParams();
    for (const [key, value] of Object.entries(params)) {
      if (value !== undefined) {
        searchParams.append(key, String(value));
      }
    }
    const queryString = searchParams.toString();
    return queryString ? `${url}?${queryString}` : url;
  }

  private async runRequestInterceptors(config: RequestConfig): Promise<RequestConfig> {
    let result = config;
    for (const interceptor of this.requestInterceptors) {
      result = await interceptor(result);
    }
    return result;
  }

  private async runResponseInterceptors(response: Response, config: RequestConfig): Promise<Response> {
    let result = response;
    for (const interceptor of this.responseInterceptors) {
      result = await interceptor(result, config);
    }
    return result;
  }

  private async runErrorInterceptors(error: ApiError, config: RequestConfig): Promise<ApiError> {
    let result = error;
    for (const interceptor of this.errorInterceptors) {
      result = await interceptor(result, config);
    }
    return result;
  }

  private prepareBody(body: unknown): { body: BodyInit | undefined; contentType: string | undefined } {
    if (body === undefined || body === null) {
      return { body: undefined, contentType: undefined };
    }

    if (body instanceof FormData) {
      return { body, contentType: undefined };
    }

    if (body instanceof URLSearchParams) {
      return { body, contentType: 'application/x-www-form-urlencoded' };
    }

    if (body instanceof Blob || body instanceof ArrayBuffer) {
      return { body: body as BodyInit, contentType: undefined };
    }

    return { body: JSON.stringify(body), contentType: 'application/json' };
  }

  async request<T>(config: RequestConfig): Promise<T> {
    const processedConfig = await this.runRequestInterceptors(config);
    const url = this.buildURL(processedConfig.url, processedConfig.params);
    const { body, contentType } = this.prepareBody(processedConfig.body);

    const headers: Record<string, string> = {
      ...this.defaultHeaders,
    };

    if (contentType) {
      headers['Content-Type'] = contentType;
    }

    if (processedConfig.headers) {
      const configHeaders = processedConfig.headers as Record<string, string>;
      Object.assign(headers, configHeaders);
    }

    const controller = new AbortController();
    const timeout = processedConfig.timeout || this.timeout;
    const timeoutId = setTimeout(() => controller.abort(), timeout);

    try {
      let response = await fetch(url, {
        ...processedConfig,
        headers,
        body,
        credentials: processedConfig.credentials || this.credentials,
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      response = await this.runResponseInterceptors(response, processedConfig);

      if (!response.ok) {
        const errorBody = await response.json().catch(() => ({}));
        const error = ApiError.fromResponse(response.status, errorBody);
        const processedError = await this.runErrorInterceptors(error, processedConfig);
        throw processedError;
      }

      const data = await response.json();
      return data as T;
    } catch (err) {
      clearTimeout(timeoutId);

      if (err instanceof ApiError) {
        throw err;
      }

      if (err instanceof Error) {
        if (err.name === 'AbortError') {
          const error = new ApiError('Request timeout', undefined, 'TIMEOUT');
          throw await this.runErrorInterceptors(error, processedConfig);
        }
        const error = new ApiError(err.message, undefined, 'NETWORK_ERROR');
        throw await this.runErrorInterceptors(error, processedConfig);
      }

      const error = new ApiError('Unknown error', undefined, 'UNKNOWN');
      throw await this.runErrorInterceptors(error, processedConfig);
    }
  }

  async get<T>(path: string, params?: Record<string, string | number | boolean | undefined>, options?: Omit<RequestConfig, 'url' | 'params' | 'method'>): Promise<T> {
    return this.request<T>({ ...options, url: path, params, method: 'GET' });
  }

  async post<T>(path: string, body?: unknown, options?: Omit<RequestConfig, 'url' | 'body' | 'method'>): Promise<T> {
    return this.request<T>({ ...options, url: path, body, method: 'POST' });
  }

  async patch<T>(path: string, body?: unknown, options?: Omit<RequestConfig, 'url' | 'body' | 'method'>): Promise<T> {
    return this.request<T>({ ...options, url: path, body, method: 'PATCH' });
  }

  async put<T>(path: string, body?: unknown, options?: Omit<RequestConfig, 'url' | 'body' | 'method'>): Promise<T> {
    return this.request<T>({ ...options, url: path, body, method: 'PUT' });
  }

  async delete<T>(path: string, options?: Omit<RequestConfig, 'url' | 'method'>): Promise<T> {
    return this.request<T>({ ...options, url: path, method: 'DELETE' });
  }

  async safeRequest<T>(config: RequestConfig): Promise<ApiResult<T>> {
    try {
      const data = await this.request<T>(config);
      return { data };
    } catch (err) {
      if (err instanceof ApiError) {
        let message = err.message;
        if (err.code) {
          message += ` (${err.code})`;
        }
        if (err.validationReason) {
          message += ` [${err.validationReason}]`;
        }
        if (err.requestId) {
          message += ` request_id=${err.requestId}`;
        }
        const detailsValidationReason = err.details?.validation_reason;
        if (typeof detailsValidationReason === 'string' && detailsValidationReason && !err.validationReason) {
          message += ` [${detailsValidationReason}]`;
        }
        return { error: message };
      }
      return { error: err instanceof Error ? err.message : 'Unknown error' };
    }
  }

  async safeGet<T>(path: string, params?: Record<string, string | number | boolean | undefined>, options?: Omit<RequestConfig, 'url' | 'params' | 'method'>): Promise<ApiResult<T>> {
    return this.safeRequest<T>({ ...options, url: path, params, method: 'GET' });
  }

  async safePost<T>(path: string, body?: unknown, options?: Omit<RequestConfig, 'url' | 'body' | 'method'>): Promise<ApiResult<T>> {
    return this.safeRequest<T>({ ...options, url: path, body, method: 'POST' });
  }

  async safePatch<T>(path: string, body?: unknown, options?: Omit<RequestConfig, 'url' | 'body' | 'method'>): Promise<ApiResult<T>> {
    return this.safeRequest<T>({ ...options, url: path, body, method: 'PATCH' });
  }
}

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export const backendClient = new HttpClient({
  baseURL: API_URL,
  defaultHeaders: {},
  credentials: 'omit',
  timeout: 150000, // 150s to account for AI timeout (120s) + processing + network latency
});

export const nextApiClient = new HttpClient({
  baseURL: '',
  defaultHeaders: {},
  credentials: 'include',
  timeout: 30000,
});

// BYOK (Bring Your Own Key): attach user's OpenAI API key to backend requests
backendClient.addRequestInterceptor((config) => {
  if (typeof window === 'undefined') return config;
  try {
    // Add session ID header (for quota tracking)
    const { getOrCreateSessionID } = require('./telemetry');
    const sessionId = getOrCreateSessionID();
    if (sessionId) {
      const headers = (config.headers || {}) as Record<string, string>;
      headers['X-Session-ID'] = sessionId;
      config.headers = headers;
    }

    // Add OpenAI API Key header (for BYOK)
    const stored = localStorage.getItem('mdflow-openai-key');
    if (stored) {
      const parsed = JSON.parse(stored);
      const apiKey = parsed?.state?.apiKey;
      if (apiKey) {
        const headers = (config.headers || {}) as Record<string, string>;
        headers['X-OpenAI-API-Key'] = apiKey;
        config.headers = headers;
      }
    }
  } catch {
    // Ignore localStorage errors
  }
  return config;
});

nextApiClient.addResponseInterceptor(async (response, config) => {
  if (response.status === 401 || response.status === 403) {
    if (typeof window !== 'undefined' && config.url?.includes('/api/gsheet')) {
      window.dispatchEvent(new CustomEvent('google-auth-required'));
    }
  }
  return response;
});

if (process.env.NODE_ENV === 'development') {
  backendClient.addErrorInterceptor((error, config) => {
    console.error(`[backendClient] ${config.method || 'GET'} ${config.url}:`, error.message);
    return error;
  });

  nextApiClient.addErrorInterceptor((error, config) => {
    console.error(`[nextApiClient] ${config.method || 'GET'} ${config.url}:`, error.message);
    return error;
  });
}
