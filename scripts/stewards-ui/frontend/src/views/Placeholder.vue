<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>

<template>
  <div class="space-y-4">
    <h2 class="text-2xl font-semibold tracking-tight">
      {{ route.meta.title || route.name }}
    </h2>
    <div class="rounded-md border border-zinc-800 bg-zinc-900/50 p-6">
      <p class="text-sm text-zinc-300">
        This view ships in a later phase. v1 phase 1 is foundation only.
      </p>
      <p class="text-xs text-zinc-500 mt-2 font-mono">
        route: {{ route.fullPath }}
      </p>
    </div>
  </div>
</template>
