<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
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
    return () => { unsub(); clearInterval(statusInterval); };
  });

  async function handleLogin() {
    loginError = '';
    try {
      await login(loginUser, loginPass);
      loginUser = '';
      loginPass = '';
    } catch {
      loginError = 'Invalid credentials. Please try again.';
    }
  }

  function uptimeFmt(sec) {
    if (!sec) return '--';
    if (sec < 60) return sec + 's';
    if (sec < 3600) return Math.floor(sec / 60) + 'm ' + (sec % 60) + 's';
    return Math.floor(sec / 3600) + 'h ' + Math.floor((sec % 3600) / 60) + 'm';
  }

  $: activePath = $page?.url?.pathname ?? '/';
</script>

<svelte:head>
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link href="https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;500;600&family=IBM+Plex+Sans:wght@400;500;600;700&display=swap" rel="stylesheet" />
  <style>
    :root, [data-theme="dark"] {
      --bg:   #07111b;
      --sf:   #0d1e2d;
      --sf2:  #132538;
      --bd:   #1a3047;
      --bd2:  #243d56;
      --tx:   #dce9f4;
      --tx2:  #a8c0d6;
      --dim:  #4c6e8a;
      --dim2: #38536a;
      --ac:   #00d9b8;
      --ac2:  #00a88f;
      --ac3:  rgba(0,217,184,0.07);
      --ac4:  rgba(0,217,184,0.14);
      --bl:   #4d9eff;
      --gn:   #22d87a;
      --rd:   #ff5757;
      --or:   #ffb830;
      --shadow: 0 4px 24px rgba(0,0,0,0.5);
    }
    [data-theme="light"] {
      --bg:   #f0f5fb;
      --sf:   #ffffff;
      --sf2:  #e8eef7;
      --bd:   #c4d4e6;
      --bd2:  #a8bdd4;
      --tx:   #0e2035;
      --tx2:  #2a4a68;
      --dim:  #5a7898;
      --dim2: #8aa0b8;
      --ac:   #009e82;
      --ac2:  #007a65;
      --ac3:  rgba(0,158,130,0.07);
      --ac4:  rgba(0,158,130,0.14);
      --bl:   #2c7cd6;
      --gn:   #15a857;
      --rd:   #d63333;
      --or:   #c47a00;
      --shadow: 0 4px 24px rgba(0,0,0,0.1);
    }
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    html { scroll-behavior: smooth; }
    body {
      background: var(--bg);
      color: var(--tx);
      font-family: 'IBM Plex Sans', system-ui, sans-serif;
      font-size: 14px;
      line-height: 1.5;
      -webkit-font-smoothing: antialiased;
      transition: background 0.25s, color 0.25s;
    }
    ::-webkit-scrollbar { width: 5px; height: 5px; }
    ::-webkit-scrollbar-track { background: transparent; }
    ::-webkit-scrollbar-thumb { background: var(--bd2); border-radius: 3px; }
  </style>
</svelte:head>

