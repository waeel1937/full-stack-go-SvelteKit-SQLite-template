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

<div class="page-head">
  <h1>Rules</h1>
  <button class="btn" on:click={() => show = !show}>{show ? 'Cancel' : '+ Add Rule'}</button>
</div>

{#if show}
  <div class="form">
    <div class="row">
      <label>Metric key<input bind:value={nr.key} placeholder="temperature" /></label>
      <label>Condition
        <select bind:value={nr.condition}>
          <option value="avg_gt">avg &gt; threshold</option>
          <option value="max_gt">max &gt; threshold</option>
          <option value="min_lt">min &lt; threshold</option>
        </select>
      </label>
      <label>Threshold<input type="number" bind:value={nr.threshold} step="0.1" /></label>
    </div>
    <label>Alert message<input bind:value={nr.message} placeholder="Temperature too high" /></label>
    <button class="btn primary" on:click={submit}>Create rule</button>
  </div>
{/if}

{#each $rules as r}
  <div class="ri" class:off={!r.enabled}>
    <div class="ri-head">
      <span class="ri-id">{r.id}</span>
      <span class="badge" class:active={r.enabled}>{r.enabled ? 'ACTIVE' : 'OFF'}</span>
    </div>
    <div class="ri-expr">{r.key} {r.condition} {r.threshold}</div>
    <div class="ri-msg">{r.message}</div>
  </div>
{:else}
  <div class="empty">No rules configured.</div>
{/each}

<style>
  .page-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }
  .page-head h1 { font-size: 1.4rem; font-family: monospace; font-weight: 700; }
  .btn { font-family: monospace; font-size: 0.8rem; padding: 0.5rem 1rem; background: var(--sf2); color: var(--tx); border: 1px solid var(--bd); border-radius: 6px; cursor: pointer; transition: all 0.2s; }
  .btn:hover { border-color: var(--ac); color: var(--ac); }
  .btn.primary { background: var(--ac); color: #0c1117; border-color: var(--ac); font-weight: 700; }
  .btn.primary:hover { background: var(--ac2); }
  .form { background: var(--sf); border: 1px solid var(--ac); border-radius: 10px; padding: 1.5rem; margin-bottom: 1.5rem; display: flex; flex-direction: column; gap: 1rem; }
  .row { display: grid; grid-template-columns: repeat(3,1fr); gap: 1rem; }
  label { display: flex; flex-direction: column; gap: 0.3rem; font-size: 0.78rem; color: var(--dim); font-weight: 600; }
  input, select { font-family: monospace; font-size: 0.85rem; padding: 0.55rem 0.7rem; background: var(--bg); color: var(--tx); border: 1px solid var(--bd); border-radius: 6px; outline: none; transition: border 0.2s; }
  input:focus, select:focus { border-color: var(--ac); }
  .ri { background: var(--sf); border: 1px solid var(--bd); border-radius: 8px; padding: 1rem; margin-bottom: 0.4rem; transition: background 0.3s; }
  .ri.off { opacity: 0.45; }
  .ri-head { display: flex; justify-content: space-between; margin-bottom: 0.3rem; }
  .ri-id { font-family: monospace; font-weight: 700; font-size: 0.88rem; }
  .ri-expr { font-family: monospace; font-size: 0.82rem; color: var(--ac); margin-bottom: 0.2rem; }
  .ri-msg { font-size: 0.82rem; color: var(--dim); }
  .badge { font-family: monospace; font-size: 0.65rem; padding: 0.1rem 0.5rem; border-radius: 4px; background: var(--sf2); color: var(--dim); }
  .badge.active { background: rgba(45,212,160,0.12); color: var(--gn); }
  .empty { padding: 3rem; text-align: center; color: var(--dim); background: var(--sf); border: 1px solid var(--bd); border-radius: 10px; }
</style>
