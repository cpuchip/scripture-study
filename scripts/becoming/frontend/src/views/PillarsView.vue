<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api, type Pillar, type Practice, type PillarLink } from '../api'

const pillars = ref<Pillar[]>([])
const practices = ref<Practice[]>([])
const loading = ref(true)
const expandedPillar = ref<number | null>(null)

// Create / edit state
const showForm = ref(false)
const editingPillar = ref<Pillar | null>(null)
const formName = ref('')
const formDescription = ref('')
const formIcon = ref('')
const formParentId = ref<number | null>(null)

// Link practice state
const linkingPillarId = ref<number | null>(null)
const selectedPracticeId = ref<number | null>(null)

// Onboarding
const showOnboarding = ref(false)
const suggestions = ref<Pillar[]>([])
const selectedSuggestions = ref<Set<number>>(new Set())

// Practice pillar mapping cache: pillar_id -> practice[]
const pillarPractices = ref<Record<number, Practice[]>>({})

const topLevelPillars = computed(() => pillars.value.filter(p => !p.parent_id))

const unlinkedPractices = computed(() => {
  const linkedIds = new Set<number>()
  for (const pList of Object.values(pillarPractices.value)) {
    for (const p of pList) linkedIds.add(p.id)
  }
  return practices.value.filter(p => !linkedIds.has(p.id))
})

function getSubPillars(parentId: number): Pillar[] {
  return pillars.value.filter(p => p.parent_id === parentId)
}

function getPillarPracticesList(pillarId: number): Practice[] {
  return pillarPractices.value[pillarId] || []
}

// Compute completion % per pillar (based on linked practices count and how many are active)
function pillarCompletionPct(pillarId: number): number {
  const pracs = getPillarPracticesList(pillarId)
  const subPillars = getSubPillars(pillarId)
  let allPracs = [...pracs]
  for (const sp of subPillars) {
    allPracs = allPracs.concat(getPillarPracticesList(sp.id))
  }
  if (allPracs.length === 0) return 0
  const active = allPracs.filter(p => p.active).length
  return Math.round((active / allPracs.length) * 100)
}

function pillarTrend(pillarId: number): string {
  const pct = pillarCompletionPct(pillarId)
  if (pct >= 70) return '▲'
  if (pct >= 40) return '→'
  return '▼'
}

function pillarTrendColor(pillarId: number): string {
  const pct = pillarCompletionPct(pillarId)
  if (pct >= 70) return 'text-green-600'
  if (pct >= 40) return 'text-yellow-600'
  return 'text-red-500'
}

async function load() {
  loading.value = true
  try {
    const [pillarsData, practicesData, has] = await Promise.all([
      api.listPillarsFlat(),
      api.listPractices(),
      api.hasPillars(),
    ])
    pillars.value = pillarsData
    practices.value = practicesData

    // Show onboarding if no pillars exist
    if (!has.has_pillars) {
      const s = await api.getPillarSuggestions()
      suggestions.value = s
      selectedSuggestions.value = new Set(s.map((_, i) => i))
      showOnboarding.value = true
      loading.value = false
      return
    }

    // Build mapping: for each practice, get its pillars, then reverse-map
    const mapping: Record<number, Practice[]> = {}
    for (const practice of practicesData) {
      try {
        const links: PillarLink[] = await api.getPracticePillars(practice.id)
        for (const link of links) {
          if (!mapping[link.pillar_id]) mapping[link.pillar_id] = []
          mapping[link.pillar_id]!.push(practice)
        }
      } catch { /* noop */ }
    }
    pillarPractices.value = mapping
  } finally {
    loading.value = false
  }
}

async function acceptOnboarding() {
  for (const [idx, sug] of suggestions.value.entries()) {
    if (selectedSuggestions.value.has(idx)) {
      await api.createPillar({ name: sug.name, description: sug.description, icon: sug.icon })
    }
  }
  showOnboarding.value = false
  await load()
}

async function skipOnboarding() {
  showOnboarding.value = false
}

function toggleExpand(id: number) {
  expandedPillar.value = expandedPillar.value === id ? null : id
}

function startCreate(parentId: number | null = null) {
  editingPillar.value = null
  formName.value = ''
  formDescription.value = ''
  formIcon.value = ''
  formParentId.value = parentId
  showForm.value = true
}

