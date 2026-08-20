const apiProxy = process.env.NUXT_API_PROXY ?? 'http://localhost:8080'

export default defineNuxtConfig({
  ssr: true,
  devtools: { enabled: false },
  compatibilityDate: '2025-08-01',
  css: ['~/assets/css/main.css'],
  components: [{ path: '~/components/admin', pathPrefix: false }, '~/components'],
  routeRules: {
    '/api/**': { proxy: `${apiProxy}/api/**` },
    '/media/**': { proxy: `${apiProxy}/media/**` },
  },
  app: {
    head: {
      title: 'portfolio',
      meta: [
        {
          name: 'viewport',
          content: 'width=device-width, initial-scale=1, viewport-fit=cover',
        },
        {
          name: 'description',
          content: 'photographer & videographer portfolio',
        },
      ],
      script: [
        {
          // resolve theme before first paint: saved choice wins, else system
          innerHTML: `try{var t=localStorage.getItem('dp-theme');if(t!=='light'&&t!=='dark'){t=window.matchMedia('(prefers-color-scheme: dark)').matches?'dark':'light'}document.documentElement.dataset.theme=t}catch(e){}`,
        },
      ],
    },
  },
})
