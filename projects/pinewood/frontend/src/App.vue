<script setup>
import { RouterView, RouterLink, useRoute } from 'vue-router'
import { computed } from 'vue'

const route = useRoute()
const raceID = computed(() => route.params.id)
const showNav = computed(() => !!raceID.value && !route.path.endsWith('/display'))
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <header v-if="showNav" class="bg-white border-b border-slate-200 px-6 py-3 flex items-center gap-4 print:hidden">
      <RouterLink to="/" class="font-bold text-lg">🏁 Pinewood</RouterLink>
      <nav class="flex gap-2 ml-4">
        <RouterLink :to="`/race/${raceID}/registration`" class="px-3 py-1 rounded hover:bg-slate-100">Cars</RouterLink>
        <RouterLink :to="`/race/${raceID}/schedule`" class="px-3 py-1 rounded hover:bg-slate-100">Schedule</RouterLink>
        <RouterLink :to="`/race/${raceID}/score`" class="px-3 py-1 rounded hover:bg-slate-100">Score</RouterLink>
        <RouterLink :to="`/race/${raceID}/display`" class="px-3 py-1 rounded hover:bg-slate-100">TV</RouterLink>
        <RouterLink :to="`/race/${raceID}/results`" class="px-3 py-1 rounded hover:bg-slate-100">Results</RouterLink>
      </nav>
    </header>
    <main class="flex-1">
      <RouterView />
    </main>
  </div>
</template>
