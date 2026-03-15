<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '@/services/api'
import type { Booking, Movie, User } from '@/types'

const bookings = ref<Booking[]>([])
const movies = ref<Movie[]>([])
const users = ref<User[]>([])
const loading = ref(true)
const error = ref('')

const filters = ref({
  movieId: '',
  userId: '',
  status: '',
  date: '',
})

const filteredBookings = computed(() => {
  return bookings.value
})

const fetchData = async () => {
  loading.value = true
  error.value = ''
  try {
    const [allBookings, allMovies, allUsers] = await Promise.all([
      api.getAllBookings(),
      api.getMovies(),
      api.getAllUsers(),
    ])

    bookings.value = Array.isArray(allBookings) ? allBookings : []
    movies.value = Array.isArray(allMovies) ? allMovies : []
    users.value = Array.isArray(allUsers) ? allUsers : []
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load data'
    bookings.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
})

const applyFilters = async () => {
  loading.value = true
  try {
    const params = new URLSearchParams()
    if (filters.value.movieId) params.append('movieId', filters.value.movieId)
    if (filters.value.userId) params.append('userId', filters.value.userId)
    if (filters.value.status) params.append('status', filters.value.status)
    if (filters.value.date) params.append('date', filters.value.date)

    const query = params.toString() ? `?${params.toString()}` : ''
    const allBookings = await api.getAllBookings(query)
    bookings.value = allBookings
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to filter bookings'
  } finally {
    loading.value = false
  }
}

const clearFilters = () => {
  filters.value = {
    movieId: '',
    userId: '',
    status: '',
    date: '',
  }
  fetchData()
}

const getMovieTitle = (movieId: string) => {
  const movie = movies.value.find((m) => m.id === movieId)
  return movie?.title || 'Unknown Movie'
}

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<template>
  <div class="admin-page">
    <h1>Admin Dashboard - Bookings</h1>

    <div class="filters">
      <div class="filter-group">
        <label>Movie</label>
        <select v-model="filters.movieId">
          <option value="">All Movies</option>
          <option v-for="movie in movies" :key="movie.id" :value="movie.id">
            {{ movie.title }}
          </option>
        </select>
      </div>

      <div class="filter-group">
        <label>User</label>
        <select v-model="filters.userId">
          <option value="">All Users</option>
          <option v-for="user in users" :key="user.id" :value="user.id">
            {{ user.username }} ({{ user.email }})
          </option>
        </select>
      </div>

      <div class="filter-group">
        <label>Status</label>
        <select v-model="filters.status">
          <option value="">All Status</option>
          <option value="pending">Pending</option>
          <option value="confirmed">Confirmed</option>
          <option value="cancelled">Cancelled</option>
          <option value="expired">Expired</option>
        </select>
      </div>

      <div class="filter-group">
        <label>Date</label>
        <input type="date" v-model="filters.date" />
      </div>

      <div class="filter-actions">
        <button @click="applyFilters" class="btn-filter">Filter</button>
        <button @click="clearFilters" class="btn-clear">Clear</button>
      </div>
    </div>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error">{{ error }}</div>

    <div v-else class="table-container">
      <table>
        <thead>
          <tr>
            <th>Booking Code</th>
            <th>Movie</th>
            <th>User ID</th>
            <th>Seats</th>
            <th>Total</th>
            <th>Status</th>
            <th>Payment Status</th>
            <th>Created At</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="booking in bookings" :key="booking.id">
            <td>{{ booking.bookingCode }}</td>
            <td>{{ getMovieTitle(booking.movieId) }}</td>
            <td>{{ booking.userId }}</td>
            <td>{{ booking.seats.map((s) => s.seatNo).join(', ') }}</td>
            <td>฿{{ booking.totalPrice }}</td>
            <td>
              <span :class="['status', booking.status]">{{ booking.status }}</span>
            </td>
            <td>
              <span :class="['payment-status', booking.paymentStatus]">{{
                booking.paymentStatus
              }}</span>
            </td>
            <td>{{ formatDate(booking.createdAt) }}</td>
          </tr>
        </tbody>
      </table>

      <div v-if="bookings.length === 0" class="empty">No bookings found</div>
    </div>
  </div>
</template>

<style scoped>
.admin-page {
  padding: 2rem;
  max-width: 1400px;
  margin: 0 auto;
}

h1 {
  color: #2c3e50;
  margin-bottom: 2rem;
}

.filters {
  display: flex;
  gap: 1rem;
  margin-bottom: 2rem;
  padding: 1rem;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.filter-group label {
  font-size: 0.85rem;
  color: #666;
}

.filter-group select,
.filter-group input {
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  min-width: 150px;
}

.filter-actions {
  display: flex;
  align-items: flex-end;
  gap: 0.5rem;
}

.btn-filter {
  padding: 0.5rem 1rem;
  background: #42b883;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.btn-clear {
  padding: 0.5rem 1rem;
  background: #95a5a6;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.loading,
.error {
  text-align: center;
  padding: 2rem;
}

.error {
  color: #e74c3c;
}

.table-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  padding: 1rem;
  text-align: left;
  border-bottom: 1px solid #eee;
}

th {
  background: #f8f9fa;
  font-weight: 600;
  color: #2c3e50;
}

.status,
.payment-status {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.85rem;
  text-transform: capitalize;
}

.status.pending {
  background: #fff3cd;
  color: #856404;
}

.status.confirmed {
  background: #d4edda;
  color: #155724;
}

.status.cancelled {
  background: #f8d7da;
  color: #721c24;
}

.status.expired {
  background: #f8d7da;
  color: #721c24;
}

.payment-status.pending {
  background: #fff3cd;
  color: #856404;
}

.payment-status.completed {
  background: #d4edda;
  color: #155724;
}

.payment-status.failed {
  background: #f8d7da;
  color: #721c24;
}

.empty {
  text-align: center;
  padding: 2rem;
  color: #666;
}
</style>
