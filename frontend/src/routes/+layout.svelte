<script>
  import { onMount } from 'svelte';
  import { fetchStatus, status, token, username, login, logout } from '$lib/stores/api.js';

  let loginUser = '';
  let loginPass = '';
  let loginError = '';

  onMount(() => {
    if ($token) {
      fetchStatus();
      const i = setInterval(fetchStatus, 5000);
      return () => clearInterval(i);
    }
  });

  async function handleLogin() {
    loginError = '';
    try {
      await login(loginUser, loginPass);
      loginUser = '';
      loginPass = '';
      fetchStatus();
    } catch(e) {
      loginError = 'Invalid credentials';
    }
  }
</script>

<svelte:head>
<style>
  :root {
    --bg: #0a0e17; --sf: #111827; --sf2: #1a2332; --bd: #1e2d3d;
    --tx: #e2e8f0; --dim: #64748b; --ac: #22d3ee; --gn: #34d399;
    --rd: #f87171; --or: #fb923c; --vi: #a78bfa;
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { background: var(--bg); color: var(--tx); font-family: system-ui, sans-serif; }
</style>
</svelte:head>

{#if !$token}
  <div class="login-wrap">
    <div class="login-box">
      <div class="login-logo">Edge<span style="color:var(--ac)">IIoT</span></div>
      <div class="login-title">Sign in</div>
      {#if loginError}
        <div class="login-err">{loginError}</div>
      {/if}
      <label>Username<input bind:value={loginUser} placeholder="admin" /></label>
      <label>Password<input type="password" bind:value={loginPass} placeholder="password" on:keydown={(e) => e.key === 'Enter' && handleLogin()} /></label>
      <button class="login-btn" on:click={handleLogin}>Login</button>
    </div>
  </div>
{:else}
  <div class="shell">
    <nav class="side">
      <div class="logo">Edge<span style="color:var(--ac)">IIoT</span></div>
      <a href="/">Dashboard</a>
      <a href="/alerts">Alerts</a>
      <a href="/rules">Rules</a>
      <div class="st">
        {#if $status}
          <div><span class="dot gn"></span> Online</div>
          <div class="dim">{$status.goroutines} goroutines</div>
          <div class="dim">{$status.memory_mb?.toFixed(1)} MB</div>
          <div class="dim">{$status.uptime_sec}s uptime</div>
        {:else}
          <div><span class="dot rd"></span> Connecting</div>
        {/if}
        <div class="user-info">
          <span class="dim">{$username}</span>
          <button class="logout-btn" on:click={logout}>Logout</button>
        </div>
      </div>
    </nav>
    <main><slot /></main>
  </div>
{/if}

<style>
  .login-wrap { display: flex; justify-content: center; align-items: center; min-height: 100vh; }
  .login-box { background: var(--sf); border: 1px solid var(--bd); border-radius: 12px; padding: 2.5rem; width: 340px; display: flex; flex-direction: column; gap: 1rem; }
  .login-logo { font-family: monospace; font-weight: 700; font-size: 1.3rem; text-align: center; }
  .login-title { text-align: center; color: var(--dim); font-size: 0.9rem; }
  .login-err { color: var(--rd); font-size: 0.8rem; text-align: center; }
  .login-box label { display: flex; flex-direction: column; gap: 0.3rem; font-size: 0.8rem; color: var(--dim); font-weight: 600; }
  .login-box input { font-family: monospace; font-size: 0.85rem; padding: 0.6rem; background: var(--bg); color: var(--tx); border: 1px solid var(--bd); border-radius: 6px; outline: none; }
  .login-box input:focus { border-color: var(--ac); }
  .login-btn { font-family: monospace; font-size: 0.85rem; padding: 0.6rem; background: var(--ac); color: var(--bg); border: none; border-radius: 6px; cursor: pointer; font-weight: 600; }
  .login-btn:hover { opacity: 0.9; }
  .shell { display: flex; min-height: 100vh; }
  .side { width: 200px; background: var(--sf); border-right: 1px solid var(--bd); padding: 1.5rem 1rem; display: flex; flex-direction: column; gap: 0.5rem; position: fixed; top: 0; left: 0; bottom: 0; }
  .logo { font-weight: 700; font-size: 1.1rem; margin-bottom: 1rem; font-family: monospace; }
  .side a { color: var(--dim); text-decoration: none; padding: 0.5rem 0.75rem; border-radius: 6px; font-size: 0.9rem; }
  .side a:hover { background: var(--sf2); color: var(--tx); }
  main { margin-left: 200px; padding: 2rem; flex: 1; }
  .st { margin-top: auto; padding: 0.75rem; background: var(--sf2); border-radius: 8px; font-size: 0.8rem; }
  .dim { color: var(--dim); font-size: 0.75rem; font-family: monospace; padding-left: 1rem; }
  .dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; margin-right: 0.4rem; }
  .dot.gn { background: var(--gn); box-shadow: 0 0 6px var(--gn); }
  .dot.rd { background: var(--rd); box-shadow: 0 0 6px var(--rd); }
  .user-info { display: flex; justify-content: space-between; align-items: center; margin-top: 0.5rem; padding-top: 0.5rem; border-top: 1px solid var(--bd); }
  .logout-btn { font-family: monospace; font-size: 0.7rem; padding: 0.2rem 0.5rem; background: none; color: var(--rd); border: 1px solid var(--rd); border-radius: 4px; cursor: pointer; }
  .logout-btn:hover { background: var(--rd); color: var(--bg); }
</style>
