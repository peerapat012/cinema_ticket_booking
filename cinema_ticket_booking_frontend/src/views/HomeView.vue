<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/services/api'
import type { Movie } from '@/types'

const router = useRouter()
const movies = ref<Movie[]>([])
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    movies.value = await api.getMovies()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load movies'
  } finally {
    loading.value = false
  }
})

const goToMovie = (id: string) => {
  router.push(`/movie/${id}`)
}
</script>

<template>
  <div class="home">
    <h1>Now Showing</h1>

    <div v-if="loading" class="loading">Loading movies...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="movies.length === 0" class="empty">No movies available</div>

    <div v-else class="movie-grid">
      <div v-for="movie in movies" :key="movie.id" class="movie-card" @click="goToMovie(movie.id)">
        <div class="poster">
          <img v-if="movie.posterUrl" :src="movie.posterUrl" :alt="movie.title" />
          <div v-else class="poster-placeholder">🎬</div>
        </div>
        <div class="movie-info">
          <h3>{{ movie.title }}</h3>
          <p class="duration">{{ movie.durationMinutes }} min</p>
          <p class="price">฿{{ movie.price }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.home {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

h1 {
  color: #2c3e50;
  margin-bottom: 2rem;
}

.loading,
.error,
.empty {
  text-align: center;
  padding: 2rem;
  color: #666;
}

.error {
  color: #e74c3c;
}

.movie-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 2rem;
}

.movie-card {
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: transform 0.2s;
}

.movie-card:hover {
  transform: translateY(-4px);
}

.poster {
  width: 100%;
  height: 280px;
  background: #eee;
  display: flex;
  align-items: center;
  justify-content: center;
}

.poster img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.poster-placeholder {
  font-size: 4rem;
}

.movie-info {
  padding: 1rem;
}

.movie-info h3 {
  margin: 0 0 0.5rem;
  color: #2c3e50;
}

.duration {
  color: #666;
  font-size: 0.9rem;
  margin: 0 0 0.5rem;
}

.price {
  color: #42b883;
  font-weight: bold;
  margin: 0;
}
</style>
