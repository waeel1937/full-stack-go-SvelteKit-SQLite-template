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
    { key: 'temperature', label: 'Temperature', unit: '°C', color: '#f87171' },
    { key: 'pressure',    label: 'Pressure',    unit: 'bar', color: '#22d3ee' },
    { key: 'vibration',   label: 'Vibration',   unit: 'mm/s', color: '#fb923c' },
    { key: 'rpm',         label: 'Motor RPM',   unit: 'rpm', color: '#a78bfa' },
  ];

  function lastVal(data, key) {
    if (!data || !data.length) return '--';
    const found = [...data].reverse().find(x => x.key === key);
    return found ? found.value.toFixed(1) : '--';
  }
</script>

<h1 class="hdr">
  Dashboard <span class="live">LIVE</span>
</h1>

{#if loading}
  <p class="load">Loading...</p>
{:else}
  <div class="cards">
    {#each cards as c}
      <div class="card">
        <div class="cl">{c.label}</div>
        <div class="cv" style="color:{c.color}">{lastVal($metrics, c.key)}<span class="cu">{c.unit}</span></div>
      </div>
    {/each}
  </div>

  <h2 class="sh">Aggregate History</h2>
  <div class="tw">
    <table>
      <thead><tr><th>Time</th><th>Metric</th><th>Avg</th><th>Min</th><th>Max</th><th>Count</th></tr></thead>
      <tbody>
        {#each $history.slice(0, 20) as r}
          <tr>
            <td class="m">{new Date(r.time * 1000).toLocaleTimeString()}</td>
            <td>{r.metric}</td>
            <td class="m">{r.avg.toFixed(2)}</td>
            <td class="m">{r.min.toFixed(2)}</td>
            <td class="m">{r.max.toFixed(2)}</td>
            <td class="m">{r.count}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}

<style>
  .hdr { font-size: 1.4rem; margin-bottom: 1.5rem; font-family: monospace; }
  .live { font-size: 0.7rem; color: var(--ac); border: 1px solid var(--ac); padding: 0.15rem 0.5rem; border-radius: 4px; animation: pulse 2s infinite; }
  @keyframes pulse { 0%,100%{opacity:1} 50%{opacity:0.4} }
  .load { color: var(--dim); text-align: center; padding: 3rem; }
  .cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px,1fr)); gap: 1rem; }
  .card { background: var(--sf); border: 1px solid var(--bd); border-radius: 10px; padding: 1.2rem; }
  .cl { color: var(--dim); font-size: 0.85rem; margin-bottom: 0.5rem; }
  .cv { font-family: monospace; font-size: 2rem; font-weight: 700; }
  .cu { font-size: 0.8rem; color: var(--dim); font-weight: 400; }
  .sh { font-family: monospace; font-size: 1rem; color: var(--dim); margin: 2rem 0 1rem; }
  .tw { overflow-x: auto; background: var(--sf); border: 1px solid var(--bd); border-radius: 10px; }
  table { width: 100%; border-collapse: collapse; font-size: 0.85rem; }
  th { text-align: left; padding: 0.7rem 1rem; color: var(--dim); font-size: 0.75rem; text-transform: uppercase; border-bottom: 1px solid var(--bd); }
  td { padding: 0.5rem 1rem; border-bottom: 1px solid var(--bd); }
  tr:last-child td { border: none; }
  tr:hover { background: var(--sf2); }
  .m { font-family: monospace; font-size: 0.8rem; }
</style>
