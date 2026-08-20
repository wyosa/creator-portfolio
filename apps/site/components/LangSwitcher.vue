<script setup lang="ts">
import { LOCALE_NAMES, type Locale } from '~/types/i18n'

const { locale, active, setLocale } = useI18n()

defineProps<{ dropup?: boolean }>()

const open = ref(false)
const root = ref<HTMLElement | null>(null)
const route = useRoute()

watch(
  () => route.fullPath,
  () => {
    open.value = false
  },
)

function choose(l: Locale) {
  setLocale(l)
  open.value = false
}

function onDocClick(e: MouseEvent) {
  if (root.value && !root.value.contains(e.target as Node)) open.value = false
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') open.value = false
}

onMounted(() => {
  document.addEventListener('click', onDocClick)
  window.addEventListener('keydown', onKeydown)
})
onBeforeUnmount(() => {
  document.removeEventListener('click', onDocClick)
  window.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <div ref="root" class="langsw">
    <button
      type="button"
      class="langsw__btn"
      :aria-expanded="open"
      aria-haspopup="listbox"
      aria-label="language"
      @click="open = !open"
    >
      <svg
        width="14"
        height="14"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
        aria-hidden="true"
      >
        <circle cx="12" cy="12" r="9" />
        <path
          d="M3 12h18M12 3c2.5 2.6 3.8 5.7 3.8 9S14.5 18.4 12 21c-2.5-2.6-3.8-5.7-3.8-9S9.5 5.6 12 3z"
        />
      </svg>
      <span>{{ LOCALE_NAMES[locale] }}</span>
    </button>

    <Transition name="langsw">
      <div
        v-if="open"
        class="langsw__popup"
        :class="{ 'langsw__popup--up': dropup }"
        role="listbox"
        aria-label="languages"
      >
        <button
          v-for="l in active"
          :key="l"
          type="button"
          role="option"
          :aria-selected="locale === l"
          class="langsw__opt"
          :class="{ 'langsw__opt--active': locale === l }"
          @click="choose(l)"
        >
          {{ LOCALE_NAMES[l] }}
        </button>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.langsw {
  position: relative;
}

.langsw__btn {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.75rem;
  letter-spacing: 0.08em;
  text-transform: lowercase;
  color: var(--color);
  opacity: 0.7;
  transition: opacity 0.2s var(--transition-smooth);
}

.langsw__btn:hover {
  opacity: 1;
}

.langsw__popup {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  z-index: 150;
  background: var(--background);
  border: 1px solid var(--line);
  padding: 0.4rem 0;
  display: flex;
  flex-direction: column;
}

.langsw__popup--up {
  top: auto;
  bottom: calc(100% + 8px);
}

.langsw__opt {
  padding: 0.45rem 1rem;
  text-align: left;
  font-size: 0.75rem;
  letter-spacing: 0.08em;
  text-transform: lowercase;
  color: var(--color);
  transition: opacity 0.15s var(--transition-smooth);
}

.langsw__opt:hover {
  opacity: 0.5;
}

.langsw__opt--active {
  text-decoration: underline;
  text-underline-offset: 3px;
}

.langsw-enter-active,
.langsw-leave-active {
  transition: opacity 0.2s var(--transition-smooth);
}

.langsw-enter-from,
.langsw-leave-to {
  opacity: 0;
}
</style>
