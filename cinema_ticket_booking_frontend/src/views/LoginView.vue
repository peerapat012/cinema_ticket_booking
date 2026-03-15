<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/services/api'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const error = ref('')
const loading = ref(false)

const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID || ''

const handleCredentialResponse = async (response: { credential: string }) => {
  loading.value = true
  error.value = ''

  try {
    const result = await api.loginWithGoogle(response.credential)
    auth.setAuth(result.token, result.user)
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Login failed'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (window.google && clientId) {
    window.google.accounts.id.initialize({
      client_id: clientId,
      callback: handleCredentialResponse,
    })

    const container = document.getElementById('g_id_signin')
    if (container) {
      window.google.accounts.id.renderButton(container, {
        theme: 'outline',
        size: 'large',
        width: '100%',
      })
    }
  }
})
</script>

<template>
  <div class="login-container">
    <div class="login-card">
      <h1>Cinema Booking</h1>
      <p class="subtitle">Sign in to book your tickets</p>

      <div class="google-btn">
        <div id="g_id_signin"></div>
      </div>

      <div v-if="error" class="error">{{ error }}</div>
      <div v-if="loading" class="loading">Signing in...</div>
    </div>
  </div>
</template>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 64px);
  background: #f5f5f5;
}

.login-card {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  width: 100%;
  max-width: 400px;
  text-align: center;
}

h1 {
  margin: 0 0 0.5rem;
  color: #2c3e50;
}

.subtitle {
  color: #666;
  margin-bottom: 2rem;
}

.google-btn {
  display: flex;
  justify-content: center;
  margin: 1.5rem 0;
}

.error {
  color: #e74c3c;
  margin-top: 1rem;
}

.loading {
  color: #666;
  margin-top: 1rem;
}
</style>
