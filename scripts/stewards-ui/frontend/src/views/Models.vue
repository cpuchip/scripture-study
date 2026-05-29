<script setup lang="ts">
// Models catalog — browse + copy-paste view for the substrate's
// model_pricing table joined with live providers_loaded(). Backs the
// 2026-05-29 "listable in settings" half of Michael's ask.
//
// Read-only by design. Model registration happens via SQL migrations
// (see extension/4a-cost-tracking.sql and follow-on price updates).
import { ref, computed, onMounted } from 'vue'
import { api, type ModelRow, type ProviderRow } from '@/api'

const models = ref<ModelRow[]>([])
const providers = ref<ProviderRow[]>([])
const loading = ref(true)
const error = ref('')
const filter = ref('')

const formatUSD = (microPerMtok: number) => {
  // input/output are in micro-USD per million tokens; convert to $/Mtok
  const dollarsPerMtok = microPerMtok / 1_000_000
  return '$' + dollarsPerMtok.toFixed(2)
}

const groupedByProvider = computed(() => {
  const grouped: Record<string, ModelRow[]> = {}
  const f = filter.value.trim().toLowerCase()
  for (const m of models.value) {
    if (f && !m.provider.toLowerCase().includes(f) && !m.model.toLowerCase().includes(f)) {
      continue
    }
    const bucket = grouped[m.provider] ?? (grouped[m.provider] = [])
    bucket.push(m)
  }
  return grouped
})

