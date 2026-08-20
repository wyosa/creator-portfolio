<script setup lang="ts">
useSeo()

const { data: all, error } = await useMedia()
const { t } = useI18n()

/* home = featured when any exist, otherwise everything */
const featured = computed(() => (all.value ?? []).filter((m) => m.featured))
const items = computed(() => (featured.value.length ? featured.value : (all.value ?? [])))
const pageLabel = computed(() => (featured.value.length ? t('featured') : t('all')))
</script>

<template>
  <div>
    <p v-if="error" class="home-empty tiny">{{ t('unavailable') }}</p>
    <template v-else>
      <h1 v-if="items.length" class="page-title">{{ pageLabel }}</h1>
      <JustifiedGrid v-if="items.length" :items="items" />
      <p v-else class="home-empty tiny">{{ t('nothingHere') }}</p>
    </template>
  </div>
</template>

<style scoped>
.home-empty {
  display: flex;
  justify-content: center;
  padding: 6rem 0;
  color: var(--muted);
}
</style>