function startEdit(pillar: Pillar) {
  editingPillar.value = pillar
  formName.value = pillar.name
  formDescription.value = pillar.description || ''
  formIcon.value = pillar.icon || ''
  formParentId.value = pillar.parent_id || null
  showForm.value = true
}

async function saveForm() {
  const data: Partial<Pillar> = {
    name: formName.value.trim(),
    description: formDescription.value.trim(),
    icon: formIcon.value.trim(),
    parent_id: formParentId.value,
  }
  if (!data.name) return

  if (editingPillar.value) {
    await api.updatePillar(editingPillar.value.id, data)
  } else {
    await api.createPillar(data)
  }
  showForm.value = false
  await load()
}

async function deletePillar(id: number) {
  if (!confirm('Delete this pillar? Practices won\'t be deleted, only unlinked.')) return
  await api.deletePillar(id)
  await load()
}

function startLinkPractice(pillarId: number) {
  linkingPillarId.value = pillarId
  selectedPracticeId.value = null
}

async function linkPractice() {
  if (!linkingPillarId.value || !selectedPracticeId.value) return
  // Get current pillars for this practice, add the new one
  const existing = await api.getPracticePillars(selectedPracticeId.value)
  const pillarIds = [...existing.map(l => l.pillar_id), linkingPillarId.value]
  await api.setPracticePillars(selectedPracticeId.value, pillarIds)
  linkingPillarId.value = null
  selectedPracticeId.value = null
  await load()
}

async function unlinkPractice(practiceId: number, pillarId: number) {
  const existing = await api.getPracticePillars(practiceId)
  const pillarIds = existing.map(l => l.pillar_id).filter(id => id !== pillarId)
  await api.setPracticePillars(practiceId, pillarIds)
  await load()
}

function practiceTypeLabel(type: string): string {
  const labels: Record<string, string> = {
    memorize: 'memorize', tracker: 'tracker', habit: 'habit',
    task: 'task', scheduled: 'scheduled',
  }
  return labels[type] || type
}

onMounted(load)
</script>

