<script>
  import { onMount } from 'svelte';
  import { fetchMetrics, fetchHistory, metrics, history } from '$lib/stores/api.js';

  let loading = true;
  onMount(() => {
    Promise.all([fetchMetrics(), fetchHistory()]).then(() => loading = false);
    const i = setInterval(() => { fetchMetrics(); fetchHistory(); }, 2000);
    return () => clearInterval(i);
  });

  const cards = [
    { key: 'temperature', label: 'Temperature', unit: 'C' },
    { key: 'pressure',    label: 'Pressure',    unit: 'bar' },
    { key: 'vibration',   label: 'Vibration',   unit: 'mm/s' },
    { key: 'rpm',         label: 'Motor RPM',   unit: 'rpm' },
  ];

  function lastVal(data, key) {
    if (!data || !data.length) return '--';
    const found = [...data].reverse().find(x => x.key === key);
    return found ? found.value.toFixed(1) : '--';
  }
</script>

<div class="page-head">
  <h1>Dashboard</h1>
  <span class="live-badge">LIVE</span>
</div>

{#if loading}
  <div class="loading">Loading sensor data...</div>
{:else}
  <div class="cards">
    {#each cards as c}
      <div class="card">
        <div class="card-label">{c.label}</div>
        <div class="card-value">{lastVal($metrics, c.key)}<span class="card-unit">{c.unit}</span></div>
        <div class="card-bar"><div class="card-bar-fill" style="width: {Math.min(100, Math.max(0, parseFloat(lastVal($metrics, c.key)) / (c.key === 'rpm' ? 20 : 1)))}%"></div></div>
      </div>
    {/each}
  </div>

  <h2 class="section-title">Aggregate History</h2>
  <div class="table-wrap">
    <table>
      <thead><tr><th>Time</th><th>Metric</th><th>Avg</th><th>Min</th><th>Max</th><th>Count</th></tr></thead>
      <tbody>
        {#each $history.slice(0, 20) as r}
          <tr>
            <td class="mono">{new Date(r.time * 1000).toLocaleTimeString()}</td>
            <td>{r.metric}</td>
            <td class="mono">{r.avg.toFixed(2)}</td>
            <td class="mono">{r.min.toFixed(2)}</td>
            <td class="mono">{r.max.toFixed(2)}</td>
            <td class="mono">{r.count}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}

<style>
  .page-head { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 1.5rem; }
  .page-head h1 { font-size: 1.4rem; font-family: monospace; font-weight: 700; }
  .live-badge { font-family: monospace; font-size: 0.65rem; font-weight: 700; color: var(--ac); border: 1px solid var(--ac); padding: 0.15rem 0.6rem; border-radius: 4px; letter-spacing: 1px; animation: pulse 2s infinite; }
  @keyframes pulse { 0%,100%{opacity:1} 50%{opacity:0.35} }
  .loading { text-align: center; color: var(--dim); padding: 4rem; font-family: monospace; }
  .cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px,1fr)); gap: 1rem; margin-bottom: 2rem; }
  .card { background: var(--sf); border: 1px solid var(--bd); border-radius: 10px; padding: 1.25rem; transition: background 0.3s, border 0.3s; }
  .card-label { color: var(--dim); font-size: 0.8rem; font-weight: 600; margin-bottom: 0.5rem; letter-spacing: 0.3px; }
  .card-value { font-family: monospace; font-size: 2rem; font-weight: 700; color: var(--ac); line-height: 1.1; }
  .card-unit { font-size: 0.75rem; color: var(--dim); font-weight: 400; margin-left: 0.2rem; }
  .card-bar { height: 3px; background: var(--bd); border-radius: 2px; margin-top: 0.75rem; overflow: hidden; }
  .card-bar-fill { height: 100%; background: var(--ac); border-radius: 2px; transition: width 0.5s ease; }
  .section-title { font-family: monospace; font-size: 0.95rem; color: var(--dim); margin-bottom: 0.75rem; font-weight: 600; }
  .table-wrap { overflow-x: auto; background: var(--sf); border: 1px solid var(--bd); border-radius: 10px; transition: background 0.3s; }
  table { width: 100%; border-collapse: collapse; font-size: 0.82rem; }
  th { text-align: left; padding: 0.65rem 1rem; color: var(--dim); font-size: 0.7rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.5px; border-bottom: 1px solid var(--bd); }
  td { padding: 0.5rem 1rem; border-bottom: 1px solid var(--bd); }
  tr:last-child td { border: none; }
  tr:hover { background: var(--sf2); }
  .mono { font-family: monospace; font-size: 0.78rem; }
</style>
