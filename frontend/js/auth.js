import { api, setAuthTokenProvider, setUnauthorizedHandler } from './api.js';

const TOKEN_KEY = 'devtrail_token';
let token = sessionStorage.getItem(TOKEN_KEY) || null;
let currentUser = null;

setAuthTokenProvider(() => token);
setUnauthorizedHandler(() => {
  clearAuth();
  const next = encodeURIComponent(location.pathname + location.search);
  location.href = `/pages/login.html?next=${next}`;
});

export function getToken() { return token; }
export function getUser() { return currentUser; }
export function isAuthenticated() { return !!token; }

export function setAuth(nextToken, user) {
  token = nextToken || null;
  currentUser = user || null;
  if (token) sessionStorage.setItem(TOKEN_KEY, token);
  else sessionStorage.removeItem(TOKEN_KEY);
}

export function clearAuth() {
  token = null;
  currentUser = null;
  sessionStorage.removeItem(TOKEN_KEY);
}

export async function login(email, password) {
  const res = await api.post('/api/v1/auth/login', { email, password }, { auth: false });
  setAuth(res.token, res.user);
  return res;
}

export async function register(displayName, email, password) {
  const res = await api.post('/api/v1/auth/register', { display_name: displayName, email, password }, { auth: false });
  setAuth(res.token, res.user);
  return res;
}

export async function logout() {
  try { if (token) await api.post('/api/v1/auth/logout'); } catch { /* clear local state regardless */ }
  clearAuth();
  location.href = '/pages/login.html';
}

export async function fetchMe() {
  if (!token) return null;
  const user = await api.get('/api/v1/me');
  currentUser = user;
  return user;
}

export async function requireAuth() {
  if (!token) {
    const next = encodeURIComponent(location.pathname + location.search);
    location.href = `/pages/login.html?next=${next}`;
    throw new Error('redirecting to login');
  }
  if (currentUser) return currentUser;
  try {
    return await fetchMe();
  } catch (err) {
    clearAuth();
    const next = encodeURIComponent(location.pathname + location.search);
    location.href = `/pages/login.html?next=${next}`;
    throw err;
  }
}

export function redirectIfAuthenticated() {
  if (token) { location.href = '/pages/overview.html'; return true; }
  return false;
}
