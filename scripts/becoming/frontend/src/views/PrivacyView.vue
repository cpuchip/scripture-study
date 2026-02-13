<script setup lang="ts">
</script>

<template>
  <div class="max-w-3xl mx-auto px-4 py-8">
    <h1 class="text-3xl font-bold mb-2">Privacy Policy</h1>
    <p class="text-sm text-gray-400 mb-8">Effective: February 13, 2026 · Last updated: February 13, 2026</p>

    <div class="prose prose-gray max-w-none space-y-6">
      <section>
        <h2 class="text-xl font-semibold mb-2">Our Commitment</h2>
        <p>
          Become (<a href="https://ibeco.me" class="text-indigo-600 hover:underline">ibeco.me</a>) is a personal growth and scripture study tracker built on principles of honesty, integrity, and respect for your agency. We believe your personal data — your reflections, your practices, your spiritual journey — is sacred and belongs to you. We will never sell, share, or monetize it.
        </p>
      </section>

      <section>
        <h2 class="text-xl font-semibold mb-2">What We Collect</h2>
        <ul class="list-disc pl-5 space-y-1">
          <li><strong>Account information:</strong> Your name (optional), email address, and a securely hashed password used for authentication.</li>
          <li><strong>Practice data:</strong> The practices, logs, reflections, notes, tasks, pillars, and memorization cards you create within the app.</li>
          <li><strong>Authentication tokens:</strong> Session cookies used to keep you signed in. These are httpOnly and secure.</li>
          <li><strong>Google account data (if you sign in with Google):</strong> Your name, email, and profile picture provided by Google OAuth. We do not access your Google contacts, calendar, files, or any other Google service data.</li>
        </ul>
      </section>

      <section>
        <h2 class="text-xl font-semibold mb-2">What We Do NOT Collect</h2>
        <ul class="list-disc pl-5 space-y-1">
          <li>We do not use analytics trackers, advertising pixels, or third-party tracking scripts.</li>
          <li>We do not collect your IP address for tracking purposes.</li>
          <li>We do not use cookies for advertising or cross-site tracking.</li>
          <li>We do not build behavioral profiles or use your data for machine learning training.</li>
        </ul>
      </section>

      <section>
        <h2 class="text-xl font-semibold mb-2">How Your Data Is Stored</h2>
        <p>
          Your data is stored in a PostgreSQL database hosted on our private server. It is not shared with third parties. All connections are encrypted via TLS (HTTPS). Passwords are hashed using bcrypt and are never stored in plain text.
        </p>
      </section>

      <section>
        <h2 class="text-xl font-semibold mb-2">Your Rights</h2>
        <ul class="list-disc pl-5 space-y-1">
          <li><strong>Access:</strong> All your data is visible to you within the app at all times.</li>
          <li><strong>Deletion:</strong> You can delete your account and all associated data at any time from the Settings page. Deletion is immediate and permanent.</li>
          <li><strong>Portability:</strong> Your data belongs to you. We are committed to providing data export capabilities.</li>
          <li><strong>Correction:</strong> You can edit or update any of your data within the app.</li>
        </ul>
      </section>

      <section>
        <h2 class="text-xl font-semibold mb-2">Data Sharing</h2>
        <p>
          <strong>We will never sell your data.</strong> We do not share your personal data with any third party for any reason, except:
        </p>
        <ul class="list-disc pl-5 space-y-1">
          <li>If required by law (e.g., a valid court order).</li>
          <li>Google OAuth: During sign-in, Google receives confirmation that you authorized the connection. No app data is sent to Google.</li>
        </ul>
      </section>

      <section>
        <h2 class="text-xl font-semibold mb-2">Children's Privacy</h2>
        <p>
          Become is not directed at children under 13. We do not knowingly collect data from children under 13. If you believe a child has created an account, please contact us and we will delete it promptly.
        </p>
      </section>

      <section>
        <h2 class="text-xl font-semibold mb-2">Changes to This Policy</h2>
        <p>
          If we make changes to this policy, we will update the "Last updated" date above. For significant changes, we will notify users through the app.
        </p>
      </section>

      <section>
        <h2 class="text-xl font-semibold mb-2">Contact</h2>
        <p>
          Questions about this policy? Reach out at <a href="mailto:stuffleberryco+privacy@gmail.com" class="text-indigo-600 hover:underline">stuffleberryco+privacy@gmail.com</a>.
        </p>
      </section>

      <section class="border-t pt-4 mt-8">
        <p class="text-sm text-gray-500 italic">
          "We believe in being honest, true, chaste, benevolent, virtuous, and in doing good to all men."
          — <a href="/gospel-library/eng/scriptures/pgp/a-of-f/1.md" class="text-indigo-500 hover:underline">Article of Faith 13</a>
        </p>
      </section>
    </div>

    <div class="mt-8 text-center">
      <router-link to="/login" class="text-indigo-600 hover:underline text-sm">← Back to Become</router-link>
    </div>
  </div>
</template>
