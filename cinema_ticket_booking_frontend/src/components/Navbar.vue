<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()

const handleLogout = () => {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <nav class="navbar">
    <div class="nav-brand">
      <router-link to="/">Cinema Booking</router-link>
    </div>
    <div class="nav-links">
      <router-link to="/">Movies</router-link>
      <router-link v-if="auth.isAuthenticated" to="/bookings">My Bookings</router-link>
      <router-link v-if="auth.user?.role === 'admin'" to="/admin/bookings">Admin</router-link>
    </div>
    <div class="nav-user">
      <template v-if="auth.isAuthenticated">
        <span class="username">{{ auth.user?.username }}</span>
        <button @click="handleLogout" class="btn-logout">Logout</button>
      </template>
      <template v-else>
        <router-link to="/login" class="btn-login">Login</router-link>
      </template>
    </div>
  </nav>
</template>

<style scoped>
.navbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 2rem;
  background: #2c3e50;
  color: white;
}

.nav-brand a {
  font-size: 1.5rem;
  font-weight: bold;
  color: white;
  text-decoration: none;
}

.nav-links {
  display: flex;
  gap: 1.5rem;
}

.nav-links a {
  color: white;
  text-decoration: none;
}

.nav-links a:hover {
  color: #42b883;
}

.nav-user {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.username {
  color: #42b883;
}

.btn-login,
.btn-logout {
  padding: 0.5rem 1rem;
  border-radius: 4px;
  text-decoration: none;
  cursor: pointer;
  border: none;
}

.btn-login {
  background: #42b883;
  color: white;
}

.btn-logout {
  background: #e74c3c;
  color: white;
}
</style>
