<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import type { Movie } from '@/types'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const movieId = route.params.id as string
const movie = ref<Movie | null>(null)
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    const movies = await api.getMovies()
    movie.value = movies.find((m) => m.id === movieId) || null
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load movie'
  } finally {
    loading.value = false
  }
})

const proceedToSeats = () => {
  if (!auth.isAuthenticated) {
    router.push({
      name: 'login',
      query: { redirect: `/seats/${movieId}` },
    })
    return
  }
  router.push(`/seats/${movieId}`)
}
</script>

<template>
  <div class="movie-page">
    <button class="back-btn" @click="router.push('/')">← Back to Movies</button>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="!movie" class="error">Movie not found</div>

    <div v-else class="movie-detail">
      <div class="movie-poster">
        <img v-if="movie.posterUrl" :src="movie.posterUrl" :alt="movie.title" />
        <div v-else class="poster-placeholder">🎬</div>
      </div>

      <div class="movie-info">
        <h1>{{ movie.title }}</h1>
        <p class="description">{{ movie.description || 'No description available' }}</p>
        <p class="duration">Duration: {{ movie.durationMinutes }} minutes</p>
        <p class="price">Price: ฿{{ movie.price }} per seat</p>

        <button class="btn-continue" @click="proceedToSeats">
          {{ auth.isAuthenticated ? 'Select Seats' : 'Login to Book' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.movie-page {
  padding: 2rem;
  max-width: 900px;
  margin: 0 auto;
}

.back-btn {
  background: none;
  border: none;
  color: #42b883;
  cursor: pointer;
  font-size: 1rem;
  margin-bottom: 1rem;
}

.loading,
.error {
  text-align: center;
  padding: 2rem;
  color: #666;
}

.movie-detail {
  display: grid;
  grid-template-columns: 300px 1fr;
  gap: 2rem;
}

.movie-poster {
  background: #eee;
  border-radius: 8px;
  overflow: hidden;
}

.movie-poster img {
  width: 100%;
  height: auto;
}

.poster-placeholder {
  height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 5rem;
}

.movie-info h1 {
  color: #2c3e50;
  margin: 0 0 1rem;
}

.description {
  color: #666;
  margin-bottom: 1rem;
}

.duration,
.price {
  color: #333;
  margin-bottom: 0.5rem;
}

.btn-continue {
  margin-top: 2rem;
  width: 100%;
  padding: 1rem;
  background: #42b883;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 1.1rem;
  cursor: pointer;
}

.btn-continue:hover {
  background: #3aa876;
}
</style>
