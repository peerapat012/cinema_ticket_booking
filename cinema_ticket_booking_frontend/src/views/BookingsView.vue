<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/services/api'
import type { Booking, Movie } from '@/types'

const bookings = ref<(Booking & { movie?: Movie })[]>([])
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    const [bookingResponse, movies] = await Promise.all([api.getMyBookings(), api.getMovies()])

    const bookingList = Array.isArray(bookingResponse) ? bookingResponse : []
    const movieMap = new Map(movies.map((m) => [m.id, m]))

    bookings.value = bookingList.map((b) => ({
      ...b,
      movie: movieMap.get(b.movieId),
    }))
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load bookings'
  } finally {
    loading.value = false
  }
})

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}
</script>

<template>
  <div class="bookings-page">
    <h1>My Bookings</h1>

    <div v-if="loading" class="loading">Loading bookings...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="bookings.length === 0" class="empty">
      <p>No bookings yet.</p>
      <router-link to="/" class="btn-book">Book a Movie</router-link>
    </div>

    <div v-else class="booking-list">
      <div v-for="booking in bookings" :key="booking.id" class="booking-card">
        <div class="booking-header">
          <span class="booking-code">#{{ booking.bookingCode }}</span>
          <span :class="['status', booking.status]">{{ booking.status }}</span>
        </div>

        <div class="booking-body">
          <h3>{{ booking.movie?.title || 'Unknown Movie' }}</h3>
          <p>Seats: {{ booking.seats.map((s) => s.seatNo).join(', ') }}</p>
          <p class="total">Total: ฿{{ booking.totalPrice }}</p>
          <p v-if="booking.status === 'pending'" class="deadline">
            Payment deadline: {{ formatDate(booking.paymentDeadline || '') }}
          </p>
        </div>

        <div class="booking-footer">
          <span class="booked-at">Booked: {{ formatDate(booking.bookedAt) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.bookings-page {
  padding: 2rem;
  max-width: 800px;
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

.empty p {
  margin-bottom: 1rem;
}

.btn-book {
  display: inline-block;
  padding: 0.75rem 1.5rem;
  background: #42b883;
  color: white;
  text-decoration: none;
  border-radius: 4px;
}

.booking-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.booking-card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.booking-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: #f8f9fa;
}

.booking-code {
  font-weight: bold;
  color: #2c3e50;
}

.status {
  padding: 0.25rem 0.75rem;
  border-radius: 4px;
  font-size: 0.85rem;
  text-transform: capitalize;
}

.status.confirmed {
  background: #d4edda;
  color: #155724;
}

.status.pending {
  background: #fff3cd;
  color: #856404;
}

.status.cancelled {
  background: #f8d7da;
  color: #721c24;
}

.status.expired {
  background: #f8d7da;
  color: #721c24;
}

.booking-body {
  padding: 1rem;
}

.booking-body h3 {
  margin: 0 0 0.5rem;
  color: #2c3e50;
}

.booking-body p {
  margin: 0.25rem 0;
  color: #666;
}

.booking-body .total {
  font-weight: bold;
  color: #42b883;
  font-size: 1.1rem;
  margin-top: 0.5rem;
}

.booking-body .deadline {
  color: #d32f2f;
  font-size: 0.9rem;
  margin-top: 0.5rem;
}

.booking-footer {
  padding: 0.75rem 1rem;
  border-top: 1px solid #eee;
}

.booked-at {
  color: #999;
  font-size: 0.85rem;
}
</style>
