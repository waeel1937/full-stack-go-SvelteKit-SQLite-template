// This is a private authenticated dashboard — disable SSR so the auth state
// is never rendered server-side (avoids the login flash on F5 refresh).
export const ssr = false;
