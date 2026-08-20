<script setup lang="ts">
definePageMeta({
  validate: (route) => ['photo', 'film'].includes(route.params.type as string),
})

const route = useRoute()
const { t } = useI18n()

const pageType = computed(() => (route.params.type === 'film' ? 'film' : 'photo'))
usePageTitle(() => t(pageType.value))

const { data: all, error } = await useMedia()
const items = computed(() =>
  (all.value ?? []).filter((m) => m.type === (pageType.value === 'film' ? 'video' : 'photo')),
)

/* api failure → friendly notice; a genuinely empty feed is a fatal 404 */
if (!error.value && !items.value.length) {
  throw createError({ statusCode: 404, statusMessage: 'page not found', fatal: true })
}

// client-side nav between /photo and /film reuses this page component
watch(pageType, () => {
  if (!error.value && !items.value.length) {
    showError({ statusCode: 404, statusMessage: 'page not found', fatal: true })
  }
})
</script>

<template>
  <p v-if="error" class="tiny feed-error">{{ t('unavailable') }}</p>
  <template v-else>
    <h1 class="page-title">{{ t(pageType) }}</h1>
    <MediaGrid :items="items" />
  </template>
</template>

<style scoped>
.feed-error {
  display: flex;
  justify-content: center;
  padding: 6rem 0;
  color: var(--muted);
}
</style>
