<script>
  import { alerts } from '$lib/stores/api.js';

  function timeAgo(ms) {
    const diff = Date.now() - ms;
    if (diff < 60000)  return Math.floor(diff / 1000) + 's ago';
    if (diff < 3600000) return Math.floor(diff / 60000) + 'm ago';
    return Math.floor(diff / 3600000) + 'h ago';
  }

  function severity(value, key) {
    if (key === 'temperature' && value > 85) return 'critical';
    if (key === 'vibration'   && value > 20) return 'critical';
    if (key === 'rpm'         && value > 1900) return 'critical';
    return 'warning';
  }
</script>

<div class="ph">
  <div class="ph-left">
    <h1>Alerts</h1>
    {#if $alerts.length > 0}
      <span class="count-badge">{$alerts.length}</span>
    {/if}
  </div>
  {#if $alerts.length > 0}
    <span class="ph-hint">Latest first · real-time via SSE</span>
  {/if}
</div>

{#if $alerts.length === 0}
  <div class="empty-state">
    <div class="empty-icon">
      <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
        <polyline points="22 4 12 14.01 9 11.01"/>
      </svg>
    </div>
    <div class="empty-title">All clear</div>
    <div class="empty-sub">No alerts triggered. Rules are evaluating incoming aggregates in real-time.</div>
  </div>
{:else}
  <div class="alert-list">
    {#each $alerts as a (a.time + a.key + a.message)}
      {@const sev = severity(a.value, a.key)}
      <div class="alert-item {sev}">
        <div class="al-accent"></div>
        <div class="al-body">
          <div class="al-top">
            <span class="al-msg">{a.message}</span>
            <span class="sev-badge {sev}">{sev.toUpperCase()}</span>
          </div>
          <div class="al-meta">
            <span class="tag">{a.key}</span>
            <span class="sep">·</span>
            <span class="al-val">{a.value.toFixed(2)}</span>
            <span class="sep">·</span>
            <span class="al-time" title={new Date(a.time).toLocaleString()}>
              {timeAgo(a.time)}
            </span>
          </div>
        </div>
      </div>
    {/each}
  </div>
{/if}

<style>
  /* ── HEADER ───────────────────────────────────────────────────── */
  .ph { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1.75rem; }
  .ph-left { display: flex; align-items: center; gap: .75rem; }
  .ph-left h1 { font-size: 1.35rem; font-weight: 700; }
  .count-badge {
    font-family: 'IBM Plex Mono', monospace; font-size: .68rem; font-weight: 700;
    background: rgba(255,87,87,.12); color: var(--rd);
    border: 1px solid rgba(255,87,87,.25);
    padding: .15rem .5rem; border-radius: 20px;
    min-width: 26px; text-align: center;
  }
  .ph-hint { font-size: .73rem; color: var(--dim); }

  /* ── EMPTY ───────────────────────────────────────────────────── */
  .empty-state {
    display: flex; flex-direction: column; align-items: center; gap: .625rem;
    padding: 4rem 2rem;
    background: var(--sf); border: 1px solid var(--bd); border-radius: 12px;
    text-align: center;
  }
  .empty-icon {
    width: 56px; height: 56px; border-radius: 50%;
    background: rgba(34,216,122,.08); border: 1px solid rgba(34,216,122,.15);
    display: flex; align-items: center; justify-content: center;
    color: var(--gn); margin-bottom: .25rem;
  }
  .empty-title { font-size: 1rem; font-weight: 700; color: var(--gn); }
  .empty-sub { font-size: .82rem; color: var(--dim); max-width: 380px; line-height: 1.6; }

  /* ── ALERT LIST ──────────────────────────────────────────────── */
  .alert-list { display: flex; flex-direction: column; gap: .5rem; }
  .alert-item {
    display: flex; overflow: hidden;
    background: var(--sf); border: 1px solid var(--bd);
    border-radius: 10px;
    transition: border-color .15s;
  }
  .alert-item:hover { border-color: var(--bd2); }
  .alert-item.critical { border-color: rgba(255,87,87,.3); }
  .alert-item.warning  { border-color: rgba(255,184,48,.2); }
  .al-accent {
    width: 3px; flex-shrink: 0;
  }
  .alert-item.critical .al-accent { background: var(--rd); }
  .alert-item.warning  .al-accent { background: var(--or); }
  .al-body { padding: .875rem 1rem; flex: 1; display: flex; flex-direction: column; gap: .4rem; }
  .al-top { display: flex; align-items: flex-start; justify-content: space-between; gap: .75rem; }
  .al-msg { font-size: .875rem; font-weight: 600; color: var(--tx); }
  .sev-badge {
    font-family: 'IBM Plex Mono', monospace;
    font-size: .58rem; font-weight: 700; letter-spacing: .6px;
    padding: .18rem .5rem; border-radius: 4px;
    white-space: nowrap; flex-shrink: 0;
  }
  .sev-badge.critical { background: rgba(255,87,87,.12); color: var(--rd); border: 1px solid rgba(255,87,87,.2); }
  .sev-badge.warning  { background: rgba(255,184,48,.1);  color: var(--or); border: 1px solid rgba(255,184,48,.2); }
  .al-meta {
    display: flex; align-items: center; gap: .5rem;
    font-size: .75rem; color: var(--dim); flex-wrap: wrap;
  }
  .tag {
    font-family: 'IBM Plex Mono', monospace; font-size: .7rem;
    background: var(--sf2); border: 1px solid var(--bd);
    padding: .1rem .42rem; border-radius: 5px; color: var(--tx2);
  }
  .sep { color: var(--dim2); }
  .al-val { font-family: 'IBM Plex Mono', monospace; color: var(--or); font-size: .75rem; }
  .al-time { font-family: 'IBM Plex Mono', monospace; font-size: .72rem; }
</style>
