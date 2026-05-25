<script setup lang="ts">
import { RouterLink, RouterView, useRoute } from 'vue-router'
import StudyTreePanel from '@/components/StudyTreePanel.vue'
import { panelOpen, panelPinned, session } from '@/composables/useStudyTree'
import { clickMode } from '@/composables/useClickMode'
import { computed } from 'vue'

function toggleClickMode() {
  clickMode.value = clickMode.value === 'definition' ? 'scripture' : 'definition'
}

const route = useRoute()
// Hide the study tree on /present (fullscreen tablet mode) per the
// proposal D-ST-10 ratification — Present is for distraction-free reading.
const showTreePanel = computed(() => route.name !== 'present')

// When pinned + open, the panel renders INLINE as the second column of
// the page layout (not as a viewport-pinned overlay). The whole layout
// widens via max-w-7xl so the content stays at a comfortable reading
// width AND the panel sits beside it inside the same centered container —
// no cream-zone gap between viewport edge and content. The header + footer
// widen in lockstep so the visual frame stays aligned.
const isInlinePin = computed(
  () => showTreePanel.value && panelOpen.value && panelPinned.value,
)
const containerMax = computed(() =>
  isInlinePin.value ? 'max-w-7xl' : 'max-w-5xl',
)
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <header class="border-b border-stone-300 bg-[var(--paper-2)]">
      <div
        :class="[containerMax, 'mx-auto px-6 py-4 flex items-baseline justify-between transition-[max-width] duration-200']"
      >
        <RouterLink to="/" class="flex items-baseline gap-2 group">
          <span class="text-xl font-serif font-semibold tracking-tight">1828 Illuminated</span>
          <span class="text-xs text-stone-500 hidden sm:inline">— scripture in its Restoration-era language frame</span>
        </RouterLink>
        <nav class="flex gap-5 text-sm items-baseline">
          <RouterLink to="/word" class="text-stone-700 hover:text-stone-900" active-class="text-amber-700 font-medium">Word Search</RouterLink>
          <RouterLink to="/verse" class="text-stone-700 hover:text-stone-900" active-class="text-amber-700 font-medium">Verse Explorer</RouterLink>
          <RouterLink to="/dictionary" class="text-stone-700 hover:text-stone-900" active-class="text-amber-700 font-medium">Dictionary <span class="text-xs text-amber-700">·preview</span></RouterLink>
          <RouterLink to="/about" class="text-stone-700 hover:text-stone-900" active-class="text-amber-700 font-medium">About</RouterLink>
          <!-- Click-mode toggle: clicking words goes to definitions (default)
               or to scripture occurrences. Persists per browser. -->
          <button
            @click="toggleClickMode"
            class="inline-flex items-baseline gap-1 px-2 py-0.5 rounded border text-xs transition"
            :class="clickMode === 'scripture'
              ? 'border-sky-400 bg-sky-50 text-sky-800 hover:bg-sky-100'
              : 'border-stone-300 text-stone-600 hover:border-stone-400'"
            :title="clickMode === 'definition'
              ? 'Click mode: definition (showing 1828 + modern). Press to switch to scripture mode (word click finds occurrences).'
              : 'Click mode: scripture (word click finds occurrences). Press to switch back to definition mode.'"
          >
            <span>{{ clickMode === 'scripture' ? '📖' : '📚' }}</span>
            <span class="hidden sm:inline">{{ clickMode === 'scripture' ? 'scripture' : 'definition' }}</span>
          </button>
          <RouterLink to="/settings" class="text-stone-500 hover:text-stone-900 text-xs" active-class="text-amber-700 font-medium" title="LLM endpoint settings">⚙</RouterLink>
          <span v-if="session.authenticated" class="text-xs text-stone-500 font-serif border-l pl-3 border-stone-300" title="Signed in to cloud account">👤 {{ session.user?.name || session.user?.email }}</span>
          <a v-else href="https://ibeco.me/login" target="_blank" class="text-xs text-stone-500 hover:text-amber-700 transition border-l pl-3 border-stone-300">Sign In</a>
        </nav>
      </div>
    </header>

    <main class="flex-1">
      <div
        :class="[containerMax, 'mx-auto px-6 transition-[max-width] duration-200']"
      >
        <div :class="isInlinePin ? 'flex gap-6 items-start' : ''">
          <div :class="isInlinePin ? 'flex-1 min-w-0' : ''">
            <RouterView />
          </div>
          <!-- When pinned, the tree mounts HERE as the second column.
               StudyTreePanel disables its <Teleport> in pinned mode so
               it renders in-flow instead of overlaying. -->
          <aside
            v-if="isInlinePin"
            class="w-80 shrink-0 hidden lg:block"
          >
            <StudyTreePanel inline />
          </aside>
        </div>
      </div>
    </main>

    <!-- Drawer mode (un-pinned overlay) — the panel renders Teleport'd to
         body with fixed positioning. Mounted unconditionally so the
         floating "Study tree" pill toggle is available on every surface
         except /present, where showTreePanel is false. -->
    <StudyTreePanel v-if="showTreePanel && !isInlinePin" />

    <footer class="border-t border-stone-300 bg-[var(--paper-2)] mt-12">
      <div
        :class="[containerMax, 'mx-auto px-6 py-6 text-xs text-stone-600 flex flex-wrap items-baseline justify-between gap-2 transition-[max-width] duration-200']"
      >
        <div>
          1828 Illuminated · scripture text via
          <a href="https://www.churchofjesuschrist.org/study/scriptures" class="underline hover:text-stone-900" target="_blank" rel="noopener">churchofjesuschrist.org</a>
          — not bundled here. 1828 definitions from <a href="https://webstersdictionary1828.com" class="underline hover:text-stone-900" target="_blank" rel="noopener">Webster 1828</a>; modern definitions from the Free Dictionary API.
        </div>
        <div class="text-stone-500">
          1828 illuminates — it doesn't decode.
        </div>
      </div>
    </footer>
  </div>
</template>
