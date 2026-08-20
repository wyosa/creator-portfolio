<script setup lang="ts">
type Theme = 'light' | 'dark'

const STORAGE_KEY = 'dp-theme'
const theme = ref<Theme>('light')

function apply(t: Theme) {
  theme.value = t
  document.documentElement.dataset.theme = t
}

function toggle() {
  const next: Theme = theme.value === 'dark' ? 'light' : 'dark'
  try {
    localStorage.setItem(STORAGE_KEY, next)
  } catch {
    // private mode etc — theme still applies for this session
  }
  apply(next)
}

let mq: MediaQueryList | null = null
const onSystem = (e: MediaQueryListEvent) => {
  // follow the system only while the user has not chosen explicitly
  let saved: string | null = null
  try {
    saved = localStorage.getItem(STORAGE_KEY)
  } catch {
    saved = null
  }
  if (!saved) apply(e.matches ? 'dark' : 'light')
}

onMounted(() => {
  const current = document.documentElement.dataset.theme
  if (current === 'dark' || current === 'light') theme.value = current
  mq = window.matchMedia('(prefers-color-scheme: dark)')
  mq.addEventListener('change', onSystem)
})

onBeforeUnmount(() => mq?.removeEventListener('change', onSystem))
</script>

<template>
  <button
    type="button"
    class="theme-toggle"
    :aria-label="theme === 'dark' ? 'switch to light theme' : 'switch to dark theme'"
    @click="toggle"
  >
    <svg
      v-if="theme === 'dark'"
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="1.5"
      aria-hidden="true"
    >
      <circle cx="12" cy="12" r="4.5" />
      <path
        d="M12 2v2.5M12 19.5V22M2 12h2.5M19.5 12H22M4.9 4.9l1.8 1.8M17.3 17.3l1.8 1.8M19.1 4.9l-1.8 1.8M6.7 17.3l-1.8 1.8"
      />
    </svg>
    <svg
      v-else
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="1.5"
      aria-hidden="true"
    >
      <path d="M20 13.5A8 8 0 0 1 10.5 4 8 8 0 1 0 20 13.5z" />
    </svg>
  </button>
</template>

<style scoped>
.theme-toggle {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color);
  opacity: 0.7;
  transition: opacity 0.2s var(--transition-smooth);
}

.theme-toggle:hover {
  opacity: 1;
}
</style>
