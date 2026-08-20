<script setup lang="ts">
usePageTitle('login')

/* already logged in → straight to the admin */
const { error: meError } = await useFetch('/api/auth/me', {
  headers: import.meta.server ? useRequestHeaders(['cookie']) : undefined,
})
if (!meError.value) await navigateTo('/admin')

const username = ref('')
const password = ref('')
const errorMessage = ref('')
const pending = ref(false)

async function submit() {
  if (pending.value) return
  errorMessage.value = ''
  pending.value = true

  try {
    await $fetch('/api/auth/login', {
      method: 'POST',
      body: { username: username.value, password: password.value },
    })
  } catch (error) {
    const status = (error as { response?: { status?: number } })?.response?.status
    errorMessage.value = status === 401 ? 'wrong credentials' : 'something went wrong'
    pending.value = false
    return
  }

  pending.value = false
  await navigateTo('/admin')
}
</script>

<template>
  <div class="login">
    <form class="login__card" @submit.prevent="submit">
      <label class="field">
        <span>username</span>
        <input v-model="username" type="text" name="username" autocomplete="username" required />
      </label>

      <label class="field">
        <span>password</span>
        <input
          v-model="password"
          type="password"
          name="password"
          autocomplete="current-password"
          required
        />
      </label>

      <p v-if="errorMessage" class="tiny login__error">{{ errorMessage }}</p>

      <button class="btn" type="submit" :disabled="pending">
        {{ pending ? 'logging in...' : 'log in' }}
      </button>
    </form>
  </div>
</template>

<style scoped>
.login {
  min-height: 55vh;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: loginIn 0.4s var(--transition-smooth);
}

.login__card {
  width: min(320px, 100%);
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.login__error {
  margin: 0;
}

@keyframes loginIn {
  from {
    opacity: 0;
    transform: translateY(12px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
