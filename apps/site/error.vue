<script setup lang="ts">
import type { NuxtError } from '#app'

const props = defineProps<{ error: NuxtError }>()

const { t } = useI18n()
/* 404 has its own text; anything else is a generic error */
const msgKey = computed(() =>
  (props.error?.statusCode ?? 404) === 404 ? 'pageNotFound' : 'somethingWentWrong',
)
usePageTitle(() => t(msgKey.value))
</script>

<template>
  <div class="err">
    <p class="err__code">{{ error?.statusCode ?? 404 }}</p>
    <p class="err__msg">{{ t(msgKey) }}</p>
    <NuxtLink to="/" class="err__link">{{ t('backHome') }}</NuxtLink>
  </div>
</template>

<style scoped>
.err {
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  padding: var(--page-padding);
  text-align: center;
  text-transform: lowercase;
}

.err__code {
  font-size: 2.25rem;
  font-weight: 400;
  letter-spacing: -0.04em;
}

.err__msg {
  font-size: 0.75rem;
  letter-spacing: 0.12em;
  color: var(--muted);
}

.err__link {
  margin-top: 1rem;
  font-size: 0.75rem;
  letter-spacing: 0.08em;
  color: var(--color);
  text-decoration: underline;
  text-underline-offset: 3px;
  transition: opacity 0.2s var(--transition-smooth);
}

.err__link:hover {
  opacity: 0.5;
}
</style>