{#if !$token}
  <!-- ── LOGIN ── -->
  <div class="login-scene">
    <div class="login-glow"></div>
    <div class="login-panel">
      <div class="login-brand">
        <div class="brand-mark">
          <svg width="26" height="26" viewBox="0 0 32 32" fill="none">
            <rect x="2"  y="2"  width="12" height="12" rx="2" fill="var(--ac)"/>
            <rect x="18" y="2"  width="12" height="12" rx="2" fill="var(--ac)" opacity=".4"/>
            <rect x="2"  y="18" width="12" height="12" rx="2" fill="var(--ac)" opacity=".4"/>
            <rect x="18" y="18" width="12" height="12" rx="2" fill="var(--ac)"/>
          </svg>
        </div>
        <div>
          <div class="brand-name">Edge<span class="ac">IIoT</span></div>
          <div class="brand-sub">Industrial Edge Platform</div>
        </div>
      </div>

      {#if loginError}
        <div class="login-err">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
          {loginError}
        </div>
      {/if}

      <div class="lf-fields">
        <label class="lf-field">
          <span>Username</span>
          <input bind:value={loginUser} placeholder="admin" autocomplete="username" />
        </label>
        <label class="lf-field">
          <span>Password</span>
          <input type="password" bind:value={loginPass} placeholder="••••••••"
            autocomplete="current-password"
            onkeydown={(e) => e.key === 'Enter' && handleLogin()} />
        </label>
        <button class="btn-signin" onclick={handleLogin}>Sign in →</button>
      </div>

      <button class="btn-theme-toggle" onclick={toggleTheme}>
        {dark ? '☀ Light mode' : '☽ Dark mode'}
      </button>
    </div>
  </div>

{:else}
  <!-- ── APP SHELL ── -->
  <div class="shell">
    <aside class="sidebar">
      <!-- Brand -->
      <div class="sb-brand">
        <div class="sb-mark">
          <svg width="18" height="18" viewBox="0 0 32 32" fill="none">
            <rect x="2"  y="2"  width="12" height="12" rx="2" fill="var(--ac)"/>
            <rect x="18" y="2"  width="12" height="12" rx="2" fill="var(--ac)" opacity=".4"/>
            <rect x="2"  y="18" width="12" height="12" rx="2" fill="var(--ac)" opacity=".4"/>
            <rect x="18" y="18" width="12" height="12" rx="2" fill="var(--ac)"/>
          </svg>
        </div>
        <span class="sb-logo">Edge<span class="ac">IIoT</span></span>
      </div>

      <!-- Navigation -->
      <nav class="sb-nav">
        <div class="nav-section-label">Monitor</div>
        <a href="/" class="nav-link" class:active={activePath === '/'}>
          <svg class="ni" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
            <rect x="3" y="3" width="7" height="7" rx="1"/>
            <rect x="14" y="3" width="7" height="7" rx="1"/>
            <rect x="3" y="14" width="7" height="7" rx="1"/>
            <rect x="14" y="14" width="7" height="7" rx="1"/>
          </svg>
          Dashboard
        </a>
        <a href="/alerts" class="nav-link" class:active={activePath === '/alerts'}>
          <svg class="ni" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
            <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/>
            <path d="M13.73 21a2 2 0 0 1-3.46 0"/>
          </svg>
          Alerts
        </a>

        <div class="nav-section-label" style="margin-top:.75rem">Config</div>
        <a href="/rules" class="nav-link" class:active={activePath === '/rules'}>
          <svg class="ni" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
            <line x1="4" y1="21" x2="4" y2="14"/>
            <line x1="4" y1="10" x2="4" y2="3"/>
            <line x1="12" y1="21" x2="12" y2="12"/>
            <line x1="12" y1="8" x2="12" y2="3"/>
            <line x1="20" y1="21" x2="20" y2="16"/>
            <line x1="20" y1="12" x2="20" y2="3"/>
            <line x1="1" y1="14" x2="7" y2="14"/>
            <line x1="9" y1="8" x2="15" y2="8"/>
            <line x1="17" y1="16" x2="23" y2="16"/>
          </svg>
          Rules
        </a>
      </nav>

      <!-- Footer / Status -->
      <div class="sb-footer">
        {#if $status}
          <div class="sf-status-row">
            <span class="dot-live"></span>
            <span class="sf-status-label">System Online</span>
          </div>
          <div class="sf-metrics">
            <div class="sf-metric">
              <span class="sf-mk">UPTIME</span>
              <span class="sf-mv">{uptimeFmt($status.uptime_sec)}</span>
            </div>
            <div class="sf-metric">
              <span class="sf-mk">RAM</span>
              <span class="sf-mv">{$status.memory_mb?.toFixed(1)} MB</span>
            </div>
            <div class="sf-metric">
              <span class="sf-mk">GO</span>
              <span class="sf-mv">{$status.goroutines}</span>
            </div>
          </div>
        {:else}
          <div class="sf-status-row">
            <span class="dot-off"></span>
            <span class="sf-connecting">Connecting…</span>
          </div>
        {/if}

        <div class="sf-user">
          <div class="sf-user-left">
            <div class="sf-avatar">{($username || 'U')[0].toUpperCase()}</div>
            <span class="sf-uname">{$username}</span>
          </div>
          <div class="sf-user-right">
            <button class="icon-btn" title="{dark ? 'Light mode' : 'Dark mode'}" onclick={toggleTheme}>
              {dark ? '☀' : '☽'}
            </button>
            <button class="btn-out" onclick={logout}>Out</button>
          </div>
        </div>
      </div>
    </aside>

    <main class="main">
      <slot />
    </main>
  </div>
{/if}

<style>
  .ac { color: var(--ac); }

  /* ─── LOGIN ──────────────────────────────────────────────────────── */
  .login-scene {
    min-height: 100vh;
    display: flex; align-items: center; justify-content: center;
    position: relative; overflow: hidden;
  }
  .login-glow {
    position: absolute; inset: 0; pointer-events: none;
    background:
      radial-gradient(ellipse 55% 45% at 25% 35%, rgba(0,217,184,.07) 0%, transparent 65%),
      radial-gradient(ellipse 45% 55% at 75% 65%, rgba(77,158,255,.05) 0%, transparent 65%);
  }
  .login-panel {
    position: relative;
    background: var(--sf);
    border: 1px solid var(--bd);
    border-radius: 16px;
    padding: 2.25rem 2.25rem 1.75rem;
    width: 380px;
    box-shadow: var(--shadow);
    display: flex; flex-direction: column; gap: 1.25rem;
  }
  .login-brand {
    display: flex; align-items: center; gap: .75rem;
    padding-bottom: 1rem;
    border-bottom: 1px solid var(--bd);
  }
  .brand-mark {
    width: 42px; height: 42px;
    background: var(--sf2); border: 1px solid var(--bd);
    border-radius: 10px;
    display: flex; align-items: center; justify-content: center;
    flex-shrink: 0;
  }
  .brand-name {
    font-family: 'IBM Plex Mono', monospace;
    font-weight: 700; font-size: 1.15rem; line-height: 1;
  }
  .brand-sub { font-size: .72rem; color: var(--dim); margin-top: .2rem; letter-spacing: .3px; }
  .login-err {
    display: flex; align-items: center; gap: .4rem;
    color: var(--rd); font-size: .8rem;
    padding: .5rem .75rem;
    background: rgba(255,87,87,.08);
    border: 1px solid rgba(255,87,87,.2);
    border-radius: 8px;
  }
  .lf-fields { display: flex; flex-direction: column; gap: .875rem; }
  .lf-field { display: flex; flex-direction: column; gap: .3rem; }
  .lf-field span { font-size: .73rem; font-weight: 600; color: var(--dim); letter-spacing: .3px; }
  .lf-field input {
    font-family: 'IBM Plex Mono', monospace; font-size: .875rem;
    padding: .6rem .875rem;
    background: var(--bg); color: var(--tx);
    border: 1px solid var(--bd); border-radius: 8px;
    outline: none; transition: border-color .15s, box-shadow .15s;
  }
  .lf-field input:focus { border-color: var(--ac); box-shadow: 0 0 0 3px var(--ac3); }
  .btn-signin {
    font-family: 'IBM Plex Mono', monospace; font-size: .875rem; font-weight: 600;
    padding: .7rem; margin-top: .125rem;
    background: var(--ac); color: #030a0f;
    border: none; border-radius: 8px; cursor: pointer;
    transition: background .15s, transform .1s;
    letter-spacing: .3px;
  }
  .btn-signin:hover { background: var(--ac2); }
  .btn-signin:active { transform: scale(.98); }
  .btn-theme-toggle {
    font-size: .73rem; padding: .3rem .7rem;
    background: none; color: var(--dim);
    border: 1px solid var(--bd); border-radius: 6px;
    cursor: pointer; align-self: center;
    transition: all .15s;
  }
  .btn-theme-toggle:hover { color: var(--tx); border-color: var(--bd2); }

  /* ─── SHELL ──────────────────────────────────────────────────────── */
  .shell { display: flex; min-height: 100vh; }

  /* ─── SIDEBAR ────────────────────────────────────────────────────── */
  .sidebar {
    width: 218px;
    background: var(--sf);
    border-right: 1px solid var(--bd);
    display: flex; flex-direction: column;
    position: fixed; inset: 0 auto 0 0;
    transition: background .25s;
    z-index: 10;
  }
  .sb-brand {
    display: flex; align-items: center; gap: .625rem;
    padding: 1.125rem 1rem;
    border-bottom: 1px solid var(--bd);
    flex-shrink: 0;
  }
  .sb-mark {
    width: 30px; height: 30px;
    background: var(--sf2); border: 1px solid var(--bd);
    border-radius: 7px;
    display: flex; align-items: center; justify-content: center;
    flex-shrink: 0;
  }
  .sb-logo { font-family: 'IBM Plex Mono', monospace; font-weight: 700; font-size: .95rem; }
  .sb-nav {
    flex: 1; overflow-y: auto;
    padding: .875rem .625rem;
    display: flex; flex-direction: column; gap: .1rem;
  }
  .nav-section-label {
    font-size: .63rem; font-weight: 700; letter-spacing: .8px;
    color: var(--dim2); text-transform: uppercase;
    padding: 0 .5rem; margin-bottom: .2rem;
  }
  .nav-link {
    display: flex; align-items: center; gap: .55rem;
    padding: .5rem .625rem;
    border-radius: 7px;
    font-size: .855rem; font-weight: 500;
    color: var(--dim);
    text-decoration: none;
    border-left: 2px solid transparent;
    transition: all .15s;
  }
  .nav-link:hover { background: var(--sf2); color: var(--tx2); }
  .nav-link.active { background: var(--ac3); color: var(--ac); border-left-color: var(--ac); }
  .nav-link.active .ni { opacity: 1; }
  .ni { flex-shrink: 0; opacity: .55; transition: opacity .15s; }
  .nav-link:hover .ni { opacity: .8; }

  /* ─── SIDEBAR FOOTER ─────────────────────────────────────────────── */
  .sb-footer {
    flex-shrink: 0;
    padding: .875rem;
    border-top: 1px solid var(--bd);
    display: flex; flex-direction: column; gap: .625rem;
  }
  .sf-status-row { display: flex; align-items: center; gap: .5rem; }
  .dot-live {
    width: 7px; height: 7px; border-radius: 50%;
    background: var(--gn);
    flex-shrink: 0;
    animation: livePulse 2.5s ease-in-out infinite;
  }
  @keyframes livePulse {
    0%,100% { box-shadow: 0 0 0 2px rgba(34,216,122,.2); }
    50%      { box-shadow: 0 0 0 5px rgba(34,216,122,0); }
  }
  .dot-off {
    width: 7px; height: 7px; border-radius: 50%;
    background: var(--dim2); flex-shrink: 0;
  }
  .sf-status-label { font-size: .78rem; font-weight: 600; color: var(--gn); }
  .sf-connecting { font-size: .78rem; color: var(--dim); }
  .sf-metrics {
    display: grid; grid-template-columns: repeat(3,1fr);
    background: var(--sf2); border: 1px solid var(--bd);
    border-radius: 8px; padding: .5rem .4rem;
    gap: .1rem;
  }
  .sf-metric { display: flex; flex-direction: column; gap: .05rem; padding: 0 .1rem; }
  .sf-mk {
    font-family: 'IBM Plex Mono', monospace;
    font-size: .57rem; font-weight: 600;
    color: var(--dim2); letter-spacing: .5px;
  }
  .sf-mv {
    font-family: 'IBM Plex Mono', monospace;
    font-size: .7rem; color: var(--tx);
  }
  .sf-user {
    display: flex; align-items: center; justify-content: space-between;
    padding-top: .5rem; border-top: 1px solid var(--bd);
  }
  .sf-user-left { display: flex; align-items: center; gap: .45rem; }
  .sf-avatar {
    width: 25px; height: 25px; border-radius: 50%;
    background: var(--ac3); border: 1px solid var(--ac4);
    display: flex; align-items: center; justify-content: center;
    font-family: 'IBM Plex Mono', monospace;
    font-size: .68rem; font-weight: 700; color: var(--ac);
    flex-shrink: 0;
  }
  .sf-uname {
    font-family: 'IBM Plex Mono', monospace; font-size: .72rem;
    color: var(--tx); max-width: 75px;
    overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
  }
  .sf-user-right { display: flex; align-items: center; gap: .25rem; }
  .icon-btn {
    font-size: .72rem; padding: .18rem .32rem;
    background: none; color: var(--dim);
    border: 1px solid var(--bd); border-radius: 5px;
    cursor: pointer; transition: all .15s; line-height: 1;
  }
  .icon-btn:hover { color: var(--tx); border-color: var(--bd2); }
  .btn-out {
    font-family: 'IBM Plex Mono', monospace; font-size: .65rem; font-weight: 600;
    padding: .18rem .42rem;
    background: none; color: var(--dim);
    border: 1px solid var(--bd); border-radius: 5px;
    cursor: pointer; transition: all .15s;
  }
  .btn-out:hover { color: var(--rd); border-color: var(--rd); }

  /* ─── MAIN CONTENT ───────────────────────────────────────────────── */
  .main {
    margin-left: 218px; padding: 2rem 2.5rem;
    flex: 1; min-height: 100vh;
  }
</style>
