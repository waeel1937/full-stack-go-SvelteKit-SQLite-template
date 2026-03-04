<script>
  import { onMount } from 'svelte';
  import { fetchRules, createRule, rules } from '$lib/stores/api.js';

  let nr = { id: '', key: '', condition: 'avg_gt', threshold: 0, message: '', enabled: true };
  let show = false;

  onMount(() => { fetchRules(); });

  async function submit() {
    if (!nr.key || !nr.message) return;
    await createRule(nr);
    nr = { id: '', key: '', condition: 'avg_gt', threshold: 0, message: '', enabled: true };
    show = false;
    fetchRules();
  }
</script>

<div class="top">
  <h1 class="hdr">⚡ Rules</h1>
  <button class="btn" on:click={() => show = !show}>{show ? 'Cancel' : '+ Add Rule'}</button>
</div>

{#if show}
  <div class="form">
    <div class="row">
      <label>Key<input bind:value={nr.key} placeholder="temperature" /></label>
      <label>Condition
        <select bind:value={nr.condition}>
          <option value="avg_gt">avg &gt; threshold</option>
          <option value="max_gt">max &gt; threshold</option>
          <option value="min_lt">min &lt; threshold</option>
        </select>
      </label>
      <label>Threshold<input type="number" bind:value={nr.threshold} step="0.1" /></label>
    </div>
    <label>Message<input bind:value={nr.message} placeholder="Temperature too high" /></label>
    <button class="btn ac" on:click={submit}>Create</button>
  </div>
{/if}

{#each $rules as r}
  <div class="ri" class:off={!r.enabled}>
    <div class="rh">
      <span class="rid">{r.id}</span>
      <span class="badge" class:active={r.enabled}>{r.enabled ? 'ACTIVE' : 'OFF'}</span>
    </div>
    <code>{r.key} {r.condition} {r.threshold}</code>
    <div class="rm">{r.message}</div>
  </div>
{:else}
  <div class="empty">No rules.</div>
{/each}

<style>
  .top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }
  .hdr { font-family: monospace; font-size: 1.4rem; }
  .btn { font-family: monospace; font-size: 0.8rem; padding: 0.5rem 1rem; background: var(--sf2); color: var(--tx); border: 1px solid var(--bd); border-radius: 6px; cursor: pointer; }
  .btn:hover { background: var(--bd); }
  .btn.ac { background: var(--ac); color: var(--bg); border-color: var(--ac); font-weight: 600; }
  .form { background: var(--sf); border: 1px solid var(--ac); border-radius: 10px; padding: 1.5rem; margin-bottom: 1.5rem; display: flex; flex-direction: column; gap: 1rem; }
  .row { display: grid; grid-template-columns: repeat(3,1fr); gap: 1rem; }
  label { display: flex; flex-direction: column; gap: 0.3rem; font-size: 0.8rem; color: var(--dim); font-weight: 600; }
  input, select { font-family: monospace; font-size: 0.85rem; padding: 0.5rem; background: var(--bg); color: var(--tx); border: 1px solid var(--bd); border-radius: 6px; outline: none; }
  input:focus, select:focus { border-color: var(--ac); }
  .ri { background: var(--sf); border: 1px solid var(--bd); border-radius: 8px; padding: 1rem; margin-bottom: 0.5rem; }
  .ri.off { opacity: 0.5; }
  .rh { display: flex; justify-content: space-between; margin-bottom: 0.3rem; }
  .rid { font-family: monospace; font-weight: 700; }
  code { color: var(--ac); font-size: 0.85rem; }
  .rm { font-size: 0.85rem; color: var(--dim); margin-top: 0.3rem; }
  .badge { font-family: monospace; font-size: 0.7rem; padding: 0.1rem 0.5rem; border-radius: 4px; background: var(--sf2); color: var(--dim); }
  .badge.active { background: rgba(52,211,153,0.15); color: var(--gn); }
  .empty { padding: 3rem; text-align: center; color: var(--dim); background: var(--sf); border: 1px solid var(--bd); border-radius: 10px; }
</style>
