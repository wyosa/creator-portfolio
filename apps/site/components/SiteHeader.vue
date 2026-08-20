<script setup lang="ts">
const menuOpen = ref(false)
const route = useRoute()

const { data: settings } = await useSettings()
const { data: allMedia } = useMedia()
const { active, t, pick } = useI18n()

const siteName = computed(() =>
  pick(settings.value?.translations?.site_name, settings.value?.site_name ?? ''),
)
const siteSubtitle = computed(() =>
  pick(settings.value?.translations?.site_subtitle, settings.value?.site_subtitle ?? ''),
)

/* nav links appear only for content that exists */
const hasFeatured = computed(() => (allMedia.value ?? []).some((m) => m.featured))
const hasVideo = computed(() => (allMedia.value ?? []).some((m) => m.type === 'video'))
const hasPhoto = computed(() => (allMedia.value ?? []).some((m) => m.type === 'photo'))

const navLinks = computed(() => [
  { to: '/', labelKey: 'featured', show: hasFeatured.value },
  { to: '/film', labelKey: 'film', show: hasVideo.value },
  { to: '/photo', labelKey: 'photo', show: hasPhoto.value },
  { to: '/info', labelKey: 'info', show: true },
])

const isAdmin = computed(() => route.path.startsWith('/admin') && route.path !== '/admin/login')

async function logout() {
  try {
    await $fetch('/api/auth/logout', { method: 'POST' })
  } catch {
    // session may already be gone — still leave the page
  }
  await navigateTo('/admin/login')
}

watch(
  () => route.fullPath,
  () => {
    menuOpen.value = false
  },
)

useBodyScrollLock(menuOpen)

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') menuOpen.value = false
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <header class="site-header">
    <div class="site-header__title">
      <NuxtLink to="/" class="site-header__name">{{ siteName }}</NuxtLink>
      <p class="site-header__sub">{{ siteSubtitle }}</p>
    </div>

    <nav class="site-header__nav" aria-label="main">
      <template v-for="link in navLinks" :key="link.to">
        <NuxtLink v-if="link.show" :to="link.to">{{ t(link.labelKey) }}</NuxtLink>
      </template>
    </nav>

    <div class="site-header__right">
      <button v-if="isAdmin" type="button" class="site-header__logout" @click="logout">
        log out
      </button>
      <LangSwitcher v-if="active.length > 1" />
      <ThemeToggle />
      <button
        type="button"
        class="site-header__burger"
        :class="{ 'site-header__burger--open': menuOpen }"
        :aria-expanded="menuOpen"
        aria-label="toggle menu"
        @click="menuOpen = !menuOpen"
      >
        <span /><span />
      </button>
    </div>

    <div class="nav-overlay" :class="{ 'nav-overlay--open': menuOpen }" :inert="!menuOpen">
      <nav class="nav-overlay__nav" aria-label="mobile">
        <template v-for="link in navLinks" :key="link.to">
          <NuxtLink v-if="link.show" :to="link.to" @click="menuOpen = false">{{
            t(link.labelKey)
          }}</NuxtLink>
        </template>
      </nav>
      <div v-if="active.length > 1" class="nav-overlay__langs">
        <LangSwitcher dropup />
      </div>
    </div>
  </header>
</template>

<style scoped>
.site-header {
  position: relative;
  z-index: 100;
  background: var(--background);
  padding: 24px var(--page-padding);
  padding-top: max(24px, env(safe-area-inset-top));
  padding-left: max(var(--page-padding), env(safe-area-inset-left));
  padding-right: max(var(--page-padding), env(safe-area-inset-right));
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  transition:
    background-color 0.3s var(--transition-smooth),
    color 0.3s var(--transition-smooth);
}

.site-header__title {
  grid-column: 1;
  min-width: 0;
}

.site-header__name {
  display: inline-block;
  color: var(--color);
  text-decoration: none;
  font-size: 2.25rem;
  font-weight: 400;
  letter-spacing: -0.04em;
  line-height: 1.05;
  text-transform: lowercase;
}

