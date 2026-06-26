import { writable, get } from 'svelte/store';
import { env } from '$env/dynamic/public';

// Read runtime env vars injected by the Node adapter at request time.
// Fallbacks make local dev work without any .env file.
const BACKEND   = env.PUBLIC_BACKEND_URL  || 'http://localhost:8080';
const KC_URL    = env.PUBLIC_KC_URL       || 'http://localhost:8180';
const KC_REALM  = env.PUBLIC_KC_REALM     || 'edge';
const KC_CLIENT = env.PUBLIC_KC_CLIENT    || 'edge-frontend';

const browser = typeof window !== 'undefined';

// Persistent auth state (browser only)
export const token        = writable(browser ? localStorage.getItem('edge_token')    : null);
export const refreshToken = writable(browser ? localStorage.getItem('edge_refresh')  : null);
export const username     = writable(browser ? localStorage.getItem('edge_username') : null);

// Application data
export const status  = writable(null);
export const metrics = writable({});   // { [key]: MetricEvent } — latest value per key
export const history = writable([]);   // AggregateEvent[], newest first
export const rules   = writable([]);
export const alerts  = writable([]);   // RuleAlert[], newest first

// Sync auth state to localStorage on every change
if (browser) {
	token.subscribe(v        => v ? localStorage.setItem('edge_token', v)    : localStorage.removeItem('edge_token'));
	refreshToken.subscribe(v => v ? localStorage.setItem('edge_refresh', v)  : localStorage.removeItem('edge_refresh'));
	username.subscribe(v     => v ? localStorage.setItem('edge_username', v) : localStorage.removeItem('edge_username'));
}

function authHeaders() {
	const t = get(token);
	const h = { 'Content-Type': 'application/json' };
	if (t) h['Authorization'] = 'Bearer ' + t;
	return h;
}

// Refresh the access token using the stored refresh token.
async function refreshAccessToken() {
	const rt = get(refreshToken);
	if (!rt) return false;

	const body = new URLSearchParams({
		grant_type:    'refresh_token',
		client_id:     KC_CLIENT,
		refresh_token: rt,
	});
	const res = await fetch(
		`${KC_URL}/realms/${KC_REALM}/protocol/openid-connect/token`,
		{ method: 'POST', headers: { 'Content-Type': 'application/x-www-form-urlencoded' }, body },
	);
	if (!res.ok) { logout(); return false; }

	const data = await res.json();
	token.set(data.access_token);
	if (data.refresh_token) refreshToken.set(data.refresh_token);
	return true;
}

// Authenticated fetch with automatic token refresh on 401.
async function apiFetch(path, retry = true) {
	const res = await fetch(BACKEND + path, { headers: authHeaders() });
	if (res.status === 401) {
		if (retry && await refreshAccessToken()) return apiFetch(path, false);
		logout();
		throw new Error('unauthorized');
	}
	if (!res.ok) throw new Error(`API ${res.status}`);
	return res.json();
}

export async function login(user, pass) {
	const body = new URLSearchParams({
		grant_type: 'password',
		client_id:  KC_CLIENT,
		username:   user,
		password:   pass,
	});
	const res = await fetch(
		`${KC_URL}/realms/${KC_REALM}/protocol/openid-connect/token`,
		{ method: 'POST', headers: { 'Content-Type': 'application/x-www-form-urlencoded' }, body },
	);
	if (!res.ok) throw new Error('Login failed');
	const data = await res.json();
	token.set(data.access_token);
	refreshToken.set(data.refresh_token);
	username.set(user);
	return data;
}

export function logout() {
	stopStream();
	token.set(null);
	refreshToken.set(null);
	username.set(null);
	metrics.set({});
	history.set([]);
	alerts.set([]);
}

export async function fetchStatus()  {
	const d = await apiFetch('/api/v1/status');
	status.set(d);
	return d;
}

export async function fetchMetrics() {
	const data = await apiFetch('/api/v1/raw');
	// Convert array to latest-per-key map for O(1) lookups in the UI
	const map = {};
	for (const m of data) map[m.key] = m;
	metrics.set(map);
	return map;
}

export async function fetchHistory(windowMs = 1000, limit = 100) {
	const d = await apiFetch(`/api/v1/aggregates?window_ms=${windowMs}&limit=${limit}`);
	history.set(d);
	return d;
}

export async function fetchRules() {
	const d = await apiFetch('/api/v1/rules');
	rules.set(d);
	return d;
}

export async function createRule(rule, retry = true) {
	const res = await fetch(BACKEND + '/api/v1/rules', {
		method:  'POST',
		headers: authHeaders(),
		body:    JSON.stringify(rule),
	});
	if (res.status === 401) {
		if (retry && await refreshAccessToken()) return createRule(rule, false);
		logout();
		throw new Error('unauthorized');
	}
	if (res.status === 403) throw new Error('Forbidden: admin role required');
	if (!res.ok) throw new Error(`Create rule failed: ${res.status}`);
	return res.json();
}

// ── Server-Sent Events stream ─────────────────────────────────────────────────

let eventSource = null;

export function startStream() {
	if (!browser || eventSource) return;
	const t = get(token);
	if (!t) return;

	// EventSource cannot set custom headers, so we pass the token as a query param.
	const url = new URL(`${BACKEND}/api/v1/stream`);
	url.searchParams.set('token', t);

	eventSource = new EventSource(url.toString());

	eventSource.addEventListener('metric', (e) => {
		const m = JSON.parse(e.data);
		metrics.update(map => ({ ...map, [m.key]: m }));
	});

	eventSource.addEventListener('aggregate', (e) => {
		const a = JSON.parse(e.data);
		history.update(arr => {
			const next = [a, ...arr];
			return next.length > 100 ? next.slice(0, 100) : next;
		});
	});

	eventSource.addEventListener('alert', (e) => {
		const al = JSON.parse(e.data);
		alerts.update(arr => [al, ...arr].slice(0, 50));
	});

	eventSource.onerror = () => {
		eventSource?.close();
		eventSource = null;
		// Reconnect after a short delay if still authenticated
		setTimeout(() => { if (get(token)) startStream(); }, 3000);
	};
}

export function stopStream() {
	if (eventSource) {
		eventSource.close();
		eventSource = null;
	}
}
