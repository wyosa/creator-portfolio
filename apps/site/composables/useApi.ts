/** api error normalized by useApi: http status + readable message */
export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

/**
 * $fetch wrapper for admin api calls.
 * - 401 → redirect to the login page (unless already there)
 * - 429 → localized "too many attempts" message (login is rate-limited)
 * - everything is rethrown as ApiError with a readable message
 */
export function useApi() {
  const { t } = useI18n()
  const route = useRoute()

  return async function api<T = unknown>(
    url: string,
    opts?: Parameters<typeof $fetch>[1],
  ): Promise<T> {
    try {
      return await $fetch<T>(url, opts)
    } catch (e) {
      const status = (e as { response?: { status?: number } })?.response?.status ?? 0
      if (status === 401 && route.path !== '/admin/login') {
        await navigateTo('/admin/login')
      }
      throw new ApiError(status, status === 429 ? t('tooManyAttempts') : 'something went wrong')
    }
  }
}