.site-header__sub {
  margin: 0.15rem 0 0;
  font-size: 0.7rem;
  font-weight: 400;
  letter-spacing: 0.12em;
  text-transform: lowercase;
}

.site-header__nav {
  grid-column: 2;
  display: flex;
  gap: 1.5em;
  justify-content: center;
}

.site-header__nav a {
  position: relative;
  color: var(--color);
  text-decoration: none;
  font-size: 0.75rem;
  font-weight: 400;
  letter-spacing: 0.08em;
  text-transform: lowercase;
}

.site-header__nav a::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 0;
  width: 0;
  height: 1px;
  background: var(--color);
  transition: width 0.3s var(--transition-smooth);
}

.site-header__nav a:hover::after,
.site-header__nav a[aria-current='page']::after {
  width: 100%;
}

/* burger + mobile overlay */

.site-header__right {
  grid-column: 3;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.5rem;
}

.site-header__logout {
  font-size: 0.75rem;
  letter-spacing: 0.08em;
  text-transform: lowercase;
  text-decoration: underline;
  text-underline-offset: 3px;
  white-space: nowrap;
  transition: opacity 0.2s var(--transition-smooth);
}

.site-header__logout:hover {
  opacity: 0.5;
}

.nav-overlay__langs {
  position: absolute;
  bottom: 2.5rem;
}

.site-header__burger {
  display: none;
  position: relative;
  z-index: 95;
  width: 32px;
  height: 32px;
  flex-direction: column;
  justify-content: center;
  align-items: flex-end;
  gap: 5px;
  cursor: pointer;
}

.site-header__burger span {
  display: block;
  width: 22px;
  height: 2px;
  background: var(--color);
  transition:
    transform 0.3s var(--transition-smooth),
    opacity 0.2s var(--transition-smooth);
}

.site-header__burger--open span:first-child {
  transform: translateY(3.5px) rotate(45deg);
}

.site-header__burger--open span:last-child {
  transform: translateY(-3.5px) rotate(-45deg);
}

.nav-overlay {
  position: fixed;
  inset: 0;
  z-index: 90;
  background: var(--background);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  visibility: hidden;
  transition:
    opacity 0.3s var(--transition-smooth),
    visibility 0s linear 0.3s,
    background-color 0.3s var(--transition-smooth);
}

.nav-overlay--open {
  opacity: 1;
  visibility: visible;
  transition: opacity 0.3s var(--transition-smooth);
}

.nav-overlay__nav {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.75rem;
}

.nav-overlay__nav a {
  color: var(--color);
  text-decoration: none;
  font-size: 1.5rem;
  letter-spacing: -0.02em;
  text-transform: lowercase;
  opacity: 0;
  transform: translateY(10px);
  transition:
    opacity 0.4s var(--transition-smooth),
    transform 0.4s var(--transition-smooth);
}

.nav-overlay--open .nav-overlay__nav a {
  opacity: 1;
  transform: none;
}

.nav-overlay--open .nav-overlay__nav a:nth-child(1) {
  transition-delay: 0.08s;
}

.nav-overlay--open .nav-overlay__nav a:nth-child(2) {
  transition-delay: 0.14s;
}

.nav-overlay--open .nav-overlay__nav a:nth-child(3) {
  transition-delay: 0.2s;
}

.nav-overlay--open .nav-overlay__nav a:nth-child(4) {
  transition-delay: 0.26s;
}

.nav-overlay__nav a[aria-current='page'] {
  text-decoration: underline;
  text-underline-offset: 4px;
}

@media (max-width: 1023px) {
  .site-header__nav {
    display: none;
  }

  .site-header__burger {
    display: flex;
  }
}

@media (max-width: 767px) {
  .site-header {
    padding: 1rem 1.25rem;
  }

  .site-header__name {
    font-size: 1.3rem;
  }

  .site-header__sub {
    font-size: 0.6rem;
  }
}
</style>
