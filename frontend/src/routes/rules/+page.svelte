<script>
  import { onMount } from 'svelte';
  import { fetchRules, createRule, rules } from '$lib/stores/api.js';

  const blank = () => ({ id: '', key: '', condition: 'avg_gt', threshold: 0, message: '', enabled: true });
  let nr = blank();
  let show = false;
  let submitError = '';
  let submitting = false;

  const conditionLabels = {
    avg_gt: 'avg >',
    max_gt: 'max >',
    min_lt: 'min <',
  };

  onMount(() => { fetchRules(); });

  async function submit() {
    if (!nr.id || !nr.key || !nr.message) {
      submitError = 'Rule ID, metric key, and alert message are required.';
      return;
    }
    submitError = '';
    submitting = true;
    try {
      await createRule(nr);
      nr = blank();
      show = false;
      await fetchRules();
    } catch (e) {
      submitError = e.message;
    } finally {
      submitting = false;
    }
  }

  function cancelForm() {
    show = false;
    submitError = '';
    nr = blank();
  }
</script>

<div class="ph">
  <div class="ph-left">
    <h1>Rules</h1>
    <span class="rule-count">{$rules.length} rule{$rules.length !== 1 ? 's' : ''}</span>
  </div>
  <button class="btn {show ? 'cancel' : 'add'}" onclick={() => show ? cancelForm() : (show = true)}>
    {#if show}
      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
        <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
      </svg>
      Cancel
    {:else}
      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
        <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
      </svg>
      Add Rule
    {/if}
  </button>
</div>

{#if show}
  <div class="form-panel">
    <div class="form-title">New Threshold Rule</div>

    {#if submitError}
      <div class="form-err">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
        </svg>
        {submitError}
      </div>
    {/if}

    <div class="form-row">
      <label class="field">
        <span>Rule ID</span>
        <input bind:value={nr.id} placeholder="temp-warn" />
      </label>
      <label class="field">
        <span>Metric Key</span>
        <input bind:value={nr.key} placeholder="temperature" />
      </label>
      <label class="field">
        <span>Condition</span>
        <select bind:value={nr.condition}>
          <option value="avg_gt">avg &gt; threshold</option>
          <option value="max_gt">max &gt; threshold</option>
          <option value="min_lt">min &lt; threshold</option>
        </select>
      </label>
    </div>

    <div class="form-row narrow">
      <label class="field">
        <span>Threshold</span>
        <input type="number" bind:value={nr.threshold} step="0.1" />
      </label>
      <label class="field wide">
        <span>Alert Message</span>
        <input bind:value={nr.message} placeholder="Temperature exceeds safe limit" />
      </label>
    </div>

    <div class="form-actions">
      <button class="btn-submit" onclick={submit} disabled={submitting}>
        {submitting ? 'Creating…' : 'Create Rule'}
      </button>
    </div>
  </div>
{/if}

{#if $rules.length === 0 && !show}
  <div class="empty-state">
    <div class="empty-icon">
      <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <line x1="4" y1="21" x2="4" y2="14"/><line x1="4" y1="10" x2="4" y2="3"/>
        <line x1="12" y1="21" x2="12" y2="12"/><line x1="12" y1="8" x2="12" y2="3"/>
        <line x1="20" y1="21" x2="20" y2="16"/><line x1="20" y1="12" x2="20" y2="3"/>
        <line x1="1" y1="14" x2="7" y2="14"/><line x1="9" y1="8" x2="15" y2="8"/>
        <line x1="17" y1="16" x2="23" y2="16"/>
      </svg>
    </div>
    <div class="empty-title">No rules configured</div>
    <div class="empty-sub">Add a threshold rule to trigger alerts when sensor values exceed limits.</div>
  </div>
{:else}
  <div class="rule-list">
    {#each $rules as r (r.id)}
      <div class="ri" class:disabled={!r.enabled}>
        <div class="ri-header">
          <div class="ri-id-row">
            <span class="ri-id">{r.id}</span>
            <span class="ri-badge" class:active={r.enabled}>{r.enabled ? 'ACTIVE' : 'OFF'}</span>
          </div>
          <div class="ri-key-badge">{r.key}</div>
        </div>
        <div class="ri-rule">
          <span class="ri-cond">{conditionLabels[r.condition] ?? r.condition}</span>
          <span class="ri-thresh">{r.threshold}</span>
        </div>
        <div class="ri-msg">{r.message}</div>
      </div>
    {/each}
  </div>
{/if}

<style>
  /* ── HEADER ───────────────────────────────────────────────────── */
  .ph { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1.75rem; }
  .ph-left { display: flex; align-items: center; gap: .75rem; }
  .ph-left h1 { font-size: 1.35rem; font-weight: 700; }
  .rule-count { font-family: 'IBM Plex Mono', monospace; font-size: .72rem; color: var(--dim); }
  .btn {
    display: inline-flex; align-items: center; gap: .4rem;
    font-family: 'IBM Plex Mono', monospace; font-size: .78rem; font-weight: 600;
    padding: .45rem .875rem;
    border-radius: 7px; cursor: pointer; transition: all .15s;
  }
  .btn.add {
    background: var(--ac3); color: var(--ac);
    border: 1px solid var(--ac4);
  }
  .btn.add:hover { background: var(--ac4); }
  .btn.cancel {
    background: none; color: var(--dim);
    border: 1px solid var(--bd);
  }
  .btn.cancel:hover { color: var(--tx); border-color: var(--bd2); }

  /* ── FORM PANEL ──────────────────────────────────────────────── */
  .form-panel {
    background: var(--sf); border: 1px solid var(--bd);
    border-radius: 12px; padding: 1.375rem;
    margin-bottom: 1.5rem;
    display: flex; flex-direction: column; gap: 1rem;
  }
  .form-title { font-size: .82rem; font-weight: 700; color: var(--tx2); }
  .form-err {
    display: flex; align-items: center; gap: .4rem;
    color: var(--rd); font-size: .78rem;
    padding: .45rem .75rem;
    background: rgba(255,87,87,.07);
    border: 1px solid rgba(255,87,87,.18);
    border-radius: 7px;
  }
  .form-row {
    display: grid; grid-template-columns: repeat(3, 1fr); gap: .875rem;
  }
  .form-row.narrow {
    grid-template-columns: 1fr 2fr;
  }
  .field { display: flex; flex-direction: column; gap: .28rem; }
  .field.wide { grid-column: span 1; }
  .field span {
    font-size: .7rem; font-weight: 600;
    color: var(--dim); letter-spacing: .3px;
  }
  .field input, .field select {
    font-family: 'IBM Plex Mono', monospace; font-size: .845rem;
    padding: .55rem .75rem;
    background: var(--bg); color: var(--tx);
    border: 1px solid var(--bd); border-radius: 7px;
    outline: none; transition: border-color .15s, box-shadow .15s;
  }
  .field input:focus, .field select:focus {
    border-color: var(--ac); box-shadow: 0 0 0 3px var(--ac3);
  }
  .form-actions { display: flex; }
  .btn-submit {
    font-family: 'IBM Plex Mono', monospace; font-size: .83rem; font-weight: 600;
    padding: .6rem 1.375rem;
    background: var(--ac); color: #030a0f;
    border: none; border-radius: 7px; cursor: pointer;
    transition: background .15s, transform .1s;
    letter-spacing: .3px;
  }
  .btn-submit:hover:not(:disabled) { background: var(--ac2); }
  .btn-submit:active:not(:disabled) { transform: scale(.98); }
  .btn-submit:disabled { opacity: .6; cursor: not-allowed; }

  /* ── EMPTY STATE ─────────────────────────────────────────────── */
  .empty-state {
    display: flex; flex-direction: column; align-items: center; gap: .5rem;
    padding: 4rem 2rem;
    background: var(--sf); border: 1px solid var(--bd); border-radius: 12px;
    text-align: center;
  }
  .empty-icon {
    width: 52px; height: 52px; border-radius: 50%;
    background: var(--sf2); border: 1px solid var(--bd);
    display: flex; align-items: center; justify-content: center;
    color: var(--dim); margin-bottom: .25rem;
  }
  .empty-title { font-size: .95rem; font-weight: 700; color: var(--tx2); }
  .empty-sub { font-size: .8rem; color: var(--dim); max-width: 360px; line-height: 1.6; }

  /* ── RULE LIST ───────────────────────────────────────────────── */
  .rule-list { display: flex; flex-direction: column; gap: .5rem; }
  .ri {
    background: var(--sf); border: 1px solid var(--bd);
    border-radius: 10px; padding: 1rem 1.125rem;
    display: flex; flex-direction: column; gap: .375rem;
    transition: border-color .15s;
  }
  .ri:hover { border-color: var(--bd2); }
  .ri.disabled { opacity: .4; }
  .ri-header { display: flex; align-items: center; justify-content: space-between; }
  .ri-id-row { display: flex; align-items: center; gap: .5rem; }
  .ri-id {
    font-family: 'IBM Plex Mono', monospace; font-size: .875rem; font-weight: 700;
    color: var(--tx);
  }
  .ri-badge {
    font-family: 'IBM Plex Mono', monospace; font-size: .58rem; font-weight: 700;
    letter-spacing: .5px;
    padding: .12rem .42rem; border-radius: 4px;
    background: var(--sf2); color: var(--dim);
  }
  .ri-badge.active { background: rgba(0,217,184,.1); color: var(--ac); }
  .ri-key-badge {
    font-family: 'IBM Plex Mono', monospace; font-size: .7rem;
    background: var(--sf2); border: 1px solid var(--bd);
    padding: .1rem .42rem; border-radius: 5px; color: var(--dim);
  }
  .ri-rule {
    display: flex; align-items: baseline; gap: .4rem;
    font-family: 'IBM Plex Mono', monospace; font-size: .82rem;
  }
  .ri-cond { color: var(--dim); }
  .ri-thresh { color: var(--ac); font-weight: 600; }
  .ri-msg { font-size: .8rem; color: var(--dim); }
</style>
