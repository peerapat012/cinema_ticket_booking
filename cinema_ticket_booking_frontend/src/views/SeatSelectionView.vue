<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/services/api'
import type { Movie, Seat } from '@/types'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const movieId = route.params.movieId as string

const movie = ref<Movie | null>(null)
const selectedSeats = ref<string[]>([])
const seats = ref<Map<string, Seat>>(new Map())
const lockedSeats = ref<string[]>([])
const loading = ref(true)
const processing = ref(false)
const isNavigatingToPayment = ref(false)
const error = ref('')
const wsConnected = ref(false)
const ws = ref<WebSocket | null>(null)

const rows = ['A', 'B', 'C', 'D', 'E', 'F']
const cols = 10

onMounted(async () => {
  try {
    const [movies, bookedSeats] = await Promise.all([api.getMovies(), api.getBookedSeats(movieId)])

    movie.value = movies.find((m) => m.id === movieId) || null

    for (let r of rows) {
      for (let c = 1; c <= cols; c++) {
        const seatNo = `${r}${c}`
        const isBooked = bookedSeats.includes(seatNo)
        seats.value.set(seatNo, {
          seatNo,
          status: isBooked ? 'booked' : 'available',
        })
      }
    }

    connectWebSocket()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  if (isNavigatingToPayment.value) {
    if (ws.value) {
      ws.value.close()
    }
    return
  }

  if (lockedSeats.value.length > 0) {
    if (ws.value) {
      for (const seatNo of lockedSeats.value) {
        ws.value.send(
          JSON.stringify({ type: 'release_seat', movieId, seatNo, userId: auth.user?.id }),
        )
      }
    }

    api.unlockSeats(movieId, lockedSeats.value).catch((err) => {
      console.error('Failed to unlock seats on unmount:', err)
    })
  }
  if (ws.value) {
    ws.value.close()
  }
})

const connectWebSocket = () => {
  const wsUrl = `ws://localhost:8080/ws?movieId=${movieId}&userId=${auth.user?.id || ''}`
  ws.value = new WebSocket(wsUrl)

  ws.value.onopen = () => {
    wsConnected.value = true
  }

  ws.value.onmessage = (event) => {
    try {
      const data: Seat = JSON.parse(event.data)
      if (data.seatNo) {
        seats.value.set(data.seatNo, data)
      }
    } catch (e) {
      console.error('Failed to parse seat update:', e)
    }
  }

  ws.value.onclose = () => {
    wsConnected.value = false
  }

  ws.value.onerror = () => {
    error.value = 'WebSocket connection error'
  }
}

const toggleSeat = async (seatNo: string) => {
  const seat = seats.value.get(seatNo)
  if (!seat) return

  if (seat.status === 'booked') return

  if (selectedSeats.value.includes(seatNo)) {
    selectedSeats.value = selectedSeats.value.filter((s) => s !== seatNo)
    lockedSeats.value = lockedSeats.value.filter((s) => s !== seatNo)

    if (ws.value && wsConnected.value) {
      ws.value.send(
        JSON.stringify({
          type: 'release_seat',
          movieId,
          seatNo,
          userId: auth.user?.id || '',
        }),
      )
    }

    api.unlockSeats(movieId, [seatNo]).catch((err) => {
      console.error('Failed to unlock seat:', err)
    })
  } else {
    if (ws.value && wsConnected.value) {
      ws.value.send(
        JSON.stringify({
          type: 'lock_seat',
          movieId,
          seatNo,
          userId: auth.user?.id || '',
        }),
      )

      selectedSeats.value.push(seatNo)
      lockedSeats.value.push(seatNo)

      try {
        const result = await api.lockSeats(movieId, [seatNo])
        console.log('Seat locked:', seatNo, result)
      } catch (err) {
        console.error('Failed to lock seat:', err)
      }
    }
  }
}

const getSeatClass = (seatNo: string) => {
  const seat = seats.value.get(seatNo)
  if (selectedSeats.value.includes(seatNo)) return 'selected'
  if (seat?.status === 'booked') return 'booked'
  if (seat?.status === 'locked') return 'locked'
  return 'available'
}

const totalPrice = computed(() => {
  return selectedSeats.value.length * (movie.value?.price || 0)
})

const cancelSelection = async () => {
  if (ws.value && wsConnected.value) {
    for (const seatNo of lockedSeats.value) {
      ws.value.send(
        JSON.stringify({
          type: 'release_seat',
          movieId,
          seatNo,
          userId: auth.user?.id || '',
        }),
      )
    }
  }

  if (lockedSeats.value.length > 0) {
    try {
      await api.unlockSeats(movieId, lockedSeats.value)
    } catch (err) {
      console.error('Failed to unlock seats:', err)
    }
  }

  selectedSeats.value = []
  lockedSeats.value = []
  router.push(`/movie/${movieId}`)
}

const proceedToPayment = async () => {
  processing.value = true
  error.value = ''

  try {
    const bookingSeats = selectedSeats.value.map((seatNo) => ({
      seatNo,
      price: movie.value?.price || 0,
    }))

    const createdBooking = await api.createBooking({
      movieId,
      seats: bookingSeats,
    })

    const bookingId = createdBooking.id || (createdBooking as any)._id || ''

    isNavigatingToPayment.value = true

    router.push({
      path: '/payment',
      query: {
        bookingId,
        movieId,
        seats: selectedSeats.value.join(','),
        price: totalPrice.value.toString(),
      },
    })
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to create booking'
    processing.value = false
  }
}
</script>

<template>
  <div class="seat-page">
    <div v-if="movie" class="movie-header">
      <h2>{{ movie.title }}</h2>
    </div>

    <button class="back-btn" @click="cancelSelection">← Cancel</button>

    <div v-if="loading" class="loading">Loading seats...</div>
    <div v-else-if="error" class="error">{{ error }}</div>

    <div v-else class="seat-container">
      <div class="screen">Screen</div>

      <div class="seat-legend">
        <span class="legend-item"><span class="seat available"></span> Available</span>
        <span class="legend-item"><span class="seat selected"></span> Selected</span>
        <span class="legend-item"><span class="seat locked"></span> Locked</span>
        <span class="legend-item"><span class="seat booked"></span> Booked</span>
      </div>

      <div class="seats">
        <div v-for="row in rows" :key="row" class="row">
          <span class="row-label">{{ row }}</span>
          <button
            v-for="col in cols"
            :key="col"
            :class="['seat', getSeatClass(`${row}${col}`)]"
            :disabled="seats.get(`${row}${col}`)?.status === 'booked'"
            @click="toggleSeat(`${row}${col}`)"
          >
            {{ col }}
          </button>
        </div>
      </div>

      <div class="selection-info">
        <div v-if="selectedSeats.length > 0" class="info-section">
          <p>Selected Seats: {{ selectedSeats.join(', ') }}</p>
          <p>Total: ฿{{ totalPrice }}</p>
        </div>

        <div class="connection-status">
          <span :class="wsConnected ? 'connected' : 'disconnected'">
            {{ wsConnected ? '● Live' : '○ Connecting...' }}
          </span>
        </div>

        <button
          class="btn-pay"
          :disabled="selectedSeats.length === 0 || processing"
          @click="proceedToPayment"
        >
          {{ processing ? 'Creating Booking...' : 'Proceed to Payment' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.seat-page {
  padding: 2rem;
  max-width: 800px;
  margin: 0 auto;
}

.movie-header {
  text-align: center;
  margin-bottom: 1rem;
}

.movie-header h2 {
  color: #2c3e50;
  margin: 0;
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
}

.seat-container {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.screen {
  background: linear-gradient(to bottom, transparent, #ddd);
  height: 40px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #666;
  margin-bottom: 2rem;
}

.seat-legend {
  display: flex;
  justify-content: center;
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #666;
}

.seat {
  width: 30px;
  height: 30px;
  border-radius: 4px;
  border: none;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8rem;
}

.seat.available {
  background: #eee;
}

.seat.selected {
  background: #42b883;
  color: white;
}

.seat.locked {
  background: #f39c12;
  color: white;
}

.seat.booked {
  background: #e74c3c;
  color: white;
  cursor: not-allowed;
}

.seat:disabled {
  cursor: not-allowed;
}

.seats {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  align-items: center;
}

.row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.row-label {
  width: 24px;
  color: #666;
  font-weight: bold;
}

.selection-info {
  margin-top: 2rem;
  text-align: center;
}

.info-section p {
  margin: 0.5rem 0;
}

.connection-status {
  margin: 1rem 0;
}

.connection-status .connected {
  color: #42b883;
}

.connection-status .disconnected {
  color: #999;
}

.btn-pay {
  padding: 1rem 2rem;
  background: #42b883;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 1.1rem;
  cursor: pointer;
}

.btn-pay:hover:not(:disabled) {
  background: #3aa876;
}

.btn-pay:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
