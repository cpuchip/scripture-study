<template>
  <div class="min-h-screen bg-gradient-to-b from-orange-50 to-white">
    <!-- Header -->
    <header class="px-6 py-4 flex items-center justify-between max-w-5xl mx-auto">
      <div class="flex items-center gap-2">
        <img src="/ibecome-icon.png" alt="Become" class="h-8 w-8" />
        <span class="text-2xl font-bold text-orange-600">{{ siteName }}</span>
      </div>
      <div class="flex items-center gap-4">
        <router-link to="/privacy" class="text-sm text-gray-500 hover:text-gray-700">Privacy</router-link>
        <router-link to="/terms" class="text-sm text-gray-500 hover:text-gray-700">Terms</router-link>
        <router-link to="/login" class="text-sm font-medium text-orange-600 hover:text-orange-700">Sign In</router-link>
      </div>
    </header>

    <!-- Hero -->
    <section class="px-6 py-20 text-center max-w-3xl mx-auto">
      <img src="/ibecome-logo.png" alt="Become" class="h-28 mx-auto mb-8" />
      <h1 class="text-4xl font-bold text-gray-900 mb-4">
        Becoming Who You're Meant to Be
      </h1>
      <p class="text-lg text-gray-600 mb-8 max-w-2xl mx-auto">
        A personal growth companion for tracking daily practices, memorizing scripture,
        reflecting on your journey, and aligning your life with what matters most.
      </p>
      <div class="flex items-center justify-center gap-4">
        <router-link
          to="/register"
          class="px-6 py-3 bg-orange-600 text-white rounded-lg font-semibold hover:bg-orange-700 transition-colors"
        >
          Get Started Free
        </router-link>
        <router-link
          to="/login"
          class="px-6 py-3 border border-gray-300 text-gray-700 rounded-lg font-semibold hover:border-gray-400 transition-colors"
        >
          Sign In
        </router-link>
      </div>
    </section>

    <!-- Features -->
    <section class="px-6 py-16 max-w-4xl mx-auto">
      <div class="grid md:grid-cols-3 gap-8">
        <div class="text-center">
          <div class="text-3xl mb-3">📖</div>
          <h3 class="font-semibold text-gray-900 mb-2">Scripture Memorization</h3>
          <p class="text-sm text-gray-600">
            Memorize scripture with spaced repetition. Type-to-verify quizzes help lock verses into long-term memory.
          </p>
        </div>
        <div class="text-center">
          <div class="text-3xl mb-3">🎯</div>
          <h3 class="font-semibold text-gray-900 mb-2">Daily Practices</h3>
          <p class="text-sm text-gray-600">
            Track habits, scheduled activities, and personal goals. See your progress at a glance each day.
          </p>
        </div>
        <div class="text-center">
          <div class="text-3xl mb-3">✍️</div>
          <h3 class="font-semibold text-gray-900 mb-2">Reflection & Journaling</h3>
          <p class="text-sm text-gray-600">
            Daily reflection prompts and a private journal to document your growth and insights.
          </p>
        </div>
      </div>
    </section>

    <!-- About -->
    <section class="px-6 py-16 bg-white">
      <div class="max-w-3xl mx-auto text-center">
        <h2 class="text-2xl font-bold text-gray-900 mb-4">About</h2>
        <p class="text-gray-600 mb-4">
          {{ siteName }} is a free personal development tool built for people of faith who want to be intentional
          about their daily growth. Track practices, memorize scripture, reflect on your journey, and organize
          your life around the pillars that matter most to you.
        </p>
        <p class="text-sm text-gray-500">
          This app is not affiliated with, endorsed by, or officially connected to The Church of Jesus
          Christ of Latter-day Saints. It's a personal project built to help individuals in their own
          spiritual journey.
        </p>
      </div>
    </section>

    <!-- Footer -->
    <footer class="px-6 py-8 border-t border-gray-200">
      <div class="max-w-4xl mx-auto flex flex-col md:flex-row items-center justify-between gap-4">
        <div class="flex items-center gap-2 text-sm text-gray-500">
          <img src="/ibecome-icon.png" alt="" class="h-5 w-5" />
          <span>&copy; {{ new Date().getFullYear() }} {{ siteName }}</span>
        </div>
        <div class="flex items-center gap-6 text-sm">
          <router-link to="/privacy" class="text-gray-500 hover:text-gray-700">Privacy Policy</router-link>
          <router-link to="/terms" class="text-gray-500 hover:text-gray-700">Terms of Service</router-link>
          <a href="mailto:stuffleberryco+support@gmail.com" class="text-gray-500 hover:text-gray-700">Contact</a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const siteName = computed(() => {
  const host = window.location.hostname
  if (host.includes('webeco')) return 'We Become'
  return 'I Become'
})
</script>