<template>
  <div>
    <!-- Header -->
    <div class="flex items-center justify-between mb-2">
      <h1 class="text-2xl font-bold">Pillars of Growth</h1>
      <button
        @click="startCreate()"
        class="px-3 py-1.5 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700"
      >
        + New Pillar
      </button>
    </div>
    <p class="text-gray-500 text-sm italic mb-6">
      "Jesus increased in wisdom and stature, and in favour with God and man" — Luke 2:52
    </p>

    <!-- Loading -->
    <div v-if="loading" class="text-center text-gray-400 py-12">Loading pillars…</div>

    <div v-else>
      <!-- Onboarding -->
      <div v-if="showOnboarding" class="text-center py-8">
        <h2 class="text-xl font-bold mb-2">Set Up Your Pillars</h2>
        <p class="text-gray-500 mb-6 max-w-md mx-auto">
          Based on Luke 2:52, the Savior grew in four areas. Select which pillars you'd like to start with — you can always add more later.
        </p>
        <div class="grid grid-cols-2 gap-4 max-w-lg mx-auto mb-8">
          <button
            v-for="(sug, idx) in suggestions"
            :key="idx"
            @click="selectedSuggestions.has(idx) ? selectedSuggestions.delete(idx) : selectedSuggestions.add(idx)"
            class="p-4 rounded-xl border-2 text-left transition-all"
            :class="selectedSuggestions.has(idx)
              ? 'border-indigo-400 bg-indigo-50 shadow-md'
              : 'border-gray-200 bg-white hover:border-gray-300'"
          >
            <span class="text-3xl block mb-2">{{ sug.icon }}</span>
            <span class="font-semibold">{{ sug.name }}</span>
            <p class="text-xs text-gray-500 mt-1">{{ sug.description }}</p>
          </button>
        </div>
        <div class="flex justify-center gap-3">
          <button @click="skipOnboarding"
                  class="px-4 py-2 text-gray-500 hover:text-gray-700 text-sm">
            Skip for now
          </button>
          <button @click="acceptOnboarding"
                  :disabled="selectedSuggestions.size === 0"
                  class="px-6 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm disabled:opacity-40">
            Create {{ selectedSuggestions.size }} Pillar{{ selectedSuggestions.size !== 1 ? 's' : '' }}
          </button>
        </div>
      </div>

      <!-- Empty state (after skip) -->
      <div v-else-if="pillars.length === 0" class="text-center py-12">
        <p class="text-gray-500 mb-4">No pillars yet. Start by adding pillars to organize your growth.</p>
        <button
          @click="startCreate()"
          class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700"
        >
          Create First Pillar
        </button>
      </div>

      <!-- Pillar cards -->
      <div v-else class="space-y-4">
        <div
          v-for="pillar in topLevelPillars"
          :key="pillar.id"
          class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden"
        >
          <!-- Pillar header -->
          <div
            class="flex items-center justify-between px-5 py-4 cursor-pointer hover:bg-gray-50"
            @click="toggleExpand(pillar.id)"
          >
            <div class="flex items-center gap-3">
              <span class="text-2xl">{{ pillar.icon || '◆' }}</span>
              <div>
                <h2 class="text-lg font-semibold">{{ pillar.name }}</h2>
                <p v-if="pillar.description" class="text-sm text-gray-500">{{ pillar.description }}</p>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <span class="text-sm font-medium text-gray-600">
                {{ pillarCompletionPct(pillar.id) }}%
              </span>
              <span :class="pillarTrendColor(pillar.id)" class="text-lg">
                {{ pillarTrend(pillar.id) }}
              </span>
              <span class="text-gray-400 text-sm">
                {{ expandedPillar === pillar.id ? '▾' : '▸' }}
              </span>
            </div>
          </div>

          <!-- Expanded content -->
          <div v-if="expandedPillar === pillar.id" class="border-t border-gray-100 px-5 py-4">
            <!-- Direct practices (no sub-pillar) -->
            <div v-if="getPillarPracticesList(pillar.id).length > 0" class="mb-4">
              <div v-for="practice in getPillarPracticesList(pillar.id)" :key="practice.id"
                   class="flex items-center justify-between py-1.5 pl-2 group">
                <div class="flex items-center gap-2">
                  <span class="text-indigo-400">✦</span>
                  <span class="text-sm">{{ practice.name }}</span>
                  <span class="text-xs text-gray-400 bg-gray-100 px-1.5 py-0.5 rounded">
                    {{ practiceTypeLabel(practice.type) }}
                  </span>
                </div>
                <button @click.stop="unlinkPractice(practice.id, pillar.id)"
                        class="text-xs text-red-400 opacity-0 group-hover:opacity-100 hover:text-red-600">
                  unlink
                </button>
              </div>
            </div>

            <!-- Sub-pillars -->
            <div v-for="sub in getSubPillars(pillar.id)" :key="sub.id"
                 class="mb-4 border-l-2 border-gray-200 pl-4">
              <div class="flex items-center justify-between mb-1">
                <h3 class="text-sm font-semibold text-gray-700">
                  {{ sub.icon }} {{ sub.name }}
                </h3>
                <div class="flex gap-2">
                  <button @click.stop="startEdit(sub)"
                          class="text-xs text-gray-400 hover:text-indigo-600">edit</button>
                  <button @click.stop="deletePillar(sub.id)"
                          class="text-xs text-gray-400 hover:text-red-600">delete</button>
                </div>
              </div>

              <div v-for="practice in getPillarPracticesList(sub.id)" :key="practice.id"
                   class="flex items-center justify-between py-1 pl-2 group">
                <div class="flex items-center gap-2">
                  <span class="text-indigo-400">✦</span>
                  <span class="text-sm">{{ practice.name }}</span>
                  <span class="text-xs text-gray-400 bg-gray-100 px-1.5 py-0.5 rounded">
                    {{ practiceTypeLabel(practice.type) }}
                  </span>
                </div>
                <button @click.stop="unlinkPractice(practice.id, sub.id)"
                        class="text-xs text-red-400 opacity-0 group-hover:opacity-100 hover:text-red-600">
                  unlink
                </button>
              </div>

              <!-- Link practice to sub-pillar -->
              <div v-if="linkingPillarId === sub.id" class="flex gap-2 mt-2">
                <select v-model="selectedPracticeId"
                        class="text-xs border rounded px-2 py-1 flex-1">
                  <option :value="null">Select practice…</option>
                  <option v-for="p in unlinkedPractices" :key="p.id" :value="p.id">{{ p.name }}</option>
                </select>
                <button @click.stop="linkPractice"
                        class="text-xs text-indigo-600 hover:text-indigo-800">Link</button>
                <button @click.stop="linkingPillarId = null"
                        class="text-xs text-gray-400 hover:text-gray-600">Cancel</button>
              </div>
              <button v-else @click.stop="startLinkPractice(sub.id)"
                      class="text-xs text-gray-400 hover:text-indigo-600 mt-1">
                + Link a practice…
              </button>
            </div>

            <!-- Actions -->
            <div class="flex gap-4 mt-3 pt-3 border-t border-gray-100">
              <!-- Link practice to top-level -->
              <div v-if="linkingPillarId === pillar.id" class="flex gap-2 flex-1">
                <select v-model="selectedPracticeId"
                        class="text-xs border rounded px-2 py-1 flex-1">
                  <option :value="null">Select practice…</option>
                  <option v-for="p in unlinkedPractices" :key="p.id" :value="p.id">{{ p.name }}</option>
                </select>
                <button @click.stop="linkPractice"
                        class="text-xs text-indigo-600 hover:text-indigo-800">Link</button>
                <button @click.stop="linkingPillarId = null"
                        class="text-xs text-gray-400 hover:text-gray-600">Cancel</button>
              </div>
              <button v-else @click.stop="startLinkPractice(pillar.id)"
                      class="text-xs text-gray-400 hover:text-indigo-600">
                + Link a practice…
              </button>
              <button @click.stop="startCreate(pillar.id)"
                      class="text-xs text-gray-400 hover:text-indigo-600">
                + Sub-pillar
              </button>
              <button @click.stop="startEdit(pillar)"
                      class="text-xs text-gray-400 hover:text-indigo-600">
                Edit
              </button>
              <button @click.stop="deletePillar(pillar.id)"
                      class="text-xs text-gray-400 hover:text-red-600">
                Delete
              </button>
            </div>
          </div>
        </div>

        <!-- Balance Bar -->
        <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5 mt-6">
          <h3 class="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-3">Balance</h3>
          <div class="space-y-2">
            <div v-for="pillar in topLevelPillars" :key="'bar-' + pillar.id"
                 class="flex items-center gap-3">
              <span class="text-lg w-8 text-center">{{ pillar.icon || '◆' }}</span>
              <div class="flex-1 bg-gray-100 rounded-full h-4 overflow-hidden">
                <div
                  class="h-full rounded-full transition-all duration-500"
                  :class="pillarCompletionPct(pillar.id) >= 70 ? 'bg-green-500' :
                           pillarCompletionPct(pillar.id) >= 40 ? 'bg-yellow-500' : 'bg-red-400'"
                  :style="{ width: pillarCompletionPct(pillar.id) + '%' }"
                ></div>
              </div>
              <span class="text-sm text-gray-600 w-10 text-right">
                {{ pillarCompletionPct(pillar.id) }}%
              </span>
              <span v-if="pillarCompletionPct(pillar.id) < 40"
                    class="text-xs text-red-400">← needs attention</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showForm" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50"
         @click.self="showForm = false">
      <div class="bg-white rounded-xl shadow-lg p-6 w-full max-w-md">
        <h2 class="text-lg font-bold mb-4">
          {{ editingPillar ? 'Edit Pillar' : (formParentId ? 'New Sub-Pillar' : 'New Pillar') }}
        </h2>
        <div class="space-y-3">
          <div>
            <label class="text-sm font-medium text-gray-700">Icon (emoji)</label>
            <input v-model="formIcon" placeholder="🕊️"
                   class="w-full border rounded-lg px-3 py-2 text-2xl text-center" />
          </div>
          <div>
            <label class="text-sm font-medium text-gray-700">Name</label>
            <input v-model="formName" placeholder="Spiritual"
                   class="w-full border rounded-lg px-3 py-2" />
          </div>
          <div>
            <label class="text-sm font-medium text-gray-700">Description</label>
            <textarea v-model="formDescription" rows="2" placeholder="Optional description"
                      class="w-full border rounded-lg px-3 py-2"></textarea>
          </div>
        </div>
        <div class="flex justify-end gap-2 mt-4">
          <button @click="showForm = false"
                  class="px-3 py-1.5 text-gray-600 hover:text-gray-800 text-sm">Cancel</button>
          <button @click="saveForm"
                  class="px-4 py-1.5 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700">
            {{ editingPillar ? 'Save' : 'Create' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
