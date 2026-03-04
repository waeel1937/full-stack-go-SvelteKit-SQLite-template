import { writable, get } from 'svelte/store';

const BACKEND = typeof window !== 'undefined' ? 'http://localhost:8080' : 'http://backend:8080';
const KC_URL = 'http://localhost:8180';
const KC_REALM = 'edge';
const KC_CLIENT = 'edge-frontend';

export const token = writable(null);
export const username = writable(null);
export const status = writable(null);
export const metrics = writable([]);
export const history = writable([]);
export const rules = writable([]);

function headers() {
  const t = get(token);
  const h = { 'Content-Type': 'application/json' };
  if (t) h['Authorization'] = 'Bearer ' + t;
  return h;
}

async function f(path) {
  const res = await fetch(BACKEND + path, { headers: headers() });
  if (res.status === 401) { token.set(null); username.set(null); throw new Error('unauthorized'); }
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
  username.set(user);
  return data;
}

export function logout() {
  token.set(null);
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
