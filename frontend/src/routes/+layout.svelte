<script>
  import { onMount } from 'svelte';
  import { token, username, status, fetchStatus, login, logout, startStream, stopStream } from '$lib/stores/api.js';

  let loginUser = '';
  let loginPass = '';
  let loginError = '';
  let dark = true;
  let statusInterval = null;

  function toggleTheme() {
    dark = !dark;
    document.documentElement.setAttribute('data-theme', dark ? 'dark' : 'light');
  }

  onMount(() => {
    document.documentElement.setAttribute('data-theme', 'dark');

    // React to auth state changes: start/stop status polling and SSE stream
    const unsub = token.subscribe(t => {
      clearInterval(statusInterval);
      statusInterval = null;
      if (t) {
        fetchStatus();
        startStream();
        statusInterval = setInterval(fetchStatus, 5000);
      } else {
        stopStream();
      }
    });

    return () => {
      unsub();
      clearInterval(statusInterval);
    };
  });

  async function handleLogin() {
    loginError = '';
    try {
      await login(loginUser, loginPass);
      loginUser = '';
      loginPass = '';
    } catch {
      loginError = 'Invalid credentials';
    }
  }
</script>

<svelte:head>
<style>
  :root, [data-theme="dark"] {
    --bg: #0c1117; --sf: #141c24; --sf2: #1a252f; --bd: #243040;
    --tx: #d4dce6; --dim: #5a6e82; --ac: #2dd4a0; --ac2: #1a9e76;
    --gn: #2dd4a0; --rd: #e85454; --or: #e8a54e;
  }
  [data-theme="light"] {
    --bg: #f0f4f2; --sf: #ffffff; --sf2: #e8eeea; --bd: #c8d4cc;
    --tx: #1a2e24; --dim: #5a7266; --ac: #1a9e76; --ac2: #148060;
    --gn: #1a9e76; --rd: #c43c3c; --or: #c48a30;
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { background: var(--bg); color: var(--tx); font-family: 'IBM Plex Sans', system-ui, sans-serif; transition: background 0.3s, color 0.3s; }
</style>
</svelte:head>

{#if !$token}
  <div class="login-wrap">
    <div class="login-box">
      <div class="login-logo">Edge<span class="accent">IIoT</span></div>
      <div class="login-sub">Industrial Edge Platform</div>
      {#if loginError}
        <div class="login-err">{loginError}</div>
      {/if}
      <label>Username<input bind:value={loginUser} placeholder="admin" /></label>
      <label>Password<input type="password" bind:value={loginPass} placeholder="password"
        onkeydown={(e) => e.key === 'Enter' && handleLogin()} /></label>
      <button class="btn-primary" onclick={handleLogin}>Sign in</button>
      <button class="btn-toggle" onclick={toggleTheme}>{dark ? 'Light mode' : 'Dark mode'}</button>
    </div>
  </div>
{:else}
  <div class="shell">
    <nav class="side">
      <div class="logo">Edge<span class="accent">IIoT</span></div>
      <div class="nav-section">
        <a href="/" class="nav-link">Dashboard</a>
        <a href="/alerts" class="nav-link">Alerts</a>
        <a href="/rules" class="nav-link">Rules</a>
      </div>
      <div class="side-footer">
        <button class="btn-toggle small" onclick={toggleTheme}>{dark ? 'Light' : 'Dark'}</button>
        {#if $status}
          <div class="status-row"><span class="dot on"></span> Online</div>
          <div class="status-detail">{$status.goroutines} goroutines</div>
          <div class="status-detail">{$status.memory_mb?.toFixed(1)} MB</div>
          <div class="status-detail">{$status.uptime_sec}s uptime</div>
        {:else}
          <div class="status-row"><span class="dot off"></span> Connecting</div>
        {/if}
        <div class="user-row">
          <span class="status-detail">{$username}</span>
          <button class="btn-logout" onclick={logout}>Logout</button>
        </div>
      </div>
    </nav>
    <main><slot /></main>
  </div>
{/if}

<style>
  .accent { color: var(--ac); }
  .login-wrap { display: flex; justify-content: center; align-items: center; min-height: 100vh; }
  .login-box { background: var(--sf); border: 1px solid var(--bd); border-radius: 12px; padding: 2.5rem; width: 360px; display: flex; flex-direction: column; gap: 1rem; }
  .login-logo { font-family: monospace; font-weight: 700; font-size: 1.5rem; text-align: center; letter-spacing: -0.5px; }
  .login-sub { text-align: center; color: var(--dim); font-size: 0.8rem; letter-spacing: 0.5px; }
  .login-err { color: var(--rd); font-size: 0.8rem; text-align: center; padding: 0.4rem; background: rgba(232,84,84,0.1); border-radius: 6px; }
  .login-box label { display: flex; flex-direction: column; gap: 0.3rem; font-size: 0.8rem; color: var(--dim); font-weight: 600; }
  .login-box input { font-family: monospace; font-size: 0.85rem; padding: 0.65rem 0.75rem; background: var(--bg); color: var(--tx); border: 1px solid var(--bd); border-radius: 8px; outline: none; transition: border 0.2s; }
  .login-box input:focus { border-color: var(--ac); }
  .btn-primary { font-family: monospace; font-size: 0.85rem; padding: 0.7rem; background: var(--ac); color: #0c1117; border: none; border-radius: 8px; cursor: pointer; font-weight: 700; transition: background 0.2s; }
  .btn-primary:hover { background: var(--ac2); }
  .btn-toggle { font-family: monospace; font-size: 0.75rem; padding: 0.4rem; background: none; color: var(--dim); border: 1px solid var(--bd); border-radius: 6px; cursor: pointer; transition: all 0.2s; }
  .btn-toggle:hover { color: var(--tx); border-color: var(--ac); }
  .btn-toggle.small { width: 100%; margin-bottom: 0.75rem; }
  .shell { display: flex; min-height: 100vh; }
  .side { width: 210px; background: var(--sf); border-right: 1px solid var(--bd); padding: 1.5rem 1rem; display: flex; flex-direction: column; position: fixed; top: 0; left: 0; bottom: 0; transition: background 0.3s; }
  .logo { font-weight: 700; font-size: 1.2rem; margin-bottom: 2rem; font-family: monospace; letter-spacing: -0.5px; }
  .nav-section { display: flex; flex-direction: column; gap: 0.25rem; }
  .nav-link { color: var(--dim); text-decoration: none; padding: 0.6rem 0.75rem; border-radius: 8px; font-size: 0.88rem; font-weight: 500; transition: all 0.15s; border-left: 2px solid transparent; }
  .nav-link:hover { background: var(--sf2); color: var(--ac); border-left-color: var(--ac); }
  main { margin-left: 210px; padding: 2rem 2.5rem; flex: 1; }
  .side-footer { margin-top: auto; padding: 0.75rem; background: var(--sf2); border-radius: 8px; font-size: 0.8rem; transition: background 0.3s; }
  .status-row { display: flex; align-items: center; gap: 0.4rem; font-weight: 600; font-size: 0.8rem; margin-bottom: 0.3rem; }
  .status-detail { color: var(--dim); font-size: 0.72rem; font-family: monospace; padding-left: 1rem; line-height: 1.5; }
  .dot { display: inline-block; width: 7px; height: 7px; border-radius: 50%; }
  .dot.on { background: var(--gn); box-shadow: 0 0 8px var(--gn); }
  .dot.off { background: var(--rd); box-shadow: 0 0 8px var(--rd); }
  .user-row { display: flex; justify-content: space-between; align-items: center; margin-top: 0.5rem; padding-top: 0.5rem; border-top: 1px solid var(--bd); }
  .btn-logout { font-family: monospace; font-size: 0.68rem; padding: 0.2rem 0.5rem; background: none; color: var(--dim); border: 1px solid var(--bd); border-radius: 4px; cursor: pointer; transition: all 0.2s; }
  .btn-logout:hover { color: var(--rd); border-color: var(--rd); }
</style>
