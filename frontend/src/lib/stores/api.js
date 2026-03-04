import { writable } from 'svelte/store';

const BACKEND = typeof window !== 'undefined' ? 'http://localhost:8080' : 'http://backend:8080';

async function f(path) {
  const res = await fetch(BACKEND + path);
  if (!res.ok) throw new Error('API ' + res.status);
  return res.json();
}

export const status = writable(null);
export const metrics = writable([]);
export const history = writable([]);
export const rules = writable([]);

export async function fetchStatus()  { const d = await f('/api/v1/status');  status.set(d);  return d; }
export async function fetchMetrics() { const d = await f('/api/v1/raw'); metrics.set(d); return d; }
export async function fetchHistory() { const d = await f('/api/v1/aggregates?window_ms=1000&limit=100'); history.set(d); return d; }
export async function fetchRules()   { const d = await f('/api/v1/rules'); rules.set(d); return d; }

export async function createRule(rule) {
  const res = await fetch(BACKEND + '/api/v1/rules', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(rule)
  });
  return res.json();
}
