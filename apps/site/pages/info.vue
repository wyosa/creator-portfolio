<script setup lang="ts">
const { pick, locale, t } = useI18n()
usePageTitle(() => t('info'))

const { data: settings } = await useSettings()

/* only safe link schemes are rendered (http/https/mailto) */
const links = computed(() =>
  (settings.value?.info_links ?? []).filter(
    (l) => l.enabled && /^(https?:\/\/|mailto:)/i.test(l.url),
  ),
)

/* info text is markdown — rendered + sanitized on the client */
const infoHtml = ref('')

function renderInfo() {
  infoHtml.value = renderMarkdown(
    pick(settings.value?.translations?.info_text, settings.value?.info_text ?? ''),
  )
}

onMounted(renderInfo)
watch(locale, renderInfo)

function isExternal(url: string) {
  return /^https?:\/\//.test(url)
}
</script>

<template>
  <div class="info">
    <div class="info__text" v-html="infoHtml" />
    <p v-for="link in links" :key="link.url + link.label">
      <a
        class="info__link"
        :href="link.url"
        :target="isExternal(link.url) ? '_blank' : undefined"
        :rel="isExternal(link.url) ? 'noopener' : undefined"
      >
        <InstaIcon v-if="link.label.toLowerCase().includes('instagram')" />
        <svg
          v-else-if="link.label.toLowerCase().includes('telegram')"
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="currentColor"
          aria-hidden="true"
        >
          <path
            d="M21.9 4.6c.2-1-.6-1.7-1.5-1.3L2.7 10c-1 .4-.9 1.8.1 2.1l4.6 1.4 1.7 5.4c.3.9 1.4 1.1 1.9.3l2.5-3.4 4.4 3.2c.8.6 2 .2 2.2-.8l1.8-13.6zM8.7 12.8l9.7-6.1c.2-.1.4.1.2.3l-7.9 7.1-.3 3-1.7-4.3z"
          />
        </svg>
        <span>{{ link.label }}</span>
      </a>
    </p>
  </div>
</template>

<style scoped>
.info {
  padding: 2rem var(--page-padding) 6rem;
  max-width: 40rem;
  margin: 0 auto;
  font-size: 0.75rem;
  font-weight: 400;
  letter-spacing: 0.08em;
  text-transform: lowercase;
  line-height: 1.6;
  animation: infoIn 0.5s var(--transition-smooth);
}

.info p {
  margin-bottom: 1rem;
}

/* markdown content is injected via v-html — needs :deep */
.info__text :deep(p) {
  margin-bottom: 1rem;
}

.info__text :deep(a) {
  color: inherit;
  text-decoration: underline;
  text-underline-offset: 2px;
  transition: opacity 0.2s var(--transition-smooth);
}

.info__text :deep(a:hover) {
  opacity: 0.5;
}

.info__text :deep(ul),
.info__text :deep(ol) {
  padding-left: 1.2em;
  margin-bottom: 1rem;
  list-style: disc;
}

.info__text :deep(ol) {
  list-style: decimal;
}

.info a {
  color: inherit;
  text-decoration: none;
  border-bottom: 1px solid var(--line);
  transition: border-color 0.2s var(--transition-smooth);
}

.info a:hover {
  border-bottom-color: var(--color);
}

.info__link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

@keyframes infoIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
