<script>
  import { onMount } from 'svelte';
  import { fetchStatus, status } from '$lib/stores/api.js';
  onMount(() => {
    fetchStatus();
    const i = setInterval(fetchStatus, 5000);
    return () => clearInterval(i);
  });
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
    </div>
  </nav>
  <main><slot /></main>
</div>

<style>
  .shell { display: flex; min-height: 100vh; }
  .side {
    width: 200px; background: var(--sf); border-right: 1px solid var(--bd);
    padding: 1.5rem 1rem; display: flex; flex-direction: column; gap: 0.5rem;
    position: fixed; top: 0; left: 0; bottom: 0;
  }
  .logo { font-weight: 700; font-size: 1.1rem; margin-bottom: 1rem; font-family: monospace; }
  .side a {
    color: var(--dim); text-decoration: none; padding: 0.5rem 0.75rem;
    border-radius: 6px; font-size: 0.9rem;
  }
  .side a:hover { background: var(--sf2); color: var(--tx); }
  main { margin-left: 200px; padding: 2rem; flex: 1; }
  .st { margin-top: auto; padding: 0.75rem; background: var(--sf2); border-radius: 8px; font-size: 0.8rem; }
  .dim { color: var(--dim); font-size: 0.75rem; font-family: monospace; padding-left: 1rem; }
  .dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; margin-right: 0.4rem; }
  .dot.gn { background: var(--gn); box-shadow: 0 0 6px var(--gn); }
  .dot.rd { background: var(--rd); box-shadow: 0 0 6px var(--rd); }
</style>
