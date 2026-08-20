export default defineNuxtRouteMiddleware(async () => {
  const { error } = await useFetch('/api/auth/me', {
    headers: import.meta.server ? useRequestHeaders(['cookie']) : undefined,
  })

  if (error.value) {
    return navigateTo('/admin/login')
  }
})
