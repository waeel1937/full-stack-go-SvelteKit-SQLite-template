<script>
  import { alerts } from '$lib/stores/api.js';
  // Alerts are pushed in real-time via the SSE stream started in +layout.svelte.
  // No polling or manual fetch needed.
</script>

<div class="page-head">
  <h1>Alerts</h1>
  <span class="count">{$alerts.length} events</span>
</div>

{#if $alerts.length === 0}
  <div class="empty">No alerts triggered. Rules are evaluating incoming aggregates in real-time.</div>
{:else}
  <div class="list">
    {#each $alerts as a (a.time + a.key + a.message)}
      <div class="item">
        <div class="item-bar"></div>
        <div class="item-body">
          <div class="item-msg">{a.message}</div>
          <div class="item-meta">
            <span class="tag">{a.key}</span>
            <span class="val">{a.value.toFixed(2)}</span>
            <span class="time">{new Date(a.time).toLocaleTimeString()}</span>
          </div>
        </div>
      </div>
    {/each}
  </div>
{/if}

<style>
  .page-head { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 1.5rem; }
  .page-head h1 { font-size: 1.4rem; font-family: monospace; font-weight: 700; }
  .count { font-family: monospace; font-size: 0.75rem; color: var(--dim); }
  .empty { padding: 3rem; text-align: center; color: var(--dim); background: var(--sf); border: 1px solid var(--bd); border-radius: 10px; font-size: 0.88rem; }
  .list { display: flex; flex-direction: column; gap: 0.4rem; }
  .item { display: flex; background: var(--sf); border: 1px solid var(--bd); border-radius: 8px; overflow: hidden; transition: background 0.3s; }
  .item-bar { width: 3px; background: var(--or); flex-shrink: 0; }
  .item-body { padding: 0.8rem 1rem; flex: 1; }
  .item-msg { font-weight: 600; font-size: 0.88rem; margin-bottom: 0.3rem; }
  .item-meta { display: flex; gap: 0.75rem; flex-wrap: wrap; font-size: 0.75rem; color: var(--dim); }
  .tag { font-family: monospace; background: var(--sf2); padding: 0.1rem 0.5rem; border-radius: 4px; }
  .val { font-family: monospace; color: var(--or); }
  .time { font-family: monospace; }
</style>
