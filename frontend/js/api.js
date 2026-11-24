const API_BASE = window.DEVTRAIL_API_BASE || (
  location.protocol === 'file:' ? 'http://localhost:8080' :
  (location.port && location.port !== '8080') ? 'http://localhost:8080' : ''
);

export class ApiError extends Error {
  constructor(message, { status, code, details, payload } = {}) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
    this.details = details;
    this.payload = payload;
  }
}

let authTokenProvider = () => null;
let onUnauthorized = () => {};

export function setAuthTokenProvider(fn) { authTokenProvider = fn; }
export function setUnauthorizedHandler(fn) { onUnauthorized = fn; }

function buildUrl(path, query) {
  const base = path.startsWith('http') ? path : `${API_BASE}${path}`;
  const url = new URL(base, location.origin);
  if (query) {
    for (const [k, v] of Object.entries(query)) {
      if (v == null || v === '') continue;
      url.searchParams.set(k, v);
    }
  }
  return url.toString();
}

async function parseResponse(res) {
  const ct = res.headers.get('content-type') || '';
  if (ct.includes('application/json')) return res.json();
  if (ct.includes('text/markdown') || ct.includes('text/plain')) return res.text();
  const text = await res.text();
  try { return JSON.parse(text); } catch { return text; }
}

export async function request(path, { method = 'GET', body, query, auth = true, timeout = 15000, headers = {} } = {}) {
  const controller = new AbortController();
  const t = setTimeout(() => controller.abort(), timeout);
  try {
    const h = { Accept: 'application/json', ...headers };
    const token = authTokenProvider();
    if (auth && token) h.Authorization = `Bearer ${token}`;
    let payloadBody;
    if (body !== undefined) {
      h['Content-Type'] = 'application/json';
      payloadBody = JSON.stringify(body);
    }
    const res = await fetch(buildUrl(path, query), { method, headers: h, body: payloadBody, signal: controller.signal });
    const payload = await parseResponse(res);
    if (!res.ok) {
      const err = payload?.error || {};
      const message = err.message || `Request failed (${res.status})`;
      const apiErr = new ApiError(message, { status: res.status, code: err.code, details: err.details, payload });
      if (res.status === 401 && auth) onUnauthorized(apiErr);
      throw apiErr;
    }
    return payload;
  } catch (err) {
    if (err.name === 'AbortError') throw new ApiError('Request timed out', { status: 0, code: 'timeout' });
    if (err instanceof ApiError) throw err;
    throw new ApiError(err.message || 'Network request failed', { status: 0, code: 'network_error' });
  } finally {
    clearTimeout(t);
  }
}

export const api = {
  get: (path, opts) => request(path, { ...opts, method: 'GET' }),
  post: (path, body, opts) => request(path, { ...opts, method: 'POST', body }),
  patch: (path, body, opts) => request(path, { ...opts, method: 'PATCH', body }),
  del: (path, opts) => request(path, { ...opts, method: 'DELETE' }),
};