const providerInfo = computed(() => {
  const info: Record<string, ProviderRow> = {}
  for (const p of providers.value) {
    info[p.name] = p
  }
  return info
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [m, p] = await Promise.all([
      api.modelsList(),
      api.providers(),
    ])
    models.value = m.items
    providers.value = p.items
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

async function copyToClipboard(text: string, ev: MouseEvent) {
  await navigator.clipboard.writeText(text)
  // Brief visual feedback
  const el = ev.currentTarget as HTMLElement
  const orig = el.textContent
  el.textContent = '✓ copied'
  setTimeout(() => { if (el.textContent === '✓ copied') el.textContent = orig }, 800)
}

onMounted(load)
</script>

<template>
  <div class="models-view">
    <header class="page-header">
      <h1>Models catalog</h1>
      <p class="subtitle">
        Models known to the substrate via <code>stewards.model_pricing</code>, grouped by provider.
        Click any model name to copy it for pasting into the Brainstorm form or stewards-cli.
      </p>
    </header>

    <div class="controls">
      <input
        type="text"
        v-model="filter"
        placeholder="Filter by provider or model name…"
        class="filter-input"
      />
      <span class="count">{{ models.length }} models · {{ Object.keys(providerInfo).length }} providers loaded</span>
    </div>

    <div v-if="loading" class="status">Loading…</div>
    <div v-else-if="error" class="error">{{ error }}</div>

    <div v-else>
      <section
        v-for="(provModels, provName) in groupedByProvider"
        :key="provName"
        class="provider-section"
      >
        <header class="provider-header">
          <h2>
            <span class="prov-name">{{ provName }}</span>
            <span v-if="providerInfo[provName]" class="prov-meta">
              {{ providerInfo[provName].kind }} ·
              {{ providerInfo[provName].has_api_key ? 'key configured' : 'no key' }}
            </span>
            <span v-else class="prov-meta warn">
              not currently loaded
            </span>
          </h2>
          <span v-if="providerInfo[provName]?.default_model" class="prov-default">
            default: <code>{{ providerInfo[provName].default_model }}</code>
          </span>
        </header>

        <table class="model-table">
          <thead>
            <tr>
              <th>Model</th>
              <th class="num">Input $/Mtok</th>
              <th class="num">Output $/Mtok</th>
              <th class="num">Cache write</th>
              <th class="num">Cache read</th>
              <th>Notes</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="m in provModels" :key="m.provider + ':' + m.model">
              <td>
                <button
                  class="model-name-copy"
                  :class="{ 'is-default': m.is_provider_default }"
                  @click="copyToClipboard(m.model, $event)"
                  :title="`Click to copy: ${m.model}`"
                >
                  {{ m.model }}
                  <span v-if="m.is_provider_default" class="default-tag">★</span>
                </button>
              </td>
              <td class="num">{{ formatUSD(m.input_micro_per_mtok) }}</td>
              <td class="num">{{ formatUSD(m.output_micro_per_mtok) }}</td>
              <td class="num">{{ m.cache_write_micro_per_mtok != null ? formatUSD(m.cache_write_micro_per_mtok) : '—' }}</td>
              <td class="num">{{ m.cache_read_micro_per_mtok != null ? formatUSD(m.cache_read_micro_per_mtok) : '—' }}</td>
              <td class="notes">{{ m.notes || '' }}</td>
            </tr>
          </tbody>
        </table>
      </section>
    </div>
  </div>
</template>

<style scoped>
.models-view {
  max-width: 1100px;
  margin: 0 auto;
  padding: 1rem 2rem 4rem;
}

.page-header h1 {
  margin: 0 0 0.25rem;
}
.subtitle {
  color: var(--text-muted, #888);
  margin: 0 0 1.5rem;
}
.subtitle code {
  background: var(--surface, #1a1a1a);
  padding: 1px 5px;
  border-radius: 3px;
  font-size: 0.85em;
}

.controls {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.5rem;
}
.filter-input {
  flex: 1;
  max-width: 400px;
  padding: 0.4rem 0.6rem;
  border: 1px solid var(--border, #444);
  border-radius: 4px;
  background: var(--surface, #1a1a1a);
  color: inherit;
}
.count {
  font-size: 0.85rem;
  color: var(--text-muted, #888);
}

.provider-section {
  margin-bottom: 2rem;
  border: 1px solid var(--border, #444);
  border-radius: 6px;
  overflow: hidden;
}
.provider-header {
  background: var(--surface, #1a1a1a);
  padding: 0.7rem 1rem;
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  border-bottom: 1px solid var(--border, #444);
}
.provider-header h2 {
  margin: 0;
  font-size: 1.05rem;
  display: flex;
  align-items: baseline;
  gap: 0.8rem;
}
.prov-name {
  font-family: var(--font-mono, monospace);
  font-weight: 600;
}
.prov-meta {
  font-size: 0.8rem;
  font-weight: normal;
  color: var(--text-muted, #888);
}
.prov-meta.warn {
  color: #e67e22;
}
.prov-default {
  font-size: 0.85rem;
  color: var(--text-muted, #888);
}
.prov-default code {
  background: transparent;
  font-family: var(--font-mono, monospace);
  color: var(--text, #ddd);
}

.model-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}
.model-table th,
.model-table td {
  padding: 0.5rem 0.8rem;
  text-align: left;
  border-bottom: 1px solid var(--border-subtle, #2a2a2a);
}
.model-table th {
  background: var(--surface-alt, #161616);
  color: var(--text-muted, #888);
  font-weight: 500;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}
.model-table .num {
  text-align: right;
  font-family: var(--font-mono, monospace);
}
.model-table tr:last-child td { border-bottom: none; }

.model-name-copy {
  background: transparent;
  border: 1px dashed transparent;
  color: inherit;
  font-family: var(--font-mono, monospace);
  font-size: 0.95rem;
  cursor: pointer;
  padding: 0.15rem 0.4rem;
  border-radius: 3px;
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
}
.model-name-copy:hover {
  border-color: var(--border, #444);
  background: var(--surface-hover, #2a2a2a);
}
.model-name-copy.is-default {
  font-weight: 600;
}
.default-tag {
  color: #f1c40f;
  font-size: 0.8em;
}

.notes {
  color: var(--text-muted, #888);
  font-size: 0.85rem;
}

.status {
  padding: 2rem;
  color: var(--text-muted, #888);
  text-align: center;
}
.error {
  padding: 0.75rem 1rem;
  background: rgba(192, 57, 43, 0.15);
  border: 1px solid #c0392b;
  border-radius: 4px;
  color: #ff6b5b;
}
</style>
