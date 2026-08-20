import type { MaybeRefOrGetter } from 'vue'

/**
 * Seo/social meta for public pages: document title (same logic as
 * usePageTitle) + open graph / twitter cards. og:image is the first photo
 * of the media list; og:url is the canonical url of the current route.
 */
export function useSeo(label?: MaybeRefOrGetter<string>) {
  const { data: settings } = useSettings()
  const { data: media } = useMedia()
  const { pick } = useI18n()
  const route = useRoute()
  const siteUrl = useRuntimeConfig().public.siteUrl.replace(/\/+$/, '')

  const title = useSiteTitle(label)

  const description = computed(() =>
    pick(settings.value?.translations?.site_subtitle, settings.value?.site_subtitle || ''),
  )

  const image = computed(() => {
    const photo = (media.value ?? []).find((m) => m.type === 'photo')
    return photo ? `${siteUrl}${photo.path}` : undefined
  })

  useSeoMeta({
    title,
    ogTitle: title,
    description,
    ogDescription: description,
    ogType: 'website',
    ogUrl: computed(() => `${siteUrl}${route.path}`),
    ogImage: image,
    twitterCard: 'summary_large_image',
  })
}
