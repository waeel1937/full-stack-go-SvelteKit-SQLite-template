<script>
  import { onMount } from 'svelte';
  import { rules } from '$lib/stores/api.js';

  let alertLog = [];

  onMount(() => {
    const es = new EventSource('http://localhost:8080/api/v1/status');
    const i = setInterval(async () => {
      try {
        const res = await fetch('http://localhost:8080/api/v1/aggregates?window_ms=1000&limit=5');
        const data = await res.json();
        for (const a of data) {
          if (a.avg > 80 && a.metric === 'temperature') {
            alertLog = [{time: Date.now(), metric: a.metric, msg: 'Temperature above 80', val: a.avg}, ...alertLog].slice(0, 50);
          }
          if (a.max > 100 && a.metric === 'temperature') {
            alertLog = [{time: Date.now(), metric: a.metric, msg: 'Temperature critical', val: a.max}, ...alertLog].slice(0, 50);
          }
        }
      } catch(e) {}
    }, 3000);
    return () => clearInterval(i);
  });
</script>

<h1 style="font-family:monospace;font-size:1.4rem;margin-bottom:1.5rem">🔔 Alerts</h1>

{#if alertLog.length === 0}
  <div class="empty">No alerts yet. Rules are evaluating against incoming aggregates.</div>
{:else}
  <div class="list">
    {#each alertLog as a}
      <div class="item">
        <div style="font-weight:600">⚠️ {a.msg}</div>
        <div class="meta">
          <span class="tag">{a.metric}</span>
          <span style="color:var(--or);font-family:monospace">val: {a.val.toFixed(2)}</span>
          <span style="font-family:monospace;color:var(--dim)">{new Date(a.time).toLocaleString()}</span>
        </div>
      </div>
    {/each}
  </div>
{/if}

<style>
  .empty { padding: 3rem; text-align: center; color: var(--dim); background: var(--sf); border: 1px solid var(--bd); border-radius: 10px; }
  .list { display: flex; flex-direction: column; gap: 0.5rem; }
  .item { padding: 1rem; background: var(--sf); border: 1px solid var(--bd); border-left: 3px solid var(--or); border-radius: 8px; }
  .meta { display: flex; gap: 0.75rem; flex-wrap: wrap; font-size: 0.8rem; color: var(--dim); margin-top: 0.4rem; }
  .tag { font-family: monospace; background: var(--sf2); padding: 0.1rem 0.5rem; border-radius: 4px; font-size: 0.75rem; }
</style>
