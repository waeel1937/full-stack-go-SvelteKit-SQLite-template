<script>
  import { onMount } from 'svelte';
  import { fetchMetrics, fetchHistory, metrics, history } from '$lib/stores/api.js';

  const cards = [
    { key: 'temperature', label: 'Temperature', unit: '°C',   max: 120,  icon: '🌡', color: '--or' },
    { key: 'pressure',    label: 'Pressure',    unit: 'bar',  max: 80,   icon: '⟳', color: '--bl' },
    { key: 'vibration',   label: 'Vibration',   unit: 'mm/s', max: 30,   icon: '〜', color: '--ac' },
    { key: 'rpm',         label: 'Motor RPM',   unit: 'rpm',  max: 2000, icon: '⚙', color: '--gn' },
  ];

  let loading = true;
  let error = '';

  onMount(async () => {
    try {
      await Promise.all([fetchMetrics(), fetchHistory()]);
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  });

  function lastVal(map, key) {
    const m = map?.[key];
    return m ? parseFloat(m.value.toFixed(2)) : null;
  }

  function barPct(val, max) {
    if (val == null) return 0;
    return Math.min(100, Math.max(0, (val / max) * 100));
  }

  function barColor(pct) {
    if (pct > 85) return 'var(--rd)';
    if (pct > 65) return 'var(--or)';
    return 'var(--ac)';
  }

  function sparkPath(hist, key, w = 88, h = 32) {
    const pts = hist
      .filter(h => h.metric === key)
      .slice(0, 24)
      .reverse()
      .map(h => h.avg);
    if (pts.length < 2) return '';
    const mn = Math.min(...pts);
    const mx = Math.max(...pts);
    const range = mx - mn || 1;
    return 'M' + pts.map((v, i) => {
      const x = ((i / (pts.length - 1)) * w).toFixed(1);
      const y = (h - 2 - ((v - mn) / range) * (h - 4)).toFixed(1);
      return `${x},${y}`;
    }).join('L');
  }

  function trendLabel(hist, key) {
    const pts = hist.filter(h => h.metric === key).slice(0, 6).map(h => h.avg);
    if (pts.length < 2) return '';
    const delta = pts[0] - pts[pts.length - 1];
    if (Math.abs(delta) < 0.5) return '→';
    return delta > 0 ? '↑' : '↓';
  }

  function trendClass(hist, key) {
    const pts = hist.filter(h => h.metric === key).slice(0, 6).map(h => h.avg);
    if (pts.length < 2) return '';
    const delta = pts[0] - pts[pts.length - 1];
    if (Math.abs(delta) < 0.5) return 'flat';
    return delta > 0 ? 'up' : 'down';
  }

  function fmtTime(sec) {
    return new Date(sec * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  }
</script>

<div class="ph">
  <div class="ph-left">
    <h1>Dashboard</h1>
    <span class="live-pill">
      <span class="live-dot"></span>LIVE
    </span>
  </div>
  <div class="ph-right">
    {#if !loading && !error}
      <span class="ph-count">{Object.keys($metrics).length} sensors active</span>
    {/if}
  </div>
</div>

{#if loading}
  <div class="state-block">
    <div class="spinner"></div>
    <span>Loading sensor data…</span>
  </div>
{:else if error}
  <div class="state-block error">
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
    </svg>
    {error}
  </div>
{:else}
  <!-- Metric Cards -->
  <div class="cards">
    {#each cards as c}
      {@const val    = lastVal($metrics, c.key)}
      {@const pct    = barPct(val, c.max)}
      {@const spark  = sparkPath($history, c.key)}
      {@const trend  = trendLabel($history, c.key)}
      {@const tclass = trendClass($history, c.key)}
      <div class="card">
        <div class="card-top">
          <div class="card-label-row">
            <span class="card-label">{c.label}</span>
            {#if trend}
              <span class="trend {tclass}">{trend}</span>
            {/if}
          </div>
          {#if spark}
            <svg class="sparkline" width="88" height="32" viewBox="0 0 88 32" preserveAspectRatio="none">
              <defs>
                <linearGradient id="sg-{c.key}" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="var({c.color})" stop-opacity=".3"/>
                  <stop offset="100%" stop-color="var({c.color})" stop-opacity="0"/>
                </linearGradient>
              </defs>
              <path d="{spark}Z L 88,32 L 0,32" fill="url(#sg-{c.key})"/>
              <path d="{spark}" fill="none" stroke="var({c.color})" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          {/if}
        </div>

        <div class="card-value">
          {val != null ? val : '--'}
          <span class="card-unit">{c.unit}</span>
        </div>

        <div class="card-bar-track">
          <div class="card-bar-fill" style="width:{pct}%; background:{barColor(pct)};"></div>
        </div>
        <div class="card-bar-labels">
          <span>0</span>
          <span class="pct">{pct.toFixed(0)}%</span>
          <span>{c.max}{c.unit}</span>
        </div>
      </div>
    {/each}
  </div>

  <!-- Aggregate History -->
  <div class="section-header">
    <h2>Aggregate History</h2>
    <span class="section-count">{$history.length} records</span>
  </div>

  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th>Time</th>
          <th>Metric</th>
          <th>Window</th>
          <th class="num">Avg</th>
          <th class="num">Min</th>
          <th class="num">Max</th>
          <th class="num">Count</th>
        </tr>
      </thead>
      <tbody>
        {#each $history.slice(0, 25) as r (r.time + r.metric)}
          <tr>
            <td class="mono dim">{fmtTime(r.time)}</td>
            <td>
              <span class="metric-tag">{r.metric}</span>
            </td>
            <td class="mono dim">{r.window_ms ?? r.window ?? 1000}ms</td>
            <td class="mono num ac">{r.avg.toFixed(2)}</td>
            <td class="mono num dim">{r.min.toFixed(2)}</td>
            <td class="mono num dim">{r.max.toFixed(2)}</td>
            <td class="mono num dim">{r.count}</td>
          </tr>
        {:else}
          <tr>
            <td colspan="7" class="empty-row">
              No aggregates yet — waiting for the first window to close…
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}

<style>
  /* ── PAGE HEADER ──────────────────────────────────────────────── */
  .ph {
    display: flex; align-items: center; justify-content: space-between;
    margin-bottom: 1.75rem;
  }
  .ph-left { display: flex; align-items: center; gap: .75rem; }
  .ph-left h1 {
    font-family: 'IBM Plex Sans', sans-serif;
    font-size: 1.35rem; font-weight: 700;
  }
  .live-pill {
    display: inline-flex; align-items: center; gap: .35rem;
    font-family: 'IBM Plex Mono', monospace;
    font-size: .62rem; font-weight: 700; letter-spacing: 1px;
    color: var(--ac);
    border: 1px solid var(--ac4);
    background: var(--ac3);
    padding: .18rem .55rem;
    border-radius: 20px;
  }
  .live-dot {
    width: 5px; height: 5px; border-radius: 50%;
    background: var(--ac);
    animation: blink 1.8s ease-in-out infinite;
  }
  @keyframes blink { 0%,100%{opacity:1} 50%{opacity:.2} }
  .ph-right { display: flex; align-items: center; gap: .5rem; }
  .ph-count { font-size: .78rem; color: var(--dim); }

  /* ── STATE BLOCKS ─────────────────────────────────────────────── */
  .state-block {
    display: flex; flex-direction: column; align-items: center;
    gap: .75rem;
    padding: 4rem 2rem;
    font-family: 'IBM Plex Mono', monospace; font-size: .875rem;
    color: var(--dim);
    background: var(--sf); border: 1px solid var(--bd); border-radius: 12px;
  }
  .state-block.error { color: var(--rd); }
  .spinner {
    width: 24px; height: 24px; border-radius: 50%;
    border: 2px solid var(--bd);
    border-top-color: var(--ac);
    animation: spin 0.8s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }

  /* ── METRIC CARDS ─────────────────────────────────────────────── */
  .cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(230px, 1fr));
    gap: 1rem;
    margin-bottom: 2rem;
  }
  .card {
    background: var(--sf); border: 1px solid var(--bd);
    border-radius: 12px; padding: 1.125rem 1.25rem 1rem;
    display: flex; flex-direction: column; gap: .625rem;
    transition: border-color .2s, box-shadow .2s;
  }
  .card:hover { border-color: var(--bd2); box-shadow: var(--shadow); }
  .card-top { display: flex; align-items: flex-start; justify-content: space-between; }
  .card-label-row { display: flex; flex-direction: column; gap: .2rem; }
  .card-label {
    font-size: .73rem; font-weight: 600; letter-spacing: .3px;
    color: var(--dim); text-transform: uppercase;
  }
  .trend {
    font-family: 'IBM Plex Mono', monospace; font-size: .75rem; font-weight: 600;
  }
  .trend.up   { color: var(--rd); }
  .trend.down { color: var(--bl); }
  .trend.flat { color: var(--dim); }
  .sparkline { flex-shrink: 0; }
  .card-value {
    font-family: 'IBM Plex Mono', monospace;
    font-size: 2.1rem; font-weight: 600;
    color: var(--tx); line-height: 1;
    letter-spacing: -1px;
  }
  .card-unit {
    font-family: 'IBM Plex Sans', sans-serif;
    font-size: .78rem; font-weight: 400;
    color: var(--dim); margin-left: .15rem;
    letter-spacing: 0;
  }
  .card-bar-track {
    height: 4px; background: var(--bd); border-radius: 2px; overflow: hidden;
  }
  .card-bar-fill {
    height: 100%; border-radius: 2px;
    transition: width .6s ease, background .4s;
  }
  .card-bar-labels {
    display: flex; justify-content: space-between;
    font-family: 'IBM Plex Mono', monospace; font-size: .62rem; color: var(--dim2);
  }
  .pct { color: var(--dim); }

  /* ── HISTORY TABLE ────────────────────────────────────────────── */
  .section-header {
    display: flex; align-items: center; gap: .75rem;
    margin-bottom: .875rem;
  }
  .section-header h2 {
    font-family: 'IBM Plex Sans', sans-serif;
    font-size: .95rem; font-weight: 600; color: var(--tx2);
  }
  .section-count {
    font-family: 'IBM Plex Mono', monospace; font-size: .72rem; color: var(--dim);
  }
  .table-wrap {
    background: var(--sf); border: 1px solid var(--bd);
    border-radius: 12px; overflow: hidden;
    overflow-x: auto;
  }
  table { width: 100%; border-collapse: collapse; font-size: .82rem; }
  thead tr { border-bottom: 1px solid var(--bd); }
  th {
    text-align: left; padding: .625rem 1rem;
    font-size: .65rem; font-weight: 700; letter-spacing: .5px;
    text-transform: uppercase; color: var(--dim);
    white-space: nowrap;
  }
  th.num { text-align: right; }
  td { padding: .5rem 1rem; border-bottom: 1px solid var(--bd); }
  tr:last-child td { border-bottom: none; }
  tr:hover td { background: var(--sf2); }
  td.num { text-align: right; }
  .mono { font-family: 'IBM Plex Mono', monospace; font-size: .78rem; }
  .dim  { color: var(--dim); }
  .ac   { color: var(--ac); }
  .metric-tag {
    font-family: 'IBM Plex Mono', monospace; font-size: .72rem;
    background: var(--sf2); border: 1px solid var(--bd);
    padding: .1rem .45rem; border-radius: 5px;
    color: var(--tx2);
  }
  .empty-row {
    text-align: center; padding: 3rem 1rem;
    color: var(--dim); font-family: 'IBM Plex Mono', monospace; font-size: .8rem;
  }
</style>
