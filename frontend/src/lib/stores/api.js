import { writable, get } from 'svelte/store';
const BACKEND = typeof window !== 'undefined' ? 'http://localhost:8080' : 'http://backend:8080';
const KC_URL = 'http://localhost:8180';
const KC_REALM = 'edge';
const KC_CLIENT = 'edge-frontend';

const browser = typeof window !== 'undefined';

// beim Start aus localStorage einlesen (nur im Browser, nicht bei SSR)
export const token        = writable(browser ? localStorage.getItem('edge_token')    : null);
export const refreshToken = writable(browser ? localStorage.getItem('edge_refresh')  : null);
export const username      = writable(browser ? localStorage.getItem('edge_username') : null);
export const status  = writable(null);
export const metrics = writable([]);
export const history = writable([]);
export const rules   = writable([]);

// jede Aenderung sofort persistieren / loeschen
if (browser) {
  token.subscribe(v        => v ? localStorage.setItem('edge_token', v)    : localStorage.removeItem('edge_token'));
  refreshToken.subscribe(v => v ? localStorage.setItem('edge_refresh', v)  : localStorage.removeItem('edge_refresh'));
  username.subscribe(v     => v ? localStorage.setItem('edge_username', v) : localStorage.removeItem('edge_username'));
}

function headers() {
  const t = get(token);
  const h = { 'Content-Type': 'application/json' };
  if (t) h['Authorization'] = 'Bearer ' + t;
  return h;
}

// Access-Token per Refresh-Token erneuern
async function refresh() {
  const rt = get(refreshToken);
  if (!rt) return false;
  const body = new URLSearchParams({
    grant_type: 'refresh_token',
    client_id: KC_CLIENT,
    refresh_token: rt
  });
  const res = await fetch(
    KC_URL + '/realms/' + KC_REALM + '/protocol/openid-connect/token',
    { method: 'POST', headers: { 'Content-Type': 'application/x-www-form-urlencoded' }, body }
  );
  if (!res.ok) { logout(); return false; }
  const data = await res.json();
  token.set(data.access_token);
  if (data.refresh_token) refreshToken.set(data.refresh_token);
  return true;
}

async function f(path, retry = true) {
  const res = await fetch(BACKEND + path, { headers: headers() });
  if (res.status === 401) {
    // einmal Token erneuern und Request wiederholen, sonst ausloggen
    if (retry && await refresh()) return f(path, false);
    logout();
    throw new Error('unauthorized');
  }
  if (!res.ok) throw new Error('API ' + res.status);
  return res.json();
}

export async function login(user, pass) {
  const body = new URLSearchParams({
    grant_type: 'password',
    client_id: KC_CLIENT,
    username: user,
    password: pass
  });
  const res = await fetch(
    KC_URL + '/realms/' + KC_REALM + '/protocol/openid-connect/token',
    { method: 'POST', headers: { 'Content-Type': 'application/x-www-form-urlencoded' }, body }
  );
  if (!res.ok) throw new Error('Login failed');
  const data = await res.json();
  token.set(data.access_token);
  refreshToken.set(data.refresh_token);
  username.set(user);
  return data;
}

export function logout() {
  token.set(null);
  refreshToken.set(null);
  username.set(null);
}

export async function fetchStatus()  { const d = await f('/api/v1/status');  status.set(d);  return d; }
export async function fetchMetrics() { const d = await f('/api/v1/raw'); metrics.set(d); return d; }
export async function fetchHistory() { const d = await f('/api/v1/aggregates?window_ms=1000&limit=100'); history.set(d); return d; }
export async function fetchRules()   { const d = await f('/api/v1/rules'); rules.set(d); return d; }

export async function createRule(rule) {
  const res = await fetch(BACKEND + '/api/v1/rules', {
    method: 'POST',
    headers: headers(),
    body: JSON.stringify(rule)
  });
  return res.json();
}
