<script setup lang="ts">
// Compact churchofjesuschrist.org iframe panel. Pattern borrowed from
// projects/cpuchip.net/src/components/ScripturePanel.vue + scripts/becoming/
// frontend/src/components/ReferencePanel.vue.
//
// Per the workspace convention, scripture TEXT is never bundled or
// proxied — only the church's official site renders it. We layer
// our 1828 lens BESIDE the iframe, not on top of it.

defineProps<{
  reference: string  // e.g. "D&C 84:33-44" or "1 Corinthians 13"
  url: string        // the prepared churchofjesuschrist.org URL
}>()
</script>

<template>
  <aside class="def-card overflow-hidden flex flex-col" style="min-height: 480px">
    <header class="px-4 py-3 border-b border-stone-200 bg-stone-50 flex items-center justify-between">
      <div>
        <h3 class="font-serif text-base">{{ reference }}</h3>
        <p class="text-xs text-stone-500 mt-0.5">via churchofjesuschrist.org</p>
      </div>
      <a :href="url" target="_blank" rel="noopener" class="text-xs text-amber-700 hover:underline">Open ↗</a>
    </header>
    <iframe
      :src="url"
      class="flex-1 w-full border-0 bg-white"
      :title="reference"
      loading="lazy"
      sandbox="allow-same-origin allow-scripts allow-popups allow-forms allow-popups-to-escape-sandbox"
    ></iframe>
  </aside>
</template>
