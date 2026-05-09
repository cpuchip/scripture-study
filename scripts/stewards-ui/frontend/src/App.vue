<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <header class="border-b border-zinc-800 px-6 py-3 flex items-center gap-6">
      <h1 class="text-lg font-semibold tracking-tight">stewards-ui</h1>
      <nav class="flex gap-4 text-sm text-zinc-400">
        <RouterLink to="/" class="hover:text-zinc-100">Dashboard</RouterLink>
        <RouterLink to="/studies" class="hover:text-zinc-100">Studies</RouterLink>
        <RouterLink to="/work-items" class="hover:text-zinc-100">Work items</RouterLink>
        <RouterLink to="/sessions" class="hover:text-zinc-100">Sessions</RouterLink>
        <RouterLink to="/watchman" class="hover:text-zinc-100">Watchman</RouterLink>
        <RouterLink to="/bridge" class="hover:text-zinc-100">Bridge</RouterLink>
        <RouterLink to="/graph" class="hover:text-zinc-100">Graph</RouterLink>
        <RouterLink to="/new" class="hover:text-zinc-100">New work</RouterLink>
      </nav>
    </header>

    <main class="flex-1 p-6">
      <RouterView />
    </main>

    <footer class="border-t border-zinc-800 px-6 py-2 text-xs text-zinc-500">
      pg-ai-stewards &middot; v1 phase 1 (foundation scaffold)
    </footer>
  </div>
</template>
