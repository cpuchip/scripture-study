<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <header class="border-b border-stone-300 bg-[var(--paper-2)]">
      <div class="max-w-5xl mx-auto px-6 py-4 flex items-baseline justify-between">
        <RouterLink to="/" class="flex items-baseline gap-2 group">
          <span class="text-xl font-serif font-semibold tracking-tight">1828 Illuminated</span>
          <span class="text-xs text-stone-500 hidden sm:inline">— scripture in its Restoration-era language frame</span>
        </RouterLink>
        <nav class="flex gap-5 text-sm">
          <RouterLink to="/word" class="text-stone-700 hover:text-stone-900" active-class="text-amber-700 font-medium">Word Search</RouterLink>
          <RouterLink to="/verse" class="text-stone-700 hover:text-stone-900" active-class="text-amber-700 font-medium">Verse Explorer</RouterLink>
          <RouterLink to="/about" class="text-stone-700 hover:text-stone-900" active-class="text-amber-700 font-medium">About</RouterLink>
        </nav>
      </div>
    </header>

    <main class="flex-1">
      <RouterView />
    </main>

    <footer class="border-t border-stone-300 bg-[var(--paper-2)] mt-12">
      <div class="max-w-5xl mx-auto px-6 py-6 text-xs text-stone-600 flex flex-wrap items-baseline justify-between gap-2">
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
