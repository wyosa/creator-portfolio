<script setup lang="ts">
const { locale } = useI18n()
useHead({ htmlAttrs: { lang: computed(() => locale.value) } })
</script>

<template>
  <div>
    <SiteHeader />
    <main id="page">
      <slot />
    </main>
  </div>
</template>
