<script setup lang="ts">
</script>

<template>
  <div class="min-h-screen bg-gray-50 text-gray-900">
    <!-- Navigation -->
    <nav class="bg-white border-b border-gray-200 px-4 py-3">
      <div class="max-w-4xl mx-auto flex items-center justify-between">
        <router-link to="/" class="text-xl font-bold text-indigo-600">Become</router-link>
        <div class="flex gap-4 text-sm">
          <router-link to="/" class="hover:text-indigo-600" active-class="text-indigo-600 font-semibold">Today</router-link>
          <router-link to="/memorize" class="hover:text-indigo-600" active-class="text-indigo-600 font-semibold">Memorize</router-link>
          <router-link to="/practices" class="hover:text-indigo-600" active-class="text-indigo-600 font-semibold">Practices</router-link>
          <router-link to="/reports" class="hover:text-indigo-600" active-class="text-indigo-600 font-semibold">Reports</router-link>
          <router-link to="/pillars" class="hover:text-indigo-600" active-class="text-indigo-600 font-semibold">Pillars</router-link>
          <router-link to="/notes" class="hover:text-indigo-600" active-class="text-indigo-600 font-semibold">Notes</router-link>
          <router-link to="/reflections" class="hover:text-indigo-600" active-class="text-indigo-600 font-semibold">Reflect</router-link>
          <router-link to="/tasks" class="hover:text-indigo-600" active-class="text-indigo-600 font-semibold">Tasks</router-link>
        </div>
      </div>
    </nav>

    <!-- Content -->
    <main class="max-w-4xl mx-auto px-4 py-6">
      <router-view />
    </main>
  </div>
</template>
