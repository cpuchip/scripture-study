import { ref, watch } from 'vue'

const darkMode = ref(localStorage.getItem('reader-dark-mode') === 'true')

// Apply the class on module load so it's set before any component mounts
document.documentElement.classList.toggle('dark-mode', darkMode.value)

watch(darkMode, (val) => {
  localStorage.setItem('reader-dark-mode', String(val))
  document.documentElement.classList.toggle('dark-mode', val)
})

export function useTheme() {
  function toggleDarkMode() {
    darkMode.value = !darkMode.value
  }

  return { darkMode, toggleDarkMode }
}
