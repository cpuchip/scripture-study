// Click-mode global toggle — tiny standalone composable so App.vue
// (eager-loaded) can import it without pulling all of useWordData
// (~248KB with the tier-words.json static import) into the entry bundle.
//
// `definition` (default): clicking a highlighted/linked word opens the
//   1828 + modern + Thummim definition card. The reader chases the
//   meaning of the word.
// `scripture`: clicking a word routes to /word-study/<word>, which shows
//   every verse containing the word (or its stem). The reader chases
//   where the word LIVES in the canon.
//
// Persisted to localStorage so the reader's preference survives reload.

import { ref, watch } from 'vue'

export type ClickMode = 'definition' | 'scripture'

const CLICK_MODE_KEY = 'click-mode-v1'

function loadInitialMode(): ClickMode {
  if (typeof localStorage === 'undefined') return 'definition'
  try {
    return localStorage.getItem(CLICK_MODE_KEY) === 'scripture' ? 'scripture' : 'definition'
  } catch {
    return 'definition'
  }
}

export const clickMode = ref<ClickMode>(loadInitialMode())

watch(clickMode, (m) => {
  try { localStorage.setItem(CLICK_MODE_KEY, m) } catch { /* storage off */ }
})
